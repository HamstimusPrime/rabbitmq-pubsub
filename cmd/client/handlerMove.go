package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	// 📌  handlerMove function  📝 🗑️
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
