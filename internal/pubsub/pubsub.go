package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

type AckType int

const (
	Ack         AckType = iota // automatically becomes 0
	NackRequeue                // automatically becomes 1
	NackDiscard                // automatically becomes 2
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes,
		},
	)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

// 📌  declareAndBind functionality 📝 🗑️
func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("unable to get client username...")
		return nil, amqp.Queue{}, err
	}

	// 📌  amqpTable config 📝 🗑️
	table := amqp.Table{}
	table["x-dead-letter-exchange"] = routing.ExchangePerilDeadLetter

	queue, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false, // noWait
		table,
	)
	if err != nil {
		log.Printf("unable to declare Queue ... error: %v\n", err)
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Println("unable to bind Queue ...")
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

// 📌  subscribeJson functionality 📝 🗑️
func SubscribeJSON[T any](

	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		fmt.Println("unable to bind Queue ...")
		return err
	}

	// 📌  consuming data from channel 📝 🗑️

	delivery, err := channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("unable to unmarshall delivery")
		fmt.Printf("from consume. err: %v\n", err)
		return err
	}
	go acknowledgeDelivery(delivery, handler)
	return nil
}

// 📌  acknowledgeDelivery 📝 🗑️
func acknowledgeDelivery[T any](delivery <-chan amqp.Delivery, handler func(T) AckType) {
	for message := range delivery {
		var payload T
		err := json.Unmarshal(message.Body, &payload)
		if err != nil {
			fmt.Println("unable to unmarshall delivery")
			fmt.Printf("from ack, err: %v\n", err)
			continue
		}

		ackType := handler(payload)
		switch ackType {
		case Ack:
			message.Ack(false)
			fmt.Println("Ack")
		case NackDiscard:
			message.Nack(false, false)
			fmt.Println("NackDiscard")
		case NackRequeue:
			message.Nack(false, true)
			fmt.Println("NackRequeue")
		}
	}
}
