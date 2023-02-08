package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip19"
	amqp "github.com/rabbitmq/amqp091-go"
)

var relays = []string{
	"wss://relay.snort.social",
	"wss://nos.lol",
	"wss://brb.io",
}

var secretKey = nostr.GeneratePrivateKey()

const (
	exchangeName = "lndhub_invoices"
	queueName    = "nostrifications"
)

type NostrificationSender struct {
	cfg NostrificationConfig
}

type NostrificationConfig struct {
	Pubkey string `envconfig:"NOSTR_PUBKEY"`
}

func NewNostrificationSender() *NostrificationSender {
	//todo clean up
	return &NostrificationSender{
		cfg: NostrificationConfig{
			Pubkey: os.Getenv("NOSTR_PUBKEY"),
		},
	}
}

func (ns *NostrificationSender) Handle(ctx context.Context, msg amqp.Delivery) error {
	payload := &Invoice{}
	err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(payload)
	if err != nil {
		return err
	}
	return sendPaymentNotification(int(payload.Amount), payload.Memo, ns.cfg.Pubkey, payload.Type)
}

func (ns *NostrificationSender) StartRabbit(ctx context.Context) (<-chan (amqp.Delivery), error) {
	conn, err := amqp.Dial(os.Getenv("AMQP_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	err = ch.QueueBind(
		q.Name,       // queue name
		"#",          // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
}
func sendPaymentNotification(amount int, msg, dest, invoiceType string) error {
	pk, _ := nostr.GetPublicKey(secretKey)
	_, theirPk, err := nip19.Decode(dest)
	theirHexPk := fmt.Sprintf("%v", theirPk)
	if err != nil {
		return err
	}
	ss, err := nip04.ComputeSharedSecret(theirHexPk, secretKey)
	if err != nil {
		return err
	}
	firstWord := "Received"
	if invoiceType == "outgoing" {
		firstWord = "Sent"
	}
	encrypted, err := nip04.Encrypt(fmt.Sprintf("💸 %s a %d sat payment. Message: %s 💸", firstWord, amount, msg), ss)
	if err != nil {
		fmt.Println(err)
		return err
	}
	ev := nostr.Event{
		ID:        "",
		PubKey:    pk,
		CreatedAt: time.Now(),
		Kind:      4,
		Tags:      nostr.Tags{[]string{"p", fmt.Sprintf("%v", theirHexPk)}},
		Content:   encrypted,
		Sig:       "",
	}

	// calling Sign sets the event ID field and the event Sig field
	ev.Sign(secretKey)

	// publish the event to two relays
	for _, url := range relays {
		relay, e := nostr.RelayConnect(context.Background(), url)
		if e != nil {
			fmt.Println(e)
			continue
		}
		fmt.Println("published to ", url, relay.Publish(context.Background(), ev))
	}
	return nil
}
