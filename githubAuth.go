package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"gopkg.in/bufio.v1"
)

type githubHook struct {
	Name   string           `json:"name"`
	Active bool             `json:"active"`
	Events []string         `json:"events"`
	Config githubHookConfig `json:"config"`
}
type githubHookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
}

func handleOauthGithubInit(res http.ResponseWriter, r *http.Request) {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")

	scopes := strings.Join([]string{
		"write:repo_hook", // We want to write the webhook to execute a gobuilder build
	}, ",")

	if code := r.URL.Query().Get("code"); code != "" {
		state := sess.Values["gh_oauth_state"].(string)

		if r.URL.Query().Get("state") != state {
			sess.AddFlash("Something went wrong while authenticating you. Please try again.", "alert_error")
			sess.Save(r, res)
			http.Redirect(res, r, "/", http.StatusTemporaryRedirect)
			return
		}

		resp, err := http.PostForm("https://github.com/login/oauth/access_token", url.Values{
			"client_id":     {os.Getenv("github_client_id")},
			"client_secret": {os.Getenv("github_client_secret")},
			"code":          {code},
			"state":         {state},
		})
		if err != nil {
			log.WithField("error", err).Error("Fetching OAuth2 AccessKey failed")
			sess.AddFlash("Something went wrong while authenticating you. Please try again.", "alert_error")
			sess.Save(r, res)
			http.Redirect(res, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		accessInformation, _ := url.ParseQuery(string(body))

		if accessInformation.Get("scope") != scopes {
			sess.AddFlash("You denied some access rights. Unable to work that way.", "alert_error")
			sess.Save(r, res)
			http.Redirect(res, r, "/", http.StatusTemporaryRedirect)
			return
		}

		sess.Values["access_token"] = accessInformation.Get("access_token")
		sess.Save(r, res)
		http.Redirect(res, r, "/", http.StatusTemporaryRedirect)
		return
	}

	query := url.Values{}
	redirURL, _ := url.Parse("https://github.com/login/oauth/authorize")
	query.Set("client_id", os.Getenv("github_client_id"))
	query.Set("scope", scopes)

	u := uuid.NewV4().String()
	sess.Values["gh_oauth_state"] = u
	query.Set("state", u)
	sess.Save(r, res)

	redirURL.RawQuery = query.Encode()
	http.Redirect(res, r, redirURL.String(), http.StatusTemporaryRedirect)
}

func handleOauthGithubLogout(res http.ResponseWriter, r *http.Request) {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")
	delete(sess.Values, "access_token")
	sess.Save(r, res)
	http.Redirect(res, r, "/", http.StatusTemporaryRedirect)
}

func getGithubUsername(r *http.Request) string {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")

	if _, ok := sess.Values["access_token"].(string); !ok {
		return ""
	}

	resp, err := http.Get("https://api.github.com/user?access_token=" + sess.Values["access_token"].(string))
	if err != nil {
		log.WithField("error", err.Error()).Error("Unable to fetch GitHub username")
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("GitHub Status %d", resp.StatusCode)
		return ""
	}

	d := struct {
		Login string `json:"login"`
	}{}
	json.NewDecoder(resp.Body).Decode(&d)

	return d.Login
}

func addGithubWebhook(res http.ResponseWriter, r *http.Request, repo string) {
	sess, _ := sessionStore.Get(r, "GoBuilderSession")
	hookURL := strings.TrimRight(os.Getenv("baseurl"), "/") + "/api/v1/webhook/github"

	if token, ok := sess.Values["access_token"].(string); !ok || len(token) == 0 {
		return
	}

	re := regexp.MustCompile("^github.com/([^/]+)/([^/]+)")
	if !re.MatchString(repo) {
		log.WithField("repo", repo).Error("Tried to add webhook to non-github-repo")
		return
	}

	matches := re.FindStringSubmatch(repo)
	owner := matches[1]
	repomatch := matches[2]

	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks?access_token=%s", owner, repomatch, sess.Values["access_token"].(string)))
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"repo":  repo,
		}).Error("Unable to fetch hooks for Repo")
		sess.AddFlash("Could not access the hooks for your repository.", "alert_error")
		sess.Save(r, res)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("GitHub Status %d", resp.StatusCode)
		return
	}

	t := []githubHook{}
	json.NewDecoder(resp.Body).Decode(&t)

	for _, v := range t {
		if v.Config.URL == hookURL {
			// We found our hook, we're happy
			return
		}
	}

	hook := githubHook{
		Name:   "web",
		Active: true,
		Events: []string{"push"},
		Config: githubHookConfig{
			URL:         hookURL,
			ContentType: "json",
		},
	}
	body, _ := json.Marshal(hook)
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repomatch), bufio.NewBuffer(body))
	req.Header.Set("Authorization", "token "+sess.Values["access_token"].(string))
	setResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"repo":  repo,
		}).Error("Unable to set hook for Repo")
		sess.AddFlash("Could not set the hook for your repository.", "alert_error")
		sess.Save(r, res)
		return
	}
	defer setResp.Body.Close()

	if setResp.StatusCode != 201 {
		body, _ = ioutil.ReadAll(setResp.Body)
		log.WithFields(logrus.Fields{
			"repo":   repo,
			"status": strconv.FormatInt(int64(setResp.StatusCode), 10),
		}).Error("Unable to set hook for Repo")
		sess.AddFlash("Could not set the hook for your repository.", "alert_error")
		sess.Save(r, res)
	}
}
