package notifier

import (
	"fmt"

	"github.com/Luzifer/gobuilder/config"
	"github.com/thorduri/pushover"
)

// NotifyPushover uses the Pushover API to send notifications about the current build
func (n *NotifyEntry) NotifyPushover(metadata NotifyMetaData, cfg *config.Config) error {
	verb := "succeeded"
	if metadata.EventType == "error" {
		verb = "failed"
	}

	message := &pushover.Message{
		Message:  fmt.Sprintf("The build for repo %s %s", metadata.Repository, verb),
		Title:    fmt.Sprintf("GoBuilder.me %s", verb),
		Url:      fmt.Sprintf("https://gobuilder.me/%s", metadata.Repository),
		UrlTitle: "Go to your build...",
		Priority: pushover.Normal,
	}

	p, err := pushover.NewPushover(cfg.Pushover.APIToken, n.Target)
	if err != nil {
		return err
	}

	_, _, err = p.Push(message)
	if err != nil {
		return err
	}

	return nil
}
