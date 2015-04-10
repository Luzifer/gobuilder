package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/buildjob"
	"github.com/Sirupsen/logrus"
	"github.com/flosch/pongo2"
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

	repo := fmt.Sprintf("github.com/%s", repoName)
	err = sendToQueue(repo)
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
	err := sendToQueue(repo)
	if err != nil {
		http.Error(res, "Could not submit build job", http.StatusInternalServerError)
	} else {
		http.Error(res, "OK", http.StatusOK)
	}
}

func webhookInterface(res http.ResponseWriter, r *http.Request) {
	repo := r.FormValue("repository")
	template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
	if len(repo) == 0 {
		template.ExecuteWriter(pongo2.Context{
			"error": "Please provide a repository.",
		}, res)
	} else {
		if !isValidRepositorySource(repo) {
			template.ExecuteWriter(pongo2.Context{
				"error": "Sorry, that does not look like a valid package. Not building that.",
			}, res)
			log.WithFields(logrus.Fields{
				"repository": repo,
			}).Warn("Refused to build repo")
			return
		}
		err := sendToQueue(repo)
		if err != nil {
			template.ExecuteWriter(pongo2.Context{
				"error": "An unknown error occured while queueing the repository.",
			}, res)
		} else {
			template.ExecuteWriter(pongo2.Context{
				"success": "Your build job has been submitted.",
				"repo":    repo,
			}, res)
		}
	}
}

func isValidRepositorySource(repository string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9/_\.-]+[^/]$`)
	return regex.Match([]byte(repository))
}

func sendToQueue(repository string) error {
	job := buildjob.BuildJob{
		Repository:         repository,
		NumberOfExecutions: 0,
	}
	queueEntry, err := job.ToByte()
	if err != nil {
		log.Error(fmt.Sprintf("%q", err))
		return err
	}

	// Put the job into the queue and give it a time to run of 900 secs
	redisClient.RPush("build-queue", string(queueEntry))

	err = s3Bucket.Put(fmt.Sprintf("%s/build.status", repository), []byte("queued"), "text/plain", s3.PublicRead)
	if err != nil {
		fmt.Printf("%+v", err)
	}

	return nil
}
