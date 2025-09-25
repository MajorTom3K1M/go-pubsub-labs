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

		fmt.Println("Move outcome:", moveOutcome)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Println("War has been initiated!", gs.GetUsername())
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		out, winner, loser := gs.HandleWar(rw)
		switch out {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Println("War lost against opponent:", winner)
			if err := publishGamelogs(ch, gs.GetUsername(), gs.NewGameLog(
				fmt.Sprintf("%s won a war against %s", winner, loser),
			)); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			fmt.Println("War won against opponent:", loser)
			if err := publishGamelogs(ch, gs.GetUsername(), gs.NewGameLog(
				fmt.Sprintf("%s won a war against %s", winner, loser),
			)); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			fmt.Println("War ended in a draw between:", winner, "and", loser)
			if err := publishGamelogs(ch, gs.GetUsername(), gs.NewGameLog(
				fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
			)); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown war outcome")
		return pubsub.NackDiscard
	}
}

func publishGamelogs(ch *amqp.Channel, username string, logs routing.GameLog) error {
	routingKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)
	if err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routingKey, logs); err != nil {
		fmt.Println("failed to publish game logs:", err)
		return err
	}
	return nil
}
