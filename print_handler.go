package main

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PrintHandler struct{}

func (ph *PrintHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	payload := map[string]interface{}{}
	err := json.Unmarshal(msg.Body, &payload)
	if err != nil {
		return err
	}
	pretty, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(pretty))
	return nil
}
