package main

import (
	"fmt"
	"go-pubsub-labs/internal/gamelogic"
	"go-pubsub-labs/internal/pubsub"
	"go-pubsub-labs/internal/routing"

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

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}

	gameState := gamelogic.NewGameState(username)

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}
	fmt.Println("Subscribed to pause messages")

	queueName = fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	key := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		queueName,
		key,
		pubsub.Transient,
		handlerArmyMove(gameState, ch),
	)
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	queueName = routing.WarRecognitionsPrefix
	key = fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		queueName,
		key,
		pubsub.Durable,
		handlerWar(gameState),
	)
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

gameLoop:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Current Units: %+v\n", gameState.GetPlayerSnap())
		case "move":
			armyMove, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				armyMove,
			)
			if err != nil {
				fmt.Println("publish error:", err)
				continue
			}
			fmt.Println("Move command published.")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break gameLoop
		default:
			fmt.Println("Unknown command:", input[0])
			gamelogic.PrintClientHelp()
			continue
		}
	}
}
