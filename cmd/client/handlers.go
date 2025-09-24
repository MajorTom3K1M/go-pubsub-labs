package main

import (
	"fmt"
	"go-pubsub-labs/internal/gamelogic"
	"go-pubsub-labs/internal/pubsub"
	"go-pubsub-labs/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerArmyMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutcomeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			username := gs.GetUsername()
			routingKey := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, username)
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routingKey, gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}); err != nil {
				fmt.Println("failed to publish war recognition:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		out, winner, loser := gs.HandleWar(rw)
		switch out {
		case gamelogic.WarOutcomeNotInvolved:
			fmt.Println("Received war recognition but not involved.")
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			fmt.Println("War cannot be resolved, no units available.")
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Println("War lost against opponent:", loser)
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			fmt.Println("War won against opponent:", winner)
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			fmt.Println("War ended in a draw between:", winner, "and", loser)
			return pubsub.Ack
		default:
			fmt.Println("Unknown war outcome.")
			return pubsub.NackDiscard
		}
	}
}
