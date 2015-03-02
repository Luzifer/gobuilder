package main

import (
	"net/http"
	"time"

	"github.com/go-martini/martini"
)

func handlerDeliverFileFromS3(params martini.Params, res http.ResponseWriter, r *http.Request) {
	t := time.Now()
	t = t.Add(1 * time.Hour)
	http.Redirect(res, r, s3bucket.SignedURL(params["file"], t), http.StatusTemporaryRedirect)
}
