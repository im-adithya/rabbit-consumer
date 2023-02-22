package main

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type PrintHandler struct{}

func (ph *PrintHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	logrus.WithField("msg", string(msg.Body)).Info("New event")
	return nil
}
