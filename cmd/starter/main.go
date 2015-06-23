package main

import (
	"os"
	"strconv"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/papertrail"
	"github.com/cenkalti/backoff"
	"github.com/fsouza/go-dockerclient"
	"github.com/robfig/cron"
	"github.com/xuyu/goredis"
)

var dockerClient *docker.Client
var log = logrus.New()
var s3Bucket *s3.Bucket
var redisClient *goredis.Redis
var currentJobs chan bool

const (
	maxJobRetries       = 5
	maxConcurrentBuilds = 2
)

func init() {
	log.Out = os.Stderr

	// Add Papertrail connection for logging
	papertrailPort, err := strconv.Atoi(os.Getenv("papertrail_port"))
	if err == nil {
		hook, err := logrus_papertrail.NewPapertrailHook(os.Getenv("papertrail_host"), papertrailPort, "GoBuilder Starter")
		if err != nil {
			log.Panic("Unable to create papertrail connection")
			os.Exit(1)
		}

		log.Hooks.Add(hook)
	} else {
		log.Info("Failed to read papertrail_port, using only STDERR")
	}

	redisClient, err = goredis.DialURL(os.Getenv("redis_url"))
	if err != nil {
		log.WithFields(logrus.Fields{
			"url": os.Getenv("redis_url"),
		}).Panic("Unable to connect to Redis")
		os.Exit(1)
	}

	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Unable to read AWS credentials")
		os.Exit(1)
	}
	s3Bucket = s3.New(awsAuth, aws.EUWest).Bucket("gobuild.luzifer.io")

	dockerClient, err = docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Unable to connect to docker daemon")
		os.Exit(1)
	}

	currentJobs = make(chan bool, maxConcurrentBuilds)
}

func main() {
	if err := backoff.Retry(pullLatestImage, backoff.NewExponentialBackOff()); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Unable to fetch docker image for builder")
		os.Exit(1)
	}

	c := cron.New()
	c.AddFunc("0 */5 * * * *", announceActiveWorker)
	c.AddFunc("0 */30 * * * *", func() {
		err := pullLatestImage()
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to refresh build image")
		}
	})
	c.AddFunc("*/10 * * * * *", func() {
		go doBuildProcess()
	})
	c.Start()

	for {
		select {}
	}
}

func doBuildProcess() {
	if len(currentJobs) == maxConcurrentBuilds {
		// If maxConcurrentBuilds are running, do not fetch more jobs
		return
	}

	currentJobs <- true
	defer func() {
		<-currentJobs
	}()

	queueLength, err := redisClient.LLen("build-queue")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unable to determine queue length.")
		return
	}

	if queueLength < 1 {
		// There is no job? Stop now.
		return
	}

	body, err := redisClient.LPop("build-queue")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("An error occurred while getting job")
		return
	}

	job, err := buildjob.FromBytes(body)
	if err != nil {
		// there was a job we could not parse throw it away and stop here
		return
	}

	builder := newBuilder(job)

	// Try to get the lock for this job and quit if we don't get it
	if builder.AquireLock() != nil {
		builder.PutBackJob(false)
		return
	}

	// Prepare everything for the build or put back the job and stop if we can't
	if err = builder.PrepareBuild(); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("PrepareBuild failed")

		builder.PutBackJob(false)
		return
	}

	// Ensure we don't make a mess after we're done
	defer builder.Cleanup()

	// Do the real build
	if err := builder.Build(); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Build failed")

		builder.PutBackJob(true)
		return
	}

	// Handle the build log
	if err := builder.FetchBuildLog(); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Was unable to fetch the log from the container")

		builder.PutBackJob(false)
		return
	}

	if err := builder.WriteBuildLog(); err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Was unable to store the build log")
	}

	// If the build was marked as failed abort now
	if !builder.BuildOK {
		log.Fatal("Build was marked as failed, requeuing now.")

		builder.PutBackJob(true)

		// Send error notifications
		builder.SendNotifications()
		return
	}

	// Handle the uploads
	if builder.UploadRequired {
		if err := builder.UploadAssets(); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("Was unable to upload the build assets")

			builder.PutBackJob(false)
			return
		}
	}

	builder.UpdateBuildStatus(BuildStatusFinished, 0)

	if builder.UploadRequired {
		if err := builder.UpdateMetaData(); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatal("There was an error while updating metadata")

			builder.PutBackJob(false)
			return
		}
	}

	// Send success notifications
	builder.SendNotifications()
	builder.TriggerSubBuilds()
}
