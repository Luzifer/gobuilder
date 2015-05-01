package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func handlerDeliverFileFromS3(res http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	t := time.Now()
	t = t.Add(1 * time.Hour)
	http.Redirect(res, r, s3Bucket.SignedURL(params["file"], t), http.StatusTemporaryRedirect)
}
