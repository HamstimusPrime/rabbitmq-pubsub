package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	// 📌  handlerPause function defer 📝 🗑️
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)

	}

}
