package main

import (
	"fmt"
	"os"

	"github.com/kiwiidb/slack-go-webhook"
)

type SlackClient struct {
	Webhook   string
	Channel   string
	Name      string
	IconEmoji string
}

func DefaultSlackClient() (result *SlackClient) {
	return &SlackClient{
		Webhook:   os.Getenv("SLACK_WEBHOOK_URL"),
		Channel:   "#notifications-ops",
		Name:      "Lightning Event Bot",
		IconEmoji: ":lightning:",
	}
}

func (sc *SlackClient) Send(text string, attachment slack.Attachment) error {
	payload := slack.Payload{
		Text:        text,
		Attachments: []slack.Attachment{attachment},
		Username:    sc.Name,
		Channel:     sc.Channel,
		UnfurlMedia: true,
		UnfurlLinks: true,
		Markdown:    true,
		IconEmoji:   sc.IconEmoji,
	}
	errs := slack.Send(sc.Webhook, "", payload)
	if len(errs) > 0 {
		return fmt.Errorf("error: %s\n", errs[0])
	}
	return nil
}
