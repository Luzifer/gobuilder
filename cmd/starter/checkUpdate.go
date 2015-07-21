package main

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

func init() {
	go checkUpdateAndRestart()
}

func checkUpdateAndRestart() {
	for {
		<-time.After(5 * time.Minute)

		lastBuild, _ := redisClient.Get("project::github.com/Luzifer/gobuilder/cmd/starter::last-build")
		newVersion := strings.TrimSpace(string(lastBuild))
		if len(lastBuild) > 0 && version != "dev" && newVersion != version {
			log.WithFields(logrus.Fields{
				"host":        hostname,
				"new_version": newVersion,
			}).Infof("Detected update from %s to %s, will quit soon.", version, newVersion)
			killswitch = true
		}
	}
}
