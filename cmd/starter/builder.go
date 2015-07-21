package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/buildconfig"
	"github.com/Luzifer/gobuilder/builddbCreator"
	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/Luzifer/gobuilder/notifier"
	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/satori/go.uuid"
)

const (
	BuildStatusQueued   = "queued"
	BuildStatusStarted  = "building"
	BuildStatusFinished = "finished"
	BuildStatusFailed   = "failed"
)

type builder struct {
	// Representation of job currently built by this builder
	job *buildjob.BuildJob

	// Some remembered things to use in different calls
	tmpDir         string
	buildStartTime time.Time
	container      *docker.Container
	buildConfig    *buildconfig.BuildConfig

	// Details about the status of the build
	BuildOK        bool
	UploadRequired bool
	BuildLog       string
	AbortReason    string
}

func newBuilder(job *buildjob.BuildJob) *builder {
	return &builder{
		job: job,
	}
}

func (b *builder) AquireLock() error {
	// Aquire lock to ensure one repo is not built twice
	lockID := uuid.NewV4().String()
	redisClient.Set(fmt.Sprintf("project::%s::build-lock", b.job.Repository), lockID, 1800, 0, false, true)

	lock, err := redisClient.Get(fmt.Sprintf("project::%s::build-lock", b.job.Repository))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("AquireLock: Unable to fetch lock state")
		return err
	}

	if string(lock) != lockID {
		return fmt.Errorf("Aquire lock failed. Locked by someone else.")
	}

	return nil
}

func (b *builder) PutBackJob(increaseFails bool) {
	if increaseFails {
		b.job.NumberOfExecutions++
	}

	if b.job.NumberOfExecutions > maxJobRetries {
		log.WithFields(logrus.Fields{
			"repository":         b.job.Repository,
			"numberOfBuildTries": maxJobRetries,
		}).Error("Finally failed build")
		b.UpdateBuildStatus(BuildStatusFailed, 0)
		return
	}

	queueEntry, err := b.job.ToByte()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Could not put job back to queue")
	}

	_, err = redisClient.RPush("build-queue", string(queueEntry))
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Could not put job back to queue")
	}

	b.UpdateBuildStatus(BuildStatusQueued, 0)
}

func (b *builder) PrepareBuild() error {
	tmpDir, err := ioutil.TempDir("", "gobuild")
	if err != nil {
		return err
	}
	b.tmpDir = tmpDir
	b.buildStartTime = time.Now()

	b.UpdateBuildStatus(BuildStatusStarted, 1800)
	return nil
}

func (b *builder) Build() error {
	log.WithFields(logrus.Fields{
		"repository": b.job.Repository,
	}).Info("Beginning to process repo")

	cfg := &docker.Config{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Image:        conf.BuildImage.ImageName,
		Env: []string{
			fmt.Sprintf("REPO=%s", b.job.Repository),
			fmt.Sprintf("GPG_DECRYPT_KEY=%s", conf.BuildImage.GPGDecryptKey),
		},
	}
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: cfg,
	})
	if err != nil {
		return err
	}

	b.container = container

	err = dockerClient.StartContainer(container.ID, &docker.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:/artifacts", b.tmpDir)},
		Privileged:   false,
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	})
	if err != nil {
		return err
	}

	if err := redisClient.Set(fmt.Sprintf("project::%s::build-status", b.job.Repository), "building", 0, 0, false, false); err != nil {
		return err
	}

	status, err := dockerClient.WaitContainer(container.ID)
	if err != nil {
		return err
	}

	switch status {
	case 0:
		b.BuildOK = true
		b.UploadRequired = true
	case 130: // Special case: Build was aborted due to redundant build request
		b.BuildOK = true
		b.UploadRequired = false
	default:
		b.BuildOK = false
		b.UploadRequired = false
	}

	b.buildConfig, err = buildconfig.LoadFromFile(fmt.Sprintf("%s/.gobuilder.yml", b.tmpDir))
	if err != nil {
		// We got no .gobuilder.yml? Assume something was terribly wrong and requeue build.
		b.BuildOK = false
	}

	return nil
}

func (b *builder) FetchBuildLog() error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := dockerClient.Logs(docker.LogsOptions{
		Container:    b.container.ID,
		Stdout:       true,
		OutputStream: w,
	})
	if err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}
	b.BuildLog = buf.String()

	return nil
}

func (b *builder) WriteBuildLog() error {
	buildID := fmt.Sprintf("%x", sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 10))))[0:16]
	if err := redisClient.Set(fmt.Sprintf("project::%s::logs::%s", b.job.Repository, buildID), b.BuildLog, 0, 0, false, false); err != nil {
		return err
	}

	logMeta := buildjob.BuildLog{
		Success: b.BuildOK,
		Time:    time.Now(),
		ID:      buildID,
	}
	meta, err := logMeta.ToString()
	if err != nil {
		return err
	}

	if _, err := redisClient.ZAdd(fmt.Sprintf("project::%s::logs", b.job.Repository), map[string]float64{
		meta: float64(time.Now().Unix()),
	}); err != nil {
		return err
	}

	return nil
}

