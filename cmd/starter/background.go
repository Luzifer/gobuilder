package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

func announceActiveWorker() {
	timestamp := float64(time.Now().Unix())
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(logrus.Fields{
			"host":  hostname,
			"error": err,
		}).Error("Unable to determine hostname")
	}

	redisClient.ZAdd("active-workers", map[string]float64{
		hostname: timestamp,
	})

	// Remove old clients to ensure the redis doesn't get filled with old data
	redisClient.ZRemRangeByScore("active-workers", "-inf", strconv.Itoa(int(time.Now().Unix()-3600)))
}

func pullLatestImage() error {
	auth := docker.AuthConfiguration{}
	authConfig, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}

	reginfo := strings.SplitN(conf.BuildImage.ImageName, "/", 2)
	if len(reginfo) == 2 {
		for s, a := range authConfig.Configs {
			if strings.Contains(s, fmt.Sprintf("://%s/", reginfo[0])) {
				auth = a
			}
		}
	}

	err = dockerClient.PullImage(docker.PullImageOptions{
		Repository: conf.BuildImage.ImageName,
		Tag:        "latest",
	}, auth)

	return err
}
