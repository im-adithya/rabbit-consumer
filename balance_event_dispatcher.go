package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type BalanceEventDispatcher struct {
	URL string
}

func (disp *BalanceEventDispatcher) Handle(ctx context.Context, msg amqp.Delivery) error {
	resp, err := http.Post(disp.URL, "application/json", bytes.NewReader(msg.Body))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Bad status code from nwc API: %d, body %s", resp.StatusCode, string(body))
	}
	return nil
}
