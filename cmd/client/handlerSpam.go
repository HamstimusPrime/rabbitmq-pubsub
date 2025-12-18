package main

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerSpam(userInput []string, username string, channel *amqp.Channel) error {
	if len(userInput) != 2 {
		log.Println("please provide a value for the number of spams")
		return errors.New("invalid number of inputs")
	}
	numOfSpams, err := strconv.Atoi(userInput[1])
	if err != nil {
		log.Println("please provide a valid integer number")
		return err
	}
	for i := 0; i < numOfSpams; i++ {
		maliciousString := gamelogic.GetMaliciousLog()
		maliciousGameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     maliciousString,
			Username:    username,
		}

		//publish Malicious gamelog as GO binary using publishGob
		err := pubsub.PublishGob(
			channel,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+username,
			maliciousGameLog,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
