// This file contains the queue implementation
package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func (a *App) connectRemoteQueue() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func (a *App) connectQueueChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	return ch
}

func (a *App) declareQueue(ch *amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare(
		"domain-queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	return q
}

func (a *App) publishMessage(q *amqp.Queue, ch *amqp.Channel, body []byte) {
	err := ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
// This function is expected to be launched in its own go routine (thread)
// It will monitor the queue and process queue entries as they arrive
func (a *App) queueProcessor() {
	log.Print("Starting queue processor")

	conn := a.connectRemoteQueue()
	defer conn.Close()

	ch := a.connectQueueChannel(conn)
	defer ch.Close()

	q := a.declareQueue(ch)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			entry := queueEntry{}
			json.Unmarshal(d.Body, &entry)
			a.processPut(entry)
		}
	}()

	<-forever
}

