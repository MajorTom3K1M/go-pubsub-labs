package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
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
			msg, err := decodeGob[T](d.Body)
			if err != nil {
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

func decodeGob[T any](data []byte) (T, error) {
	var msg T
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&msg); err != nil {
		return msg, err
	}
	return msg, nil
}
