package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	conctString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(conctString)
	if err != nil {
		fmt.Println("unable to establish connection to server...")
	}
	defer connection.Close()

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("unable to get client username...")
		return
	}

	channel, err := connection.Channel()
	if err != nil {
		fmt.Println("unable to establish create connection channel...")
		return
	}
	err = pubsub.PublishJSON(channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		fmt.Println("unable to pusblish message")
		return
	}
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%v.%v", routing.PauseKey, userName),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Println("unable to declare and bind")
		return
	}

	//use NewGameState function to create a new game state
	gameState := gamelogic.NewGameState(userName)

// 📌  use SubscribeJson 📝 🗑️
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		"pause."+userName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState))

	// 📌  collect data from queue 📝 🗑️
	for {
		userInputWords := gamelogic.GetInput()
		if len(userInputWords) == 0 {
			continue
		}

		word := userInputWords[0]
		switch word {
		case "spawn":
			err := gameState.CommandSpawn(userInputWords)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			_, err := gameState.CommandMove(userInputWords)
			if err != nil {
				fmt.Println(err)
				continue
			}

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("unknown command")
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan //this is a technique to block the execution of the code below until a value has been passed into signalChan
	fmt.Println("exiting program...")

}
