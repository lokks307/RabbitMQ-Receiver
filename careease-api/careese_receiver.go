/*
This is a receiver of careease-api

Run on Docker
5672:5672
15672:15672
port forwarded from host to docker-env


Username: c2-test
Userpw: c2-test-pw

Exchange name: careease-logs
Exchange type: direct

Routing key: careease-api-key

Dashboard address: ec2-43-203-234-165.ap-northeast-2.compute.amazonaws.com:15672
*/

/*
docker build -t careease_receiver:0.0.1 .

	docker run -d --name careease_receiver \
	  -v ~/logs/:/app/logs/ \
	  careease_receiver:0.0.1
*/
package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Specify the log file path
	logFilePath := "logs/careease-log.txt"

	// Open or create the log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	// Set output of logs to file
	log.SetOutput(file)

	conn, err := amqp.Dial("amqp://c2-test:c2-test-pw@ec2-43-203-234-165.ap-northeast-2.compute.amazonaws.com:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "careease-logs"
	routingKey := "careease-api-key"
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"careease-api-queue", // name
		false,                // durable
		false,                // delete when unused
		true,                 // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	failOnError(err, "Failed to declare a queue")

	log.Printf("conn, exchange, queue declaration success!")
	log.Printf("Binding queue %s to exchange %s with routing key %s",
		q.Name, exchangeName, routingKey)
	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil)

	failOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
