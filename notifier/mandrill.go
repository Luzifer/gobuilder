package notifier

import (
	"fmt"

	"github.com/Luzifer/gobuilder/config"
	"github.com/keighl/mandrill"
)

func (n *NotifyEntry) NotifyEMail(metadata NotifyMetaData, cfg *config.Config) error {
	verb := "successful"
	if metadata.EventType == "error" {
		verb = "failed"
	}

	c := mandrill.ClientWithKey(cfg.Mandrill.MandrillAPIKey)
	message := &mandrill.Message{}
	message.AddRecipient(n.Target, "Gopher", "to")
	message.GlobalMergeVars = mandrill.MapToVars(map[string]interface{}{
		"REPO":      metadata.Repository,
		"STATE":     verb,
		"REPO_LINK": fmt.Sprintf("https://gobuilder.me/%s", metadata.Repository),
		"EMAIL":     n.Target,
	})
	message.AutoText = true
	message.TrackOpens = true
	message.TrackClicks = true
	message.InlineCSS = true

	templateContent := map[string]string{}
	_, err := c.MessagesSendTemplate(message, "gobuilder-status-mail", templateContent)

	return err
}
