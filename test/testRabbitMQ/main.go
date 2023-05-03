package main

import (
	"log"

	"github.com/streadway/amqp"
)

func main() {
	// 连接RabbitMQ服务器
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()

	// 创建一个通道
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer ch.Close()

	// 声明一个交换机
	err = ch.ExchangeDeclare("example_exchange", "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}

	// 声明一个队列
	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	// 绑定队列到交换机上
	err = ch.QueueBind(q.Name, "", "example_exchange", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind a queue to the exchange: %s", err)
	}

	// 订阅消息
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume messages: %s", err)
	}

	// 开始处理消息
	for msg := range msgs {
		log.Printf("Received a message: %s", msg.Body)
	}
}
