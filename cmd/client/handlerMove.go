package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	// 📌  handlerMove function  📝 🗑️
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			log.Println("acknowledge type is Ack")
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			log.Println("acknowledge type is NackReque")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			log.Println("acknowledge type is NackDiscard")
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}
