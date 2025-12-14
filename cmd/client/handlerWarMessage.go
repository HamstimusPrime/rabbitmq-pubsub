package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerWarMessage(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(war)

		gameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     "",
			Username:    gs.GetUsername(),
		}

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			gameLog.Message = winner + "won a war against" + loser
			err := pubsub.PublishGameLog(gameLog, channel)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			gameLog.Message = "A war between" + winner + "and" + loser + "resulted in a draw"
			err := pubsub.PublishGameLog(gameLog, channel)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			gameLog.Message = winner + "won a war against" + loser
			err := pubsub.PublishGameLog(gameLog, channel)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			log.Printf("error determining game logic.\n")
			return pubsub.NackDiscard
		}
	}
}
