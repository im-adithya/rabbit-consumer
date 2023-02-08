package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kiwiidb/slack-go-webhook"
	"github.com/lightningnetwork/lnd/lnrpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SlackSender struct {
	rabbit    *RabbitClient
	Webhook   string
	EventType string
}

func NewSlackSender() (result *SlackSender, err error) {
	rabbit := &RabbitClient{}
	err = rabbit.Init()
	if err != nil {
		return nil, err
	}
	return &SlackSender{
		rabbit:  rabbit,
		Webhook: os.Getenv("SLACK_WEBHOOK_URL"),
	}, nil
}

func (s *SlackSender) Handle(ctx context.Context, msg amqp.Delivery) error {
	//parse msg
	chanEvent := &lnrpc.ChannelEventUpdate{}
	err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(chanEvent)
	if err != nil {
		return err
	}
	msgs := map[string]string{
		lnrpc.ChannelEventUpdate_OPEN_CHANNEL.String():   "Channel opened :zap:",
		lnrpc.ChannelEventUpdate_CLOSED_CHANNEL.String(): "Channel closed :no_entry_sign:",
	}

	//check type
	if _, ok := msgs[chanEvent.Type.String()]; !ok {
		//don't put these on slack
		return nil
	}
	attachment := createAttachment(chanEvent)
	payload := slack.Payload{
		Text:        msgs["type"],
		Attachments: []slack.Attachment{attachment},
		Username:    "Lightning Event Bot",
		Channel:     "notifications-ops",
		UnfurlMedia: true,
		UnfurlLinks: true,
		Markdown:    true,
		IconEmoji:   ":lightning:",
	}
	errs := slack.Send(s.Webhook, "", payload)
	if len(errs) > 0 {
		return fmt.Errorf("error: %s\n", err)
	}
	return nil
}

func createAttachment(chanEvent *lnrpc.ChannelEventUpdate) slack.Attachment {
	result := slack.Attachment{}
	switch chanEvent.Type {
	case lnrpc.ChannelEventUpdate_OPEN_CHANNEL:
		event := chanEvent.Channel.(*lnrpc.ChannelEventUpdate_OpenChannel)
		result.AddField(slack.Field{
			Title: "Remote peer pubkey",
			Value: event.OpenChannel.RemotePubkey,
		})
		result.AddField(slack.Field{
			Title: "Capacity",
			Value: strconv.Itoa(int(event.OpenChannel.Capacity)),
		})
	case lnrpc.ChannelEventUpdate_CLOSED_CHANNEL:
		event := chanEvent.Channel.(*lnrpc.ChannelEventUpdate_ClosedChannel)
		result.AddField(slack.Field{
			Title: "Remote peer pubkey",
			Value: event.ClosedChannel.RemotePubkey,
		})
		result.AddField(slack.Field{
			Title: "Capacity",
			Value: strconv.Itoa(int(event.ClosedChannel.Capacity)),
		})
	}
	return result
}

func (s *SlackSender) StartRabbit(ctx context.Context) (<-chan (amqp.Delivery), error) {
	return s.rabbit.ch.Consume(
		s.rabbit.cfg.RabbitMQQueueName, // queue
		"",                             // consumer
		false,                          // auto ack
		false,                          // exclusive
		false,                          // no local
		false,                          // no wait
		nil,                            // args
	)
}

func (s *SlackSender) CloseRabbit() {
	s.rabbit.Close()
}
