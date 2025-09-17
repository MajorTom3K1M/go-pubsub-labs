package main

import (
	"fmt"
	"go-pubsub-labs/internal/gamelogic"
	"go-pubsub-labs/internal/pubsub"
	"go-pubsub-labs/internal/routing"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const RABMITMQ_URL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(RABMITMQ_URL)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Error during welcome:", err)
		return
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	ch, q, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient)
	if err != nil {
		fmt.Println("setup error:", err)
		return
	}
	defer ch.Close()

	fmt.Printf("Client setup complete for user %s, queue=%s\n", username, q.Name)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	fmt.Println("Shutting down Peril client...")
}
