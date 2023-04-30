package main

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"my-queue", // queue name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatalf("failed to register a consumer: %s", err)
	}

	// forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// Handle the message here
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	// <-forever
	time.Sleep(1000 * time.Second)
}
