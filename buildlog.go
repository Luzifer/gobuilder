package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
	"github.com/segmentio/go-loggly"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
)

func handlerBuildLog(params martini.Params, res http.ResponseWriter, r *http.Request) {
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

	file, err := bucket.Get(fmt.Sprintf("%s/build.log", params["repo"]))
	if err != nil {
		file = []byte("No build log was found for this build.")
	}

	template := pongo2.Must(pongo2.FromFile("frontend/buildlog.html"))
	template.ExecuteWriter(pongo2.Context{
		"repo": params["repo"],
		"log":  logHighlight(file),
	}, res)

}

type logline struct {
	Line         string
	BuildComment bool
}

func logHighlight(log []byte) []logline {
	lines := strings.Split(string(log), "\n")
	highlightedLines := []logline{}
	for _, line := range lines {
		tmp := logline{
			Line:         line,
			BuildComment: false,
		}
		if strings.HasPrefix(line, "[") {
			tmp.BuildComment = true
		}
		highlightedLines = append(highlightedLines, tmp)
	}
	return highlightedLines
}
