package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/robfig/cron"
	"gopkg.in/yaml.v2"
)

type repoWatch struct {
	WatchedRepos []watchedRepo `yaml:"watched_repos"`
}

type watchedRepo struct {
	Repo     string `yaml:"repo"`
	Interval string `yaml:"interval"`
}

type githubCommit struct {
	SHA string `json:"sha"`
}

func init() {
	watcher := &repoWatch{}
	go watcher.Run()
}

func (r *repoWatch) Run() {
	r.loadWatches()

	c := cron.New()
	c.AddFunc("@every 1m", r.iterateRepos)
	c.Start()

	for {
		select {}
	}
}

func (r *repoWatch) iterateRepos() {
	var threadLimit = make(chan bool, 10)

	for _, repo := range r.WatchedRepos {
		if r.isLocked(repo.Repo) {
			continue
		}

		threadLimit <- true

		go func(repo watchedRepo) {
			err := backoff.Retry(func() error {
				lastBuild, err := r.getLastBuild(repo.Repo)
				if err != nil {
					return err
				}

				lastCommit, err := r.getLastCommit(repo.Repo)
				if err != nil {
					return err
				}

				if !strings.HasPrefix(lastCommit, lastBuild) {
					urlString := "https://gobuilder.me/api/v1/webhook/cli"
					resp, err := http.PostForm(urlString, url.Values{
						"repository": []string{repo.Repo},
					})
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("GoBuilder did not respond with 200 OK to trigger")
					}

					log.Printf("Triggered build for repo %s", repo.Repo)
				}

				d, err := time.ParseDuration(repo.Interval)
				if err != nil {
					log.Errorf("Error while parsing interval for repo %s: %s", repo.Repo, err)
					d = time.Minute * 10
				}
				r.lockRepo(repo.Repo, d)

				return nil
			}, backoff.NewExponentialBackOff())

			if err != nil {
				log.WithFields(logrus.Fields{}).Printf("Error while fetching watched repo: %s", err)
			}

			<-threadLimit
		}(repo)
	}

	for len(threadLimit) > 0 {
		select {}
	}
}

func (r *repoWatch) isLocked(repo string) bool {
	redisKey := fmt.Sprintf("project::%s::repowatch", repo)
	lock, err := redisClient.Get(redisKey)
	if err != nil || string(lock) == "locked" {
		if err != nil {
			log.Errorf("Error while loading repowatch lock: %s", err)
		}
		return true
	}
	return false
}

func (r *repoWatch) lockRepo(repo string, d time.Duration) error {
	redisKey := fmt.Sprintf("project::%s::repowatch", repo)
	return redisClient.Set(redisKey, "locked", int(d.Seconds()), 0, false, false)
}

func (r *repoWatch) loadWatches() error {
	content, err := ioutil.ReadFile("repowatch.yml")
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, r)
}

func (r *repoWatch) getLastCommit(repo string) (string, error) {
	// https://api.github.com/repos/Luzifer/password/commits
	rex := regexp.MustCompile(`github\.com/([^/]+/[^/]+)/?`)
	repoParts := rex.FindStringSubmatch(repo)

	if len(repoParts) < 2 {
		return "", fmt.Errorf("Repo did not match GitHub-RegEx")
	}

	repoPart := repoParts[1]
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits", repoPart)
	result := []githubCommit{}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	if len(result) < 1 {
		return "", fmt.Errorf("Did not found one or more commits")
	}

	return result[0].SHA, nil
}

func (r *repoWatch) getLastBuild(repo string) (string, error) {
	redisKey := fmt.Sprintf("project::%s::built-commits", repo)
	commits, err := redisClient.ZRevRangeByScore(redisKey, "+inf", "-inf", true, true, 0, 1)
	if err != nil {
		return "", err
	}

	if len(commits) < 1 {
		return "", nil
	}

	return commits[0], nil
}
