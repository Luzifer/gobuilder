package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Luzifer/go-openssl"
	"github.com/Luzifer/gobuilder/builddb"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

func registerAPIv1(router *mux.Router) {
	r := router.PathPrefix("/api/v1/").Subrouter()

	// Add build starters
	r.HandleFunc("/build", webhookInterface).Methods("POST")
	r.HandleFunc("/webhook/github", webhookGitHub).Methods("POST")
	r.HandleFunc("/webhook/bitbucket", webhookBitBucket).Methods("POST")
	r.HandleFunc("/webhook/cli", webhookCLI).Methods("POST")

	r.HandleFunc("/{repo:.+}/last-build", apiV1HandlerLastBuild).Methods("GET")
	r.HandleFunc("/{repo:.+}/already-built", apiV1HandlerAlreadyBuilt).Methods("GET")
	r.HandleFunc("/{repo:.+}/signed-hashes/{tag}", apiV1HandlerSignedHashes).Methods("GET")
	r.HandleFunc("/{repo:.+}/hashes/{tag}.{format:[a-z]+}", apiV1HandlerHashes).Methods("GET")
	r.HandleFunc("/{repo:.+}/rebuild", apiV1HandlerRebuild).Methods("GET")
	r.HandleFunc("/{repo:.+}/build.db", apiV1HandlerBuildDb).Methods("GET")
	r.HandleFunc("/{repo:.+}/encrypt", apiV1HandlerEncrypt).Methods("POST")
}

func apiV1HandlerAlreadyBuilt(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::built-commits", vars["repo"])
	commit := r.URL.Query().Get("commit")

	if len(commit) == 0 {
		http.Error(res, "You must pass a commit!", http.StatusBadRequest)
		return
	}

	rank, err := redisClient.ZRank(redisKey, commit)
	if err != nil {
		http.Error(res, "An error ocurred.", http.StatusInternalServerError)
		log.WithFields(logrus.Fields{
			"error":  err,
			"repo":   vars["repo"],
			"commit": commit,
		}).Error("Failed to read commit rank")
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	if rank == -1 {
		res.Write([]byte("false"))
	} else {
		res.Write([]byte(commit))
	}
}

func apiV1HandlerEncrypt(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::encryption-key", vars["repo"])

	encryptionKey, err := redisClient.Get(redisKey)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"repo":  vars["repo"],
		}).Error("Failed to get encryption key")
		http.Error(res, "Could not read encryption key", http.StatusInternalServerError)
		return
	}

	if string(encryptionKey) == "" {
		encryptionKey = []byte(uuid.NewV4().String())
		redisClient.Set(redisKey, string(encryptionKey), 0, 0, false, false)
	}

	o := openssl.New()
	enc, err := o.EncryptString(string(encryptionKey), r.FormValue("secret"))

	res.Header().Set("Content-Type", "text/plain")
	res.Write(enc)
}

func apiV1HandlerLastBuild(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::built-commits", vars["repo"])
	commits, err := redisClient.ZRevRangeByScore(redisKey, "+inf", "-inf", true, true, 0, 1)
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
	res.Write([]byte(commits[0]))
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

func apiV1HandlerHashes(res http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	redisKey := fmt.Sprintf("project::%s::hashes_yml::%s", vars["repo"], vars["tag"])
	hashList, err := redisClient.Get(redisKey)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"repo":  vars["repo"],
		}).Error("Failed to get hash list")
		http.Error(res, "Could not read hash list", http.StatusInternalServerError)
		return
	}

	switch vars["format"] {
	case "yaml":
		res.Header().Add("Content-Type", "application/x-yaml")
		res.Header().Add("Cache-Control", "no-cache")
		res.Write(hashList)
	case "json":
		out := builddb.HashDB{}
		if err := yaml.Unmarshal(hashList, &out); err != nil {
			http.Error(res, "Could not parse hash list", http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(out)
		if err != nil {
			http.Error(res, "Could not encode hash list", http.StatusInternalServerError)
			return
		}

		res.Header().Add("Content-Type", "application/json")
		res.Header().Add("Cache-Control", "no-cache")
		res.Write(data)
	default:
		http.Error(res, "Not found", http.StatusNotFound)
	}
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
	sendToQueue(parseRepoCommit(vars["repo"]))

	http.Redirect(res, r, fmt.Sprintf("/%s", vars["repo"]), http.StatusFound)
}
