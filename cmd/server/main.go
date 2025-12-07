package main

import (
	"fmt"
	"strings"

	gamelogic "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

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

	fmt.Println("exiting program...")

	gamelogic.PrintServerHelp()
	channel, err := connection.Channel()
	if err != nil {
		fmt.Println("unable to establish create connection channel...")
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable)

	if err != nil {
		fmt.Println("unable to declare and bind Queue...")
	}

	for {
		userInputs := gamelogic.GetInput()
		if len(userInputs) == 0 {
			continue
		}

		for _, val := range userInputs {
			firstWord := strings.Split(val, " ")[0]

			switch firstWord {
			case "pause":
				fmt.Println("sending pause message...")
				err = pubsub.PublishJSON(channel,
					routing.ExchangePerilDirect,
					routing.PauseKey,
					routing.PlayingState{
						IsPaused: true,
					},
				)
				if err != nil {
					fmt.Println("unable to pusblish pause message")
				}
			case "resume":
				fmt.Println("sending resume message...")
				err = pubsub.PublishJSON(channel,
					routing.ExchangePerilDirect,
					routing.PauseKey,
					routing.PlayingState{
						IsPaused: false,
					},
				)
				if err != nil {
					fmt.Println("unable to pusblish resume message")
				}
			case "quit":
				fmt.Println("exiting sending message...")
				return
			default:
				fmt.Printf("invalid command. could not process command:%v\n", firstWord)

			}
		}

	}
}
