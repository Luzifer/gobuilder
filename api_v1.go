package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func registerAPIv1(router *mux.Router) {
	r := router.PathPrefix("/api/v1/").Subrouter()

	// Add build starters
	r.HandleFunc("/build", webhookInterface).Methods("POST")
	r.HandleFunc("/webhook/github", webhookGitHub).Methods("POST")
	r.HandleFunc("/webhook/bitbucket", webhookBitBucket).Methods("POST")

	r.HandleFunc("/{repo:.+}/last-build", apiV1HandlerLastBuild).Methods("GET")
	r.HandleFunc("/{repo:.+}/signed-hashes/{tag}", apiV1HandlerSignedHashes).Methods("GET")
	r.HandleFunc("/{repo:.+}/rebuild", apiV1HandlerRebuild).Methods("GET")
	r.HandleFunc("/{repo:.+}/build.db", apiV1HandlerBuildDb).Methods("GET")
}

func apiV1HandlerLastBuild(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::last-build", vars["repo"])
	lastBuild, err := redisClient.Get(redisKey)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"repo":  vars["repo"],
		}).Error("Failed to get last build hash")
		http.Error(res, "Could not read last build hash", http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Cache-Control", "no-cache")
	res.Write(lastBuild)
}

func apiV1HandlerSignedHashes(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::hashes::%s", vars["repo"], vars["tag"])
	hashList, err := redisClient.Get(redisKey)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"repo":  vars["repo"],
		}).Error("Failed to get signed hash list")
		http.Error(res, "Could not read signed hash list", http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("Cache-Control", "no-cache")
	res.Write(hashList)
}

func apiV1HandlerBuildDb(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildDB, err := getBuildDBWithFallback(vars["repo"])
	if err != nil {
		http.Error(res, "Could not read build.db", http.StatusInternalServerError)
	}

	res.Header().Add("Content-Type", "application/json")
	res.Header().Add("Cache-Control", "no-cache")
	res.Write(buildDB)
}

func apiV1HandlerRebuild(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sendToQueue(vars["repo"])

	http.Redirect(res, r, fmt.Sprintf("/%s", vars["repo"]), http.StatusFound)
}
