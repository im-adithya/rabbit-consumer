package firebasenotifications

import (
	"context"
	"encoding/json"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type firebaseNotifications struct {
	client *messaging.Client
	ctx    context.Context
}

func NewFirebaseNotifications(ctx context.Context) *firebaseNotifications {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Messaging(ctx)
	return &firebaseNotifications{
		client: client,
		ctx:    ctx,
	}
}

type invoice struct {
	//invoice payload that needs to be consumed
	//needs to include user fcm token
	//ideally also includes nostr fields (payer name, profile pic)
	Token string
}

func (fn *firebaseNotifications) sendNotification(d rabbitmq.Delivery) rabbitmq.Action {
	invoice := invoice{}
	err := json.Unmarshal(d.Body, &invoice)
	if err != nil {
		logrus.WithField("body", string(d.Body)).
			WithError(err).
			Error("error unmarshalling body")
		return rabbitmq.NackDiscard
	}
	id, err := fn.client.Send(fn.ctx, &messaging.Message{
		//todo
		Data:         map[string]string{},
		Notification: &messaging.Notification{},
		Android:      &messaging.AndroidConfig{},
		Webpush:      &messaging.WebpushConfig{},
		Token:        invoice.Token,
	})
	if err != nil {
		logrus.WithField("body", string(d.Body)).
			WithError(err).
			Error("error unmarshalling body")
			//requeue?
		return rabbitmq.NackDiscard
	}
	logrus.WithField("firebase id", id).Info("succesfully send notification")
	return rabbitmq.Ack
}
