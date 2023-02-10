package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/sirupsen/logrus"
)

const (
	NostrificationHandler = "nostrification_handler"
	SlackHandler          = "slack_handler"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Warn("Failed to load .env file")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	handlerType := os.Getenv("HANDLER_TYPE")
	var handler Handler
	switch handlerType {
	case NostrificationHandler:
		ns, err := NewNostrificationSender()
		if err != nil {
			logrus.Fatal(err)
		}
		handler = ns
	case SlackHandler:
		sh, err := NewSlackSender()
		if err != nil {
			logrus.Fatal(err)
		}
		handler = sh
	default:
		logrus.Fatalf("Unknown handler type: %s", handlerType)
	}
	//init rabbit
	rabbit := &RabbitClient{}
	err = rabbit.Init()
	if err != nil {
		logrus.Fatal(err)
	}
	defer rabbit.Close()
	msgs, err := rabbit.ch.Consume(
		rabbit.cfg.RabbitMQQueueName, // queue
		"",                           // consumer
		false,                        // auto ack
		false,                        // exclusive
		false,                        // no local
		false,                        // no wait
		nil,                          // args
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Starting consumer...")
	logrus.Info(lnrpc.ChannelEventUpdate_OPEN_CHANNEL.String())
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Context canceled, exiting gracefully.")
			return
		case msg := <-msgs:
			err = handler.Handle(ctx, msg)
			if err != nil {
				logrus.Error(err)
				continue
			}
			msg.Ack(true)
		}
	}
}
