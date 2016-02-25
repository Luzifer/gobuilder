package notifier

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Luzifer/gobuilder/config"
	"github.com/flosch/pongo2"
)

// NotifyEMail utilizes the Mandrill API to send a predefined template for the
// current build
func (n *NotifyEntry) NotifyEMail(metadata NotifyMetaData, cfg *config.Config) error {
	verb := "successful"
	if metadata.EventType == "error" {
		verb = "failed"
	}

	ctx := pongo2.Context{
		"state":     verb,
		"repo":      metadata.Repository,
		"email":     n.Target,
		"repo_link": fmt.Sprintf("https://gobuilder.me/%s", metadata.Repository),
	}
	template := pongo2.Must(pongo2.FromFile("frontend/repository.html"))
	mailContent, err := template.Execute(ctx)
	if err != nil {
		return err
	}

	params := url.Values{
		"from":    []string{"GoBuilder.me <help@gobuilder.me>"},
		"to":      []string{n.Target},
		"html":    []string{mailContent},
		"subject": []string{"Information for your build of " + metadata.Repository},
	}
	req, _ := http.NewRequest("POST", "https://api.mailgun.net/v3/gobuilder.me/messages", bytes.NewBuffer([]byte(params.Encode())))
	req.SetBasicAuth("api", cfg.MailGun.MailGunAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New("Non successfull mail delivery")
	}

	return nil
}
