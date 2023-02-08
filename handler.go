package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Handler interface {
	Handle(ctx context.Context, msg amqp.Delivery) error
	StartRabbit(ctx context.Context) (<-chan (amqp.Delivery), error)
	CloseRabbit()
}

type RabbitClient struct {
	cfg  *RabbitConfig
	ch   *amqp.Channel
	conn *amqp.Connection
}

type RabbitConfig struct {
	RabbitMQUri        string `envconfig:"RABBITMQ_URI"`
	RabbitMQQueueName  string `envconfig:"RABBITMQ_QUEUE_NAME"`
	RabbitMQExchange   string `envconfig:"RABBITMQ_EXCHANGE_NAME"`
	RabbitMQRoutingKey string `envconfig:"RABBITMQ_ROUTING_KEY"`
}

func (rabbit *RabbitClient) Close() {
	rabbit.ch.Close()
	rabbit.conn.Close()
}

func (rabbit *RabbitClient) Init() error {
	cfg := &RabbitConfig{}
	err := envconfig.Process("", cfg)
	if err != nil {
		logrus.Fatalf("Error loading environment variables: %v", err)
	}
	rabbit.cfg = cfg
	conn, err := amqp.Dial(rabbit.cfg.RabbitMQUri)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	rabbit.ch = ch
	rabbit.conn = conn
	q, err := ch.QueueDeclare(
		rabbit.cfg.RabbitMQQueueName,
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}
	return ch.QueueBind(
		q.Name,                        // queue name
		rabbit.cfg.RabbitMQRoutingKey, // routing key
		rabbit.cfg.RabbitMQExchange,   // exchange
		false,
		nil,
	)

}
