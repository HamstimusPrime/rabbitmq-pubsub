package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	//encode val T to gob bytes
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)

	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes(),
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
	delivery, err := subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		fmt.Println("subscribe process failed ...")
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
			return
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

func PublishGameLog(gameLog routing.GameLog, channel *amqp.Channel) error {
	/*this function calls publishGob in order to serialize and
	publish the gameLog argument to a queue in an exchange with
	the following values
	exchange: Topic
	routing key: GamelogSlug.Username
	username: Name of player that initiated the war
	GamelogSlug if a constant in routing package*/

	err := PublishGob(
		channel,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+gameLog.Username,
		gameLog,
	)

	if err != nil {
		return err
	}
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	/*
		subscribe Gob would consume(have access to) data sent  to
		a certain channel. The data that it expects would come in as binary (Gob). That binary
		would need to be converted (decoded) into the generic type T that is passed into subscribeGob.
		It then returns an error if the subscribe process fails.
	*/
	delivery, err := subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		log.Println("subscribe process failed ...")
		return err
	}

	//decode delivery
	//delivery has to be converted into generic type and then called with handler
	for message := range delivery {
		//message would be a binary file. we would need to decode it into a generic and pass it into handler
		var payload T
		gameLog, err := decode(message.Body, payload)
		if err != nil {
			log.Println("unable to decode bytes")
			return err
		}
		//write to gamelogic.WriteLog
		ackType := handler(gameLog)
		if ackType == Ack {
			message.Ack(true)
		}

	}
	return nil

}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (<-chan amqp.Delivery, error) {
	channel, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		fmt.Println("unable to bind Queue ...")
		return nil, err
	}

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
		return nil, err
	}

	return delivery, nil
}

func decode[T any](data []byte, gl T) (T, error) {

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&gl)
	if err != nil {
		return gl, err
	}
	return gl, nil
}