func (b *builder) UploadAssets() error {
	assets, err := ioutil.ReadDir(b.tmpDir)
	if err != nil {
		return err
	}

	for _, f := range assets {
		if f.IsDir() {
			// Some repos are creating directories. Don't know why. Ignore them.
			continue
		}

		if strings.HasPrefix(f.Name(), ".") {
			// Dotfiles are used to transport metadata from the container
			continue
		}

		log.Debugf("Uploading asset %s...", f.Name())
		originalPath := fmt.Sprintf("%s/%s", b.tmpDir, f.Name())
		path := fmt.Sprintf("%s/%s", b.job.Repository, f.Name())
		fileContent, err := ioutil.ReadFile(originalPath)
		if err != nil {
			return err
		}

		err = s3Bucket.Put(path, fileContent, "", s3.PublicRead)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *builder) UpdateMetaData() error {
	// Only write build-duration if this was a build with assets
	redisClient.Set(fmt.Sprintf("project::%s::build-duration", b.job.Repository), fmt.Sprintf("%d", int(time.Now().Sub(b.buildStartTime).Seconds())), 0, 0, false, false)
	redisClient.ZAdd("last-builds", map[string]float64{
		b.job.Repository: float64(time.Now().Unix()),
	})

	// Handle signature output
	builtTagsRaw, err := ioutil.ReadFile(fmt.Sprintf("%s/.built_tags", b.tmpDir))
	if err != nil {
		return err
	}
	buildTags := strings.Split(string(builtTagsRaw), "\n")
	for _, tag := range buildTags {
		signature, err := ioutil.ReadFile(fmt.Sprintf("%s/.signature_%s", b.tmpDir, tag))
		if err != nil {
			redisClient.Del(fmt.Sprintf("project::%s::signatures::%s", b.job.Repository, tag))
		} else {
			redisClient.Set(fmt.Sprintf("project::%s::signatures::%s", b.job.Repository, tag), string(signature), 0, 0, false, false)
		}

		hashes, err := ioutil.ReadFile(fmt.Sprintf("%s/.hashes_%s.txt", b.tmpDir, tag))
		if err == nil {
			redisClient.Set(fmt.Sprintf("project::%s::hashes::%s", b.job.Repository, tag), string(hashes), 0, 0, false, false)
		} else {
			redisClient.Del(fmt.Sprintf("project::%s::hashes::%s", b.job.Repository, tag))
		}
	}

	// Log last build
	gitHash, err := ioutil.ReadFile(fmt.Sprintf("%s/.build_master", b.tmpDir))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to read gitHash")
		gitHash = []byte("000000")
	}
	if err := redisClient.Set(fmt.Sprintf("project::%s::last-build", b.job.Repository), string(gitHash), 0, 0, false, false); err != nil {
		log.WithFields(logrus.Fields{
			"error":      err,
			"repository": b.job.Repository,
		}).Error("Unable to write last-build")
	}

	// Upload build.db
	builddbCreator.GenerateBuildDB(b.tmpDir)
	buildDB, err := ioutil.ReadFile(fmt.Sprintf("%s/.build.db", b.tmpDir))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to read build.db")
		return err
	}
	if err := redisClient.Set(fmt.Sprintf("project::%s::builddb", b.job.Repository), string(buildDB), 0, 0, false, false); err != nil {
		log.WithFields(logrus.Fields{
			"error":      err,
			"repository": b.job.Repository,
		}).Error("Unable to write builddb")

		return err
	}

	return nil
}

func (b *builder) SendNotifications() {
	eventType := "success"
	if !b.BuildOK {
		eventType = "error"
	}

	encryptionKey, err := redisClient.Get(fmt.Sprintf("project::%s::encryption-key", b.job.Repository))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"repo":  b.job.Repository,
		}).Error("Unable to load encryption key")
	}

	if err := b.buildConfig.Notify.Execute(notifier.NotifyMetaData{
		EventType:  eventType,
		Repository: b.job.Repository,
	}, conf, string(encryptionKey)); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to send notification")
	}
}

func (b *builder) TriggerSubBuilds() {
	if len(b.buildConfig.Triggers) > 20 {
		// Flood / DDoS protection
		log.WithFields(logrus.Fields{
			"repository":         b.job.Repository,
			"number_of_triggers": len(b.buildConfig.Triggers),
		}).Error("Too many triggers passed")
		return
	}

	for _, repo := range b.buildConfig.Triggers {
		go func(repo string) {
			resp, err := http.PostForm("https://gobuilder.me/api/v1/build", url.Values{
				"repository": []string{repo},
			})
			if err != nil {
				log.WithFields(logrus.Fields{
					"repository": b.job.Repository,
					"subrepo":    repo,
					"error":      fmt.Sprintf("%v", err),
				}).Error("Could not queue SubBuild")
			} else {
				defer resp.Body.Close()
			}
		}(repo)
	}
}

func (b *builder) Cleanup() {
	redisClient.Del(fmt.Sprintf("project::%s::build-lock", b.job.Repository))
	_ = os.RemoveAll(b.tmpDir)

	log.WithFields(logrus.Fields{
		"repository": b.job.Repository,
	}).Info("Finished build")
}

func (b *builder) UpdateBuildStatus(status string, expire int) {
	if err := redisClient.Set(fmt.Sprintf("project::%s::build-status", b.job.Repository), status, expire, 0, false, false); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to set the build status to 'queued'")
	}
}

func (b *builder) IsBuildable() (bool, error) {
	if strings.Contains(b.BuildLog, "could not read Username for 'https://github.com': No such device or address") {
		// Someone tried to build a private GitHub repository
		b.AbortReason = "GoBuilder is unable to build private repositories"
	}

	if strings.Contains(b.BuildLog, "no buildable Go source files in") {
		// We got a directory without any .go files
		b.AbortReason = "In the directory used to build are no *.go files present"
	}

	if b.AbortReason != "" {
		if err := redisClient.Set(fmt.Sprintf("project::%s::abort", b.job.Repository), b.AbortReason, 0, 0, false, false); err != nil {
			return false, err
		}
	}

	return b.AbortReason == "", nil
}
