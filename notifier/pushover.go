package notifier

import (
	"fmt"
	"os"

	"github.com/thorduri/pushover"
)

func (n *NotifyEntry) NotifyPushover(metadata NotifyMetaData) error {
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

	p, err := pushover.NewPushover(os.Getenv("PUSHOVER_APPTOKEN"), n.Target)
	if err != nil {
		return err
	}

	_, _, err = p.Push(message)
	if err != nil {
		return err
	}

	return nil
}
