package main

import (
	"strings"
	"time"
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
			log.WithField("new_version", newVersion).
				Infof("Detected update from %s to %s, will quit soon.", version, newVersion)
			killswitch = true
		}
	}
}
