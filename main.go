package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

const (
	NostrificationHandler = "nostrification_handler"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	handlerType := os.Getenv("HANDLER_TYPE")
	var handler Handler
	switch handlerType {
	case NostrificationHandler:
		handler = NewNostrificationSender()
	default:
		logrus.Fatalf("Unknown handler type: %s", handlerType)
	}
	msgs, err := handler.StartRabbit(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Context canceled, exiting gracefully.")
		case msg := <-msgs:
			handler.Handle(ctx, msg)
		}
	}
}
