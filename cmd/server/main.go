package main

import (
	"fmt"
	"go-pubsub-labs/internal/pubsub"
	"go-pubsub-labs/internal/routing"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const RABMITMQ_URL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(RABMITMQ_URL)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	defer ch.Close()

	fmt.Println("Connected to RabbitMQ")

	pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	fmt.Println("Shutting down Peril server...")
}
