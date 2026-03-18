package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/events"
	"p2p-back-end/modules/entities/models"
)

type eventProducer struct {
	ch *amqp.Channel
}

func NewEventProducer(ch *amqp.Channel) models.EvenProducer {

	err := ch.ExchangeDeclare(
		"authen_event_topic", // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		logs.Errorf("Failed to declare exchange: %v", err)
	}
	return &eventProducer{ch}
}

func (obj *eventProducer) Producer(event events.Event) error {
	routingKey := event.String()

	value, err := json.Marshal(event)
	if err != nil {
		logs.Error(err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = obj.ch.PublishWithContext(ctx,
		"authen_event_topic",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        value,
		},
	)

	if err != nil {
		logs.Error(err)
		return err
	}

	logs.Info(fmt.Sprintf("sent to RabbitMQ Exchange: profile_events_topic with Key: %v", routingKey))
	return nil
}
