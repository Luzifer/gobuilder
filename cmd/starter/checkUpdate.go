package main

import "time"

func init() {
	go checkUpdateAndRestart()
}

func checkUpdateAndRestart() {
	for {
		<-time.After(5 * time.Minute)

		lastBuild, _ := redisClient.Get("project::github.com/Luzifer/gobuilder/cmd/starter::last-build")
		if len(lastBuild) > 0 && version != "dev" && string(lastBuild) != version {
			log.WithField("new_version", string(lastBuild)).
				Infof("Detected update from %s to %s, will quit soon.", version, string(lastBuild))
			killswitch = true
		}
	}
}
