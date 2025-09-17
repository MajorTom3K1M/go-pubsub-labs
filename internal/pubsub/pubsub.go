package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const (
	Durable   simpleQueueType = "durable"
	Transient simpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	ctx := context.Background()
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	}); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return nil, amqp.Queue{}, err
	}

	var queue amqp.Queue
	if queueType == Durable {
		queue, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	} else {
		queue, err = ch.QueueDeclare(queueName, false, true, true, false, nil)
	}
	if err != nil {
		fmt.Println("Failed to declare a queue:", err)
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		fmt.Println("Failed to bind a queue:", err)
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
