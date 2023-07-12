package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const (
	NostrificationHandler       = "nostrification_handler"
	SlackChannelEventHandler    = "slack_channel_handler"
	SlackRawInvoiceEventHandler = "slack_raw_invoice_handler"
	BalanceEventHandler         = "balance_event_handler"
	PrintEventHandler           = "print_handler"
	DeadLetterHandler           = "dead_letter_handler"
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
	case SlackRawInvoiceEventHandler:
		sh, err := NewInvoiceEventHandler()
		if err != nil {
			logrus.Fatal(err)
		}
		handler = sh
	case DeadLetterHandler:
		sdlh, err := NewStoreDeadLetterHandler()
		if err != nil {
			logrus.Fatal(err)
		}
		handler = sdlh
	case PrintEventHandler:
		handler = &PrintHandler{}
	case BalanceEventHandler:
		handler = &BalanceEventDispatcher{
			URL: os.Getenv("NWC_BALANCE_EVENT_URL"),
		}
	default:
		logrus.Fatalf("Unknown handler type: %s", handlerType)
	}
	//init rabbit
	rabbit := &RabbitClient{}
	err = rabbit.Init(handlerType)
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
	logrus.Infof("Started listening to exchange %s, routing key %s, handler %s",
		rabbit.cfg.RabbitMQExchange, rabbit.cfg.RabbitMQRoutingKey, handlerType)
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Context canceled, exiting gracefully.")
			return
		case msg, ok := <-msgs:
			if !ok {
				//channel closed, connection to rabbit lost
				logrus.Fatal("Disconnected from RabbitMQ")
			}
			err = handler.Handle(ctx, msg)
			if err != nil {
				logrus.WithField("payload", string(msg.Body)).Error(err)
				//don't requeue here
				//otherwise we might get into an infinite loop
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}
}
