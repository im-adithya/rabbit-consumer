package main

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(ctx context.Context, msg amqp.Delivery) error
	StartRabbit(ctx context.Context) (<-chan (amqp.Delivery), error)
}
