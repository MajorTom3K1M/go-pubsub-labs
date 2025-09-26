package main

import (
	"fmt"
	"go-pubsub-labs/internal/gamelogic"
	"go-pubsub-labs/internal/pubsub"
	"go-pubsub-labs/internal/routing"
)

func handlerLog() func(routing.GameLog) pubsub.AckType {
	return func(gamelog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		gamelogic.WriteLog(gamelog)
		return pubsub.Ack
	}
}
