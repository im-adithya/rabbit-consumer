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

func main() {
	destinationPubkey := os.Getenv("NOSTR_DESTINATION_PUBKEY")
	conn, err := amqp.Dial(os.Getenv("AMQP_CONNECTION_STRING"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"nostrifications", // name
		false,             // durable
		false,             // delete when unused
		true,              // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	err = ch.QueueBind(
		q.Name,            // queue name
		"#",               // routing key
		"lndhub_invoices", // exchange
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("listening...")
	for msg := range msgs {
		fmt.Println("received msg")
		payload := &Invoice{}
		err = json.NewDecoder(bytes.NewReader(msg.Body)).Decode(payload)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(payload)
		sendPaymentNotification(int(payload.Amount), payload.Memo, destinationPubkey, payload.Type)
		fmt.Println("sent notification")
		msg.Ack(true)
	}
}
func sendPaymentNotification(amount int, msg, dest, invoiceType string) {
	secretKey := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(secretKey)
	_, theirPk, err := nip19.Decode(dest)
	theirHexPk := fmt.Sprintf("%v", theirPk)
	if err != nil {
		fmt.Println(err)
		return
	}
	ss, err := nip04.ComputeSharedSecret(theirHexPk, secretKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	firstWord := "Received"
	if invoiceType == "outgoing" {
		firstWord = "Sent"
	}
	encrypted, err := nip04.Encrypt(fmt.Sprintf("%s a %d sat payment. Message: %s", firstWord, amount, msg), ss)
	if err != nil {
		fmt.Println(err)
		return
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
}
