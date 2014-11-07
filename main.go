package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/builddb"
	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
	"github.com/kr/beanstalk"
	"github.com/segmentio/go-loggly"

	_ "github.com/flosch/pongo2-addons"
)

var log *loggly.Client

func main() {
	log = loggly.New(os.Getenv("LOGGLY_TOKEN"))
	log.Tag("GoBuild-Frontend")

	m := martini.Classic()
	m.Use(martini.Static("frontend"))

	m.Get("/", func(res http.ResponseWriter) {
		template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
		template.ExecuteWriter(pongo2.Context{}, res)
	})

	m.Post("/build", func(res http.ResponseWriter, r *http.Request) {
		repo := r.FormValue("repository")
		template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
		if len(repo) == 0 {
			template.ExecuteWriter(pongo2.Context{
				"error": "Please provide a repository.",
			}, res)
		} else {
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
	})

	m.Post("/webhook/github", func(res http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("GitHub Hook Error", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
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
	})

	m.Post("/webhook/bitbucket", func() {
		// TODO: Handle BitBucket Hooks
	})

	m.Get("/(?P<repo>.*)$", func(params martini.Params, res http.ResponseWriter, r *http.Request) {
		branch := r.FormValue("branch")
		if branch == "" {
			branch = "master"
		}
		buildDBFile := fmt.Sprintf("%s/build.db", params["repo"])

		s3auth, err := aws.EnvAuth()
		if err != nil {
			log.Error("AWS Authentication Error", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
			template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
			template.ExecuteWriter(pongo2.Context{
				"error": "An unknown error occured while getting your build.",
			}, res)
			return
		}

		s3conn := s3.New(s3auth, aws.Regions["eu-west-1"])
		bucket := s3conn.Bucket("gobuild.luzifer.io")

		file, err := bucket.Get(buildDBFile)
		if err != nil {
			log.Error("AWS S3 Get Error", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
			template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
			template.ExecuteWriter(pongo2.Context{
				"error": "Your build is not yet known to us...",
				"value": params["repo"],
			}, res)
			return
		}

		var buildDB builddb.BuildDB
		err = json.Unmarshal(file, &buildDB)
		if err != nil {
			log.Error("AWS DB Unmarshal Error", loggly.Message{
				"error": fmt.Sprintf("%v", err),
			})
			template := pongo2.Must(pongo2.FromFile("frontend/newbuild.html"))
			template.ExecuteWriter(pongo2.Context{
				"error": "An unknown error occured while getting your build.",
			}, res)
			return
		}

		template := pongo2.Must(pongo2.FromFile("frontend/repository.html"))
		template.ExecuteWriter(pongo2.Context{
			"branch":   branch,
			"builddb":  buildDB,
			"repo":     params["repo"],
			"mybranch": buildDB[branch],
		}, res)
	})

	m.Run()
}

func sendToQueue(repository string) error {
	conn, err := beanstalk.Dial("tcp", os.Getenv("BEANSTALK_ADDR"))
	if err != nil {
		log.Error("Beanstalk-Connect", loggly.Message{
			"error": fmt.Sprintf("%v", err),
		})
		return err
	}

	defer conn.Close()

	t := beanstalk.Tube{
		Conn: conn,
		Name: "gobuild.luzifer.io",
	}
	// Put the job into the queue and give it a time to run of 300 secs
	_, err = t.Put([]byte(repository), 1, 0, 300*time.Second)
	if err != nil {
		log.Error(fmt.Sprintf("%q", err))
		return err
	}

	return nil
}
