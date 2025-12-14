package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	// 📌  handlerMove function  📝 🗑️
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			log.Println("acknowledge type is Ack")
			return pubsub.Ack
		case gamelogic.MoveOutComeSafe:
			log.Println("acknowledge type is NackReque")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			log.Printf("Move player: %+v", move.Player)
			log.Printf("Defender: %+v", gs.GetPlayerSnap())
			player := gs.GetPlayerSnap()
			warResponse := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: player,
			}
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+move.Player.Username,
				warResponse,
			)

			if err != nil {
				log.Printf("unable to publish war response. err: %v", err)
				return pubsub.NackRequeue
			}
			log.Println("acknowledge type is NackDiscard")
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}
