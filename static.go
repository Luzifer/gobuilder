package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/flosch/pongo2"
)

func handleFrontPage(res http.ResponseWriter) {
	// Fetch clients active in last 10min
	timestamp := string(time.Now().Unix() - 600)
	activeWorkers, _ := redisClient.ZCount("active-workers", timestamp, "+inf")

	queueLength, _ := redisClient.LLen("build-queue")
	lastBuilds, _ := redisClient.LRange("last-builds", 0, 10)

	template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
	template.ExecuteWriter(pongo2.Context{
		"queueLength":   queueLength,
		"lastBuilds":    lastBuilds,
		"activeWorkers": activeWorkers,
	}, res)
}

func handleImprint(res http.ResponseWriter) {
	template := pongo2.Must(pongo2.FromFile("frontend/imprint.html"))
	template.ExecuteWriter(pongo2.Context{}, res)
}

func handleHelpPage(res http.ResponseWriter) {
	content, err := ioutil.ReadFile("frontend/help.md")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%v", err),
		}).Error("HelpText Load")
		http.Error(res, "An unknown error occured.", http.StatusInternalServerError)
		return
	}
	template := pongo2.Must(pongo2.FromFile("frontend/help.html"))
	template.ExecuteWriter(pongo2.Context{
		"helptext": string(content),
	}, res)
}
