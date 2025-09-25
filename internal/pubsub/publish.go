package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := encodeGob(val)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        data,
	}); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func encodeGob[T any](val T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
