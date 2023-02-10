package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiwiidb/slack-go-webhook"
	"github.com/lightningnetwork/lnd/lnrpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TLV_WALLET_ID = 6969669
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
	//check if it's keysend and if TLV 696969 is absent
	if invoice.IsKeysend && invoice.Htlcs[0].CustomRecords[TLV_WALLET_ID] == nil {
		//attachment := createLostPodcasterAttachment(invoice)
		//return s.SlackClient.Send("A wandering podcaster. Who will show them on their way?", attachment)
		fmt.Println(invoice)
	}
	return nil
}

func createLostPodcasterAttachment(invoice *lnrpc.Invoice) slack.Attachment {
	//find boostagram
	//parse boostagram
	//construct payload
	result := slack.Attachment{}
	return result
}
