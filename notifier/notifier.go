package notifier

import "strings"

// NotifyEntry represents a configuration for a single notification method
type NotifyEntry struct {
	// Type can be one of these: "dockerhub"
	Type string `yaml:"type"`
	// Target represents a target expected by the notification Type
	Target string `yaml:"target"`
	// Filter determines whether to send a notification. Expected is a comma seperated list of EventTypes (e.g. "success,error")
	Filter string `yaml:"filter,omitempty"`
}

// NotifyMetaData contains information from the build process about the
// build for the notification process
type NotifyMetaData struct {
	EventType  string
	Repository string
}

// NotifyConfiguration represents a list of notification methods
type NotifyConfiguration []NotifyEntry

// Execute iterates over all configured notification methods and calls
// the respective methods
func (n *NotifyConfiguration) Execute(metadata NotifyMetaData) error {
	for _, method := range *n {
		if len(strings.TrimSpace(method.Filter)) == 0 || strings.Contains(method.Filter, metadata.EventType) {
			var err error
			switch method.Type {
			case "dockerhub":
				err = method.NotifyDockerHub(metadata)
			case "pushover":
				err = method.NotifyPushover(metadata)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
