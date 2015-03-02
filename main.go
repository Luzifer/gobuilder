package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/Luzifer/gobuilder/builddb"
	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
	"github.com/segmentio/go-loggly"

	_ "github.com/flosch/pongo2-addons"
)

var log *loggly.Client
var s3Bucket *s3.Bucket

func main() {
	log = loggly.New(os.Getenv("LOGGLY_TOKEN"))
	log.Tag("GoBuild-Frontend")

	connectS3()

	m := martini.Classic()
	m.Use(martini.Static("frontend"))

	// Static handlers
	m.Get("/", handleFrontPage)
	m.Get("/contact", handleImprint)
	m.Get("/help", handleHelpPage)

	// Build starters / webhooks
	m.Post("/build", webhookInterface)
	m.Post("/webhook/github", webhookGitHub)
	m.Post("/webhook/bitbucket", webhookBitBucket)

	// Build artifact displaying
	m.Get("/get/(?P<file>.*)$", handlerDeliverFileFromS3)
	m.Get("/(?P<repo>.*)/build.log$", handlerBuildLog)
	m.Get("/(?P<repo>.*)$", handlerRepositoryView)

	m.Run()
}

func handlerRepositoryView(params martini.Params, res http.ResponseWriter, r *http.Request) {
	branch := r.FormValue("branch")
	if branch == "" {
		branch = "master"
	}
	buildDBFile := fmt.Sprintf("%s/build.db", params["repo"])

	file, err := s3Bucket.Get(buildDBFile)
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

	build_status, err := s3Bucket.Get(fmt.Sprintf("%s/build.status", params["repo"]))
	if err != nil {
		build_status = []byte("unknown")
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
	branches := []builddb.BranchSortEntry{}
	for k, v := range buildDB {
		branches = append(branches, builddb.BranchSortEntry{Branch: k, BuildDate: v.BuildDate})
	}
	sort.Sort(sort.Reverse(builddb.BranchSortEntryByBuildDate(branches)))
	template.ExecuteWriter(pongo2.Context{
		"branch":       branch,
		"branches":     branches,
		"repo":         params["repo"],
		"mybranch":     buildDB[branch],
		"build_status": string(build_status),
	}, res)
}

func connectS3() {
	s3auth, err := aws.EnvAuth()
	if err != nil {
		panic(err)
	}

	s3conn := s3.New(s3auth, aws.Regions["eu-west-1"])
	bucket := s3conn.Bucket("gobuild.luzifer.io")

	s3Bucket = bucket
}
