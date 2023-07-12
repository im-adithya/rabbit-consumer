package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type StoreDeadLetterHandler struct {
	db *gorm.DB
}

func NewStoreDeadLetterHandler() (result *StoreDeadLetterHandler, err error) {
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URI")))
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&DeadLetterMessage{})
	if err != nil {
		return nil, err
	}
	return &StoreDeadLetterHandler{
		db: db,
	}, nil
}

type DeadLetterMessage struct {
	gorm.Model
	Payload JSONB
}

// https://gist.github.com/yanmhlv/d00aa61082d3b8d71bed
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}
func (sdl *StoreDeadLetterHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	logrus.Infof("dead letter: %v", string(msg.Body))
	logrus.Infof("headers: %v", msg.Headers)
	payload := map[string]interface{}{}
	err := json.Unmarshal(msg.Body, &payload)
	if err != nil {
		return err
	}
	return sdl.db.Create(&DeadLetterMessage{Payload: payload}).Error
}
