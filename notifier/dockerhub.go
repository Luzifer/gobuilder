package notifier

import (
	"errors"
	"net/http"
	"net/url"
)

// NotifyDockerHub calls a DockerHub webhook to build a container after building the artifacts
func (n *NotifyEntry) NotifyDockerHub(metadata NotifyMetaData) error {
	if metadata.EventType != "success" {
		return nil
	}

	resp, err := http.PostForm(n.Target, url.Values{"build": []string{"true"}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	return errors.New(resp.Status)
}
