package svixnotifications

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	svix "github.com/svix/svix-webhooks/go"
	"github.com/wagslane/go-rabbitmq"
)

type svixNotifications struct {
	client *svix.Svix
	ctx    context.Context
}

func NewSvixNotifications(ctx context.Context) *svixNotifications {
	svixClient := svix.New(os.Getenv("SVIX_API_KEY"), nil)
	return &svixNotifications{
		client: svixClient,
		ctx:    ctx,
	}
}

type invoice struct {
	PaymentHash     string
	PaymentRequest  string
	Description     string
	DescriptionHash string
	PaymentPreimage string
	Destination     string
	Amount          int64
	Fee             int64
	Status          string
	Type            string
	ErrorMessage    string
	SettledAt       time.Time
	ExpiresAt       time.Time
	IsPaid          bool
	Keysend         bool
	CustomRecords   map[uint64][]byte
}

func (fn *svixNotifications) SendNotification(d rabbitmq.Delivery) rabbitmq.Action {
	invoice := invoice{}
	err := json.Unmarshal(d.Body, &invoice)
	if err != nil {
		logrus.WithField("body", string(d.Body)).
			WithError(err).
			Error("error unmarshalling body")
		return rabbitmq.NackDiscard
	}

	msg, err := fn.client.Message.Create(fn.ctx, os.Getenv("SVIX_APP_ID"), &svix.MessageIn{
			EventType: "invoice.incoming.settled",
			// Needs changes
			// EventId:   svix.String("evt_Wqb1k73rXprtTm7Qdlr38G"),
			Payload: map[string]interface{}{
				"id":               invoice.PaymentHash,
				"payment_request":  invoice.PaymentRequest,
				"description":      invoice.Description,
				"description_hash": invoice.DescriptionHash,
				"preimage":         invoice.PaymentPreimage,
				"destination":      invoice.Destination,
				"amount":           invoice.Amount,
				"fee":              invoice.Fee,
				"status":           invoice.Status,
				"type":             invoice.Type,
				"error_message":    invoice.ErrorMessage,
				"settled_at":       invoice.SettledAt,
				"expires_at":       invoice.ExpiresAt,
				"is_paid":          invoice.IsPaid,
				"keysend":          invoice.Keysend,
				"custom_records":   invoice.CustomRecords,
		},
	})

	if err != nil {
		logrus.WithField("message", msg).
			WithError(err).
			Error("error sending to svix")
			//requeue?
		return rabbitmq.NackDiscard
	}
	logrus.WithField("svix id", msg.Id).Info("succesfully sent notification")
	return rabbitmq.Ack
}
