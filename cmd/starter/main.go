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
	"github.com/robfig/cron"
	"github.com/xuyu/goredis"
)

var dockerClient *docker.Client
var log = logrus.New()
var s3Bucket *s3.Bucket
var redisClient *goredis.Redis

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

	redisClient, err = goredis.DialURL(os.Getenv("redis_url"))
	if err != nil {
		log.WithFields(logrus.Fields{
			"url": os.Getenv("redis_url"),
		}).Panic("Unable to connect to Redis")
		os.Exit(1)
	}
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

	c := cron.New()
	c.AddFunc("0 */5 * * * *", announceActiveWorker)
	c.Start()

	for {
		fetchBuildJob()
		time.Sleep(10 * time.Second)
	}
}

func announceActiveWorker() {
	timestamp := float64(time.Now().Unix())
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to determine hostname")
	}

	redisClient.ZAdd("active-workers", map[string]float64{
		hostname: timestamp,
	})

	// Remove old clients to ensure the redis doesn't get filled with old data
	redisClient.ZRemRangeByScore("active-workers", "-inf", strconv.Itoa(int(time.Now().Unix()-3600)))
}

func fetchBuildJob() {
	queueLength, err := redisClient.LLen("build-queue")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unable to determine queue length.")
	}

	if queueLength < 1 {
		time.Sleep(5 * time.Second)
		return
	}
	body, err := redisClient.LPop("build-queue")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("An error occurred while getting job")
	}

	job, err := buildjob.FromBytes(body)
	if err != nil {
		// there was a job we could not parse throw it away and continue
		return
	}
	repo := job.Repository

	tmpDir, err := ioutil.TempDir("", "gobuild")
	orFail(err)
	buildStartTime := time.Now()
	buildOK, triggerUpload := build(repo, tmpDir)

	if triggerUpload {
		orFail(builddbCreator.GenerateBuildDB(tmpDir))
		uploadAssets(repo, tmpDir)
	}

	config, err := buildconfig.LoadFromFile(fmt.Sprintf("%s/.gobuilder.yml", tmpDir))
	if err != nil {
		// We got no .gobuilder.yml? Assume something was terribly wrong and requeue build.
		buildOK = false
	}

	if buildOK {
		orFail(redisClient.Set(fmt.Sprintf("project::%s::build-status", repo), "finished", 0, 0, false, false))
		redisClient.Set(fmt.Sprintf("project::%s::build-duration", repo), fmt.Sprintf("%d", int(time.Now().Sub(buildStartTime).Seconds())), 0, 0, false, false)
		redisClient.ZAdd("last-builds", map[string]float64{
			repo: float64(time.Now().Unix()),
		})
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
		orFail(redisClient.Set(fmt.Sprintf("project::%s::build-status", repo), "queued", 0, 0, false, false))
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

			redisClient.RPush("build-queue", string(queueEntry))
			orFail(err)
		} else {
			log.WithFields(logrus.Fields{
				"repository":         repo,
				"numberOfBuildTries": maxJobRetries,
			}).Error("Finally failed build")
			orFail(redisClient.Set(fmt.Sprintf("project::%s::build-status", repo), "failed", 0, 0, false, false))
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

	orFail(redisClient.Set(fmt.Sprintf("project::%s::build-status", repo), "building", 0, 0, false, false))
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
