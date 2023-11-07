package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	svixnotifications "github.com/getAlby/rabbit-consumer/internal/svix"
	"github.com/joho/godotenv"

	"github.com/wagslane/go-rabbitmq"
)

func main() {
	godotenv.Load(".env")
	conn, err := rabbitmq.NewConn(
		"amqp://guest:guest@localhost",
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	svix := svixnotifications.NewSvixNotifications(ctx)

	var wg sync.WaitGroup
  wg.Add(1)

	consumer, err := rabbitmq.NewConsumer(
		conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			defer wg.Done()
			return svix.SendNotification(d)
		},
		"my_queue",
		rabbitmq.WithConsumerOptionsRoutingKey("my_routing_key"),
		rabbitmq.WithConsumerOptionsExchangeName("my_exchange"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	consumer.Close()
}
