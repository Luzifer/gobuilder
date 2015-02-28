package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
	"github.com/segmentio/go-loggly"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
)

func handlerDeliverFileFromS3(params martini.Params, res http.ResponseWriter, r *http.Request) {
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

	t := time.Now()
	t = t.Add(1 * time.Hour)
	http.Redirect(res, r, bucket.SignedURL(params["file"], t), http.StatusTemporaryRedirect)
}
