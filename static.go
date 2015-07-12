package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/flosch/pongo2"
)

func getBasicContext(r *http.Request) pongo2.Context {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")

	ctx := pongo2.Context{
		"gh_user": getGithubUsername(r),
	}

	if errorMessages := sess.Flashes("alert_error"); len(errorMessages) > 0 {
		ctx["error"] = errorMessages[0].(string)
	}

	if successMessages := sess.Flashes("alert_success"); len(successMessages) > 0 {
		ctx["success"] = successMessages[0].(string)
	}

	return ctx
}

func getNewBuildContext(r *http.Request) pongo2.Context {
	// Fetch clients active in last 5min
	timestamp := strconv.Itoa(int(time.Now().Unix() - 300))
	activeWorkers, _ := redisClient.ZCount("active-workers", timestamp, "+inf")

	queueLength, _ := redisClient.LLen("build-queue")
	lastBuilds, _ := redisClient.ZRevRange("last-builds", 0, 10, false)

	ctx := getBasicContext(r)

	ctx["queueLength"] = queueLength
	ctx["lastBuilds"] = lastBuilds
	ctx["activeWorkers"] = activeWorkers

	return ctx
}

func handleFrontPage(res http.ResponseWriter, r *http.Request) {
	template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
	template.ExecuteWriter(getNewBuildContext(r), res)
}

func handleImprint(res http.ResponseWriter, r *http.Request) {
	template := pongo2.Must(pongo2.FromFile("frontend/imprint.html"))
	template.ExecuteWriter(getBasicContext(r), res)
}

func handleHelpPage(res http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("frontend/help.md")
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%v", err),
		}).Error("HelpText Load")
		http.Error(res, "An unknown error occured.", http.StatusInternalServerError)
		return
	}

	template := pongo2.Must(pongo2.FromFile("frontend/help.html"))
	ctx := getBasicContext(r)
	ctx["helptext"] = string(content)

	template.ExecuteWriter(ctx, res)
}
