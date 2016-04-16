package main

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	blockedRepos = &repoBlocklist{}
)

func init() {
	blockedRepos.LoadFromFile("blockedRepos.yml")
}

type repoBlocklist struct {
	Blocked []struct {
		NamePart string `yaml:"name"`
		Reason   string `yaml:"reason"`
	} `yaml:"blocked"`
}

func (r *repoBlocklist) LoadFromFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, r)
}

func (r *repoBlocklist) IsBlocked(repo string) (bool, string) {
	for _, blocked := range r.Blocked {
		if strings.Contains(repo, blocked.NamePart) {
			return true, blocked.Reason
		}
	}
	return false, ""
}
