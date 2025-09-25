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

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nack_requeue"
	NackDiscard AckType = "nack_discard"
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
	dlx := "peril_dlx"
	table := amqp.Table{
		"x-dead-letter-exchange": dlx,
	}
	var queue amqp.Queue
	if queueType == Durable {
		queue, err = ch.QueueDeclare(queueName, true, false, false, false, table)
	} else {
		queue, err = ch.QueueDeclare(queueName, false, true, true, false, table)
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType,
	handler func(T) AckType,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		fmt.Println("setup error:", err)
		return err
	}

	deliveryChan, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println("Failed to register a consumer:", err)
		return err
	}

	go func() {
		defer ch.Close()
		for d := range deliveryChan {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				fmt.Println("Failed to unmarshal JSON:", err)
				continue
			}

			switch handler(msg) {
			case Ack:
				d.Ack(false)
			case NackDiscard:
				d.Nack(false, false)
			case NackRequeue:
				d.Nack(false, true)
			}
			if err != nil {
				fmt.Println("Failed to acknowledge message:", err)
				continue
			}
		}
	}()

	return nil
}
