package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/Sirupsen/logrus"
)

func webhookGitHub(res http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%v", err),
		}).Error("GitHub Hook Error")
		http.Error(res, "GitHub request could not be read.", http.StatusInternalServerError)
		return
	}

	var tmp interface{}
	json.Unmarshal([]byte(data), &tmp)
	repoName := tmp.(map[string]interface{})["repository"].(map[string]interface{})["full_name"].(string)
	commit := tmp.(map[string]interface{})["after"].(string)

	ref := tmp.(map[string]interface{})["ref"].(string)
	if ref != "refs/heads/master" {
		http.Error(res, "OK, got your message, will not take action.", http.StatusOK)
		return
	}

	if r.URL.Query().Get("override") != "" {
		repoName = r.URL.Query().Get("override")
	}

	repo := fmt.Sprintf("github.com/%s", repoName)
	err = sendToQueue(repo, commit)
	if err != nil {
		http.Error(res, "Could not submit build job", http.StatusInternalServerError)
	} else {
		http.Error(res, "OK", http.StatusOK)
	}
}

func webhookBitBucket(res http.ResponseWriter, r *http.Request) {
	data := r.FormValue("payload")

	var tmp interface{}
	json.Unmarshal([]byte(data), &tmp)
	repoName := tmp.(map[string]interface{})["repository"].(map[string]interface{})["absolute_url"].(string)
	repoName = strings.Trim(repoName, "/")

	repo := fmt.Sprintf("bitbucket.org/%s", repoName)
	err := sendToQueue(parseRepoCommit(repo))
	if err != nil {
		http.Error(res, "Could not submit build job", http.StatusInternalServerError)
	} else {
		http.Error(res, "OK", http.StatusOK)
	}
}

func webhookInterface(res http.ResponseWriter, r *http.Request) {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")
	repo, commit := parseRepoCommit(r.FormValue("repository"))

	// No repository was given, just submitted
	if len(repo) == 0 {
		sess.AddFlash("Please provide a repository.", "alert_error")
		sess.Save(r, res)
		http.Redirect(res, r, "/", http.StatusFound)
		return
	}

	// Repository contained characters not being allowed
	if !isValidRepositorySource(repo) {
		log.WithFields(logrus.Fields{
			"repository": repo,
		}).Warn("Refused to build repo")

		sess.AddFlash("Sorry, that does not look like a valid package. Not building that.", "alert_error")
		sess.Save(r, res)
		http.Redirect(res, r, "/", http.StatusFound)
		return
	}

	addGithubWebhook(res, r, repo)

	err := sendToQueue(repo, commit)
	if err != nil {
		sess.AddFlash("An unknown error occured while queueing the repository.", "alert_error")
		sess.Save(r, res)
		http.Redirect(res, r, "/", http.StatusFound)
		return
	}

	sess.AddFlash(flashContext{
		"success": "Your build job has been submitted.",
		"repo":    repo,
	}, "context")
	sess.Save(r, res)
	http.Redirect(res, r, "/", http.StatusFound)
}

func webhookCLI(res http.ResponseWriter, r *http.Request) {
	repo, commit := parseRepoCommit(r.FormValue("repository"))

	// No repository was given, just submitted
	if len(repo) == 0 {
		http.Error(res, "Please provide a repository", http.StatusNoContent)
		return
	}

	// Repository contained characters not being allowed
	if !isValidRepositorySource(repo) {
		log.WithFields(logrus.Fields{
			"repository": repo,
		}).Warn("Refused to build repo")

		http.Error(res, "Sorry, that does not look like a valid package. Not building that.", http.StatusNotAcceptable)
		return
	}

	err := sendToQueue(repo, commit)
	if err != nil {
		http.Error(res, "An unknown error occured while queueing the repository.", http.StatusInternalServerError)
		return
	}

	http.Error(res, "OK", http.StatusOK)
}

func isValidRepositorySource(repository string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9/_\.-]+/[a-zA-Z0-9/_\.-]+[^/]$`)
	return regex.Match([]byte(repository))
}

func sendToQueue(repository, commit string) error {
	job := buildjob.BuildJob{
		Repository:         repository,
		Commit:             commit,
		NumberOfExecutions: 0,
	}
	queueEntry, err := job.ToByte()
	if err != nil {
		log.Error(fmt.Sprintf("%q", err))
		return err
	}

	// Put the job into the queue and give it a time to run of 900 secs
	redisClient.RPush("build-queue", string(queueEntry))

	err = redisClient.Set(fmt.Sprintf("project::%s::build-status", repository), "queued", 0, 0, false, false)
	if err != nil {
		fmt.Printf("%+v", err)
	}

	return nil
}

func parseRepoCommit(repo string) (string, string) {
	t := strings.Split(repo, "@")
	if len(t) == 1 {
		return repo, ""
	}
	return t[0], t[1]
}
