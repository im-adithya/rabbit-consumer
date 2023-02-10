package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/kiwiidb/slack-go-webhook"
	"github.com/lightningnetwork/lnd/lnrpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SlackChannelEventSender struct {
	SlackClient *SlackClient
}

func NewSlackSender() (result *SlackChannelEventSender, err error) {
	return &SlackChannelEventSender{
		SlackClient: &SlackClient{
			Webhook:   os.Getenv("SLACK_WEBHOOK_URL"),
			Channel:   "#notifications-ops",
			Name:      "Lightning Event Bot",
			IconEmoji: ":lightning:",
		},
	}, nil
}

func (s *SlackChannelEventSender) Handle(ctx context.Context, msg amqp.Delivery) error {
	//parse msg
	//TODO: this doesn't work, there is an issue deserializing the channel event payload
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
	attachment := createChannelEventAttachment(chanEvent)
	return s.SlackClient.Send(msgs["type"], attachment)
}

func createChannelEventAttachment(chanEvent *lnrpc.ChannelEventUpdate) slack.Attachment {
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
