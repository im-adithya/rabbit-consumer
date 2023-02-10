package main

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/kiwiidb/slack-go-webhook"
	"github.com/lightningnetwork/lnd/lnrpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TLV_WALLET_ID  = 6969669
	TLV_BOOSTAGRAM = 7629169
)

type InvoiceEventHandler struct {
	SlackClient *SlackClient
}

func NewInvoiceEventHandler() (result *InvoiceEventHandler, err error) {
	return &InvoiceEventHandler{
		SlackClient: DefaultSlackClient(),
	}, nil
}

func (s *InvoiceEventHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	invoice := &lnrpc.Invoice{}
	err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(invoice)
	if err != nil {
		return err
	}
	//check if it's keysend
	//and the user id TLV is absent
	//and the boostagram TLV is present
	recs := invoice.Htlcs[0].CustomRecords
	if invoice.IsKeysend && recs[TLV_WALLET_ID] == nil && recs[TLV_BOOSTAGRAM] != nil {
		attachment, err := createLostPodcasterAttachment(invoice)
		if err != nil {
			return err
		}
		return s.SlackClient.Send("A wandering podcaster :radio:. Show them the way!", attachment)
	}
	return nil
}

func createLostPodcasterAttachment(invoice *lnrpc.Invoice) (slack.Attachment, error) {
	//find boostagram
	result := slack.Attachment{}
	rawBoostagram := invoice.Htlcs[0].CustomRecords[TLV_BOOSTAGRAM]
	//parse boostagram
	boostagram := &Boostagram{}
	err := json.NewDecoder(bytes.NewBuffer(rawBoostagram)).Decode(boostagram)
	if err != nil {
		return result, err
	}
	//construct payload
	result.AddField(slack.Field{
		Title: "Podcast Name",
		Value: boostagram.Podcast,
	})
	result.AddField(slack.Field{
		Title: "Podcast URL",
		Value: boostagram.URL,
	})
	result.AddField(slack.Field{
		Title: "Podcaster Name",
		Value: boostagram.Name,
	})
	result.AddField(slack.Field{
		Title: "Episode Name",
		Value: boostagram.Episode,
	})
	result.AddField(slack.Field{
		Title: "App Name",
		Value: boostagram.AppName,
	})
	return result, nil
}

type Boostagram struct {
	AppName        string `json:"app_name"`
	AppVersion     string `json:"app_version"`
	ValueMsatTotal int    `json:"value_msat_total"`
	URL            string `json:"url"`
	Podcast        string `json:"podcast"`
	Action         string `json:"action"`
	Episode        string `json:"episode"`
	EpisodeGUID    string `json:"episode_guid"`
	ValueMsat      int    `json:"value_msat"`
	Ts             int    `json:"ts"`
	Name           string `json:"name"`
	SenderName     string `json:"sender_name"`
}
