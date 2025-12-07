package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	//use NewGameState function to create a new game state
	gameState := gamelogic.NewGameState(userName)

	// 📌  use SubscribeJson 📝 🗑️
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		"pause."+userName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"army_moves."+userName,
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gameState),
	)

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
			move, err := gameState.CommandMove(userInputWords)
			if err != nil {
				log.Printf("unable to pusblish message. error: %v\n", err)
				continue
			}

			err = pubsub.PublishJSON(channel,
				routing.ExchangePerilTopic,
				"army_moves."+userName,
				move,
			)
			if err != nil {
				log.Printf("unable to pusblish message. error:%v\n", err)
				continue
			}
			log.Println("message published successfully")

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
