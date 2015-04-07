package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/buildconfig"
	"github.com/Luzifer/gobuilder/builddbCreator"
	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/Luzifer/gobuilder/notifier"
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/papertrail"
	"github.com/fsouza/go-dockerclient"
	"github.com/kr/beanstalk"
)

var dockerClient *docker.Client
var log = logrus.New()
var s3Bucket *s3.Bucket

const maxJobRetries int = 5

func orFail(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	log.Out = os.Stderr

	papertrail_port, err := strconv.Atoi(os.Getenv("papertrail_port"))
	if err != nil {
		log.Info("Failed to read papertrail_port, using only STDERR")
		return
	}
	hook, err := logrus_papertrail.NewPapertrailHook(os.Getenv("papertrail_host"), papertrail_port, "GoBuilder Starter")
	if err != nil {
		log.Panic("Unable to create papertrail connection")
		os.Exit(1)
	}

	log.Hooks.Add(hook)
}

func main() {
	awsAuth, err := aws.EnvAuth()
	orFail(err)
	s3Bucket = s3.New(awsAuth, aws.EUWest).Bucket("gobuild.luzifer.io")

	dockerClientTmp, err := docker.NewClient("unix:///var/run/docker.sock")
	orFail(err)
	dockerClient = dockerClientTmp

	err = dockerClient.PullImage(docker.PullImageOptions{
		Repository: os.Getenv("BUILD_IMAGE"),
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	orFail(err)

	conn, err := beanstalk.Dial("tcp", os.Getenv("BEANSTALK_ADDR"))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%v", err),
		}).Error("Beanstalk-Connect")
		panic(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	waitForBuild(conn)
}

func waitForBuild(conn *beanstalk.Conn) {
	ts := beanstalk.NewTubeSet(conn, "gobuild.luzifer.io")
	for {
		id, body, err := ts.Reserve(10 * time.Second)
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			continue
		} else if err != nil {
			log.WithFields(logrus.Fields{
				"error": fmt.Sprintf("%v", err),
			}).Error("Tube-Reserve")
			panic(err)
		}

		job, err := buildjob.FromBytes(body)
		if err != nil {
			// there was a job we could not parse throw it away and continue
			_ = conn.Delete(id)
			continue
		}
		repo := job.Repository

		tmpDir, err := ioutil.TempDir("", "gobuild")
		orFail(err)
		buildOK, triggerUpload := build(repo, tmpDir)
		_ = conn.Delete(id)

		if triggerUpload {
			orFail(builddbCreator.GenerateBuildDB(tmpDir))
			uploadAssets(repo, tmpDir)
		}

		config, err := buildconfig.LoadFromFile(fmt.Sprintf("%s/.gobuilder.yml", tmpDir))
		orFail(err)

		if buildOK {
			orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("finished"), "text/plain", s3.PublicRead))
			_ = os.RemoveAll(tmpDir)

			log.WithFields(logrus.Fields{
				"repository": repo,
			}).Info("Finished build")

			orFail(config.Notify.Execute(notifier.NotifyMetaData{
				EventType:  "success",
				Repository: repo,
			}))

			triggerSubBuilds(repo, config.Triggers)
		} else {
			orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("queued"), "text/plain", s3.PublicRead))
			log.WithFields(logrus.Fields{
				"repository": repo,
			}).Error("Failed build")

			orFail(config.Notify.Execute(notifier.NotifyMetaData{
				EventType:  "error",
				Repository: repo,
			}))

			// Try to build the job only 5 times not to clutter the queue
			if job.NumberOfExecutions < maxJobRetries-1 {
				job.NumberOfExecutions = job.NumberOfExecutions + 1
				queueEntry, err := job.ToByte()
				orFail(err)

				_, err = conn.Put([]byte(queueEntry), 1, 0, 900*time.Second)
				orFail(err)
			} else {
				log.WithFields(logrus.Fields{
					"repository":         repo,
					"numberOfBuildTries": maxJobRetries,
				}).Error("Finally failed build")
				orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("failed"), "text/plain", s3.PublicRead))
			}
		}

	}
}

func build(repo, tmpDir string) (bool, bool) {
	log.WithFields(logrus.Fields{
		"repository": repo,
	}).Info("Beginning to process repo")

	cfg := &docker.Config{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Image:        os.Getenv("BUILD_IMAGE"),
		Env: []string{
			fmt.Sprintf("REPO=%s", repo),
		},
	}
	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: cfg,
	})
	orFail(err)
	err = dockerClient.StartContainer(container.ID, &docker.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:/artifacts", tmpDir)},
		Privileged:   false,
		PortBindings: make(map[docker.Port][]docker.PortBinding),
	})
	orFail(err)

	orFail(s3Bucket.Put(fmt.Sprintf("%s/build.status", repo), []byte("building"), "text/plain", s3.PublicRead))
	status, err := dockerClient.WaitContainer(container.ID)
	orFail(err)

	logFile, err := os.Create(fmt.Sprintf("%s/build.log", tmpDir))
	orFail(err)
	defer func() {
		_ = logFile.Close()
	}()

	err = dockerClient.Logs(docker.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		OutputStream: logFile,
	})
	orFail(err)

	if status == 0 {
		return true, true
	} else if status == 130 {
		// Special case: Build was aborted due to redundant build request
		return true, false
	}
	return false, false
}

func uploadAssets(repo, tmpDir string) {
	assets, err := ioutil.ReadDir(tmpDir)
	orFail(err)
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
		originalPath := fmt.Sprintf("%s/%s", tmpDir, f.Name())
		path := fmt.Sprintf("%s/%s", repo, f.Name())
		fileContent, err := ioutil.ReadFile(originalPath)
		orFail(err)

		err = s3Bucket.Put(path, fileContent, "", s3.PublicRead)
		orFail(err)
	}
}

func triggerSubBuilds(sourcerepo string, repos []string) {
	if len(repos) > 20 {
		// Flood / DDoS protection
		log.WithFields(logrus.Fields{
			"repository":         sourcerepo,
			"number_of_triggers": len(repos),
		}).Error("Too many triggers passed")
		return
	}

	for _, repo := range repos {
		go func(repo string) {
			resp, err := http.PostForm("https://gobuilder.me/build", url.Values{
				"repository": []string{repo},
			})
			if err != nil {
				log.WithFields(logrus.Fields{
					"repository": sourcerepo,
					"subrepo":    repo,
					"error":      fmt.Sprintf("%v", err),
				}).Error("Could not queue SubBuild")
			} else {
				defer resp.Body.Close()
			}
		}(repo)
	}
}
