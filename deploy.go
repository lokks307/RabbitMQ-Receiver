package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func main() {
	var rabbitmqID, rabbitmqPW string
	var rabbitmqServer, rabbitmqAddr, rabbitmqPort string
	var exchangeName, queueName, routingKey string
	var logPath, logFile, logFilePath string

	subs := flag.String("subs", "empty", "name of the subscriber")
	runMode := flag.String("runmode", "test", "test or prod")
	rabbitserver := flag.String("rabbitserver","empty","domain or ip of rabbit server")
	flag.Parse()

	if *subs == "empty" || *rabbitserver == "empty" {
		log.Printf("subs or rabbitserver cant be empty")
		return
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}
	currentTime := time.Now().In(loc).Format("2006-01-02")

	rabbitmqID = fmt.Sprintf("c2-%s", *runMode)
	rabbitmqPW = fmt.Sprintf("c2-%s-pw", *runMode)

	exchangeName = fmt.Sprintf("%s-%s-exchange", *subs, *runMode)
	queueName = fmt.Sprintf("%s-%s-queue", *subs, *runMode)
	routingKey = fmt.Sprintf("%s-%s-key", *subs, *runMode)

	logPath = fmt.Sprintf("logs/%s-%s-logs/", *subs, *runMode)
	logFile = fmt.Sprintf("%s-%s-%s-logs.txt", *subs, *runMode, currentTime)
	logFilePath = fmt.Sprintf("%s/%s", logPath, logFile)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	failOnError(err, "error opening log file")
	defer file.Close()
	log.SetOutput(file)

	rabbitmqServer = *rabbitserver
	rabbitmqPort = "5672"
	rabbitmqAddr = fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqID, rabbitmqPW, rabbitmqServer, rabbitmqPort)

	log.Printf("Dialing %s...", rabbitmqAddr)
	conn, err := amqp.Dial(rabbitmqAddr)
	failOnError(err, "error connecting rabbitmq")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "error making rabbitmq channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "exchange declaration failed")

	queue, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "failed to declare a queue")

	err = ch.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	failOnError(err, "failed to bind a queue")

	log.Printf("set up success", currentTime)

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,  //auto ack
		false, // exclusive
		false, //no local
		false, //no wait
		nil,
	)
	failOnError(err, "consumption failed")

	var forever chan struct{}

	go func() {
		log.Printf("go routine started")
		for d := range msgs {
			if currentTime != time.Now().In(loc).Format("2006-01-02") {
				file.Close()
				currentTime = time.Now().In(loc).Format("2006-01-02")
				logFile = fmt.Sprintf("%s-%s-%s-logs.txt", *subs, *runMode, currentTime)
				logFilePath = fmt.Sprintf("%s/%s", logPath, logFile)

				file, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				failOnError(err, "open file rotation failed "+currentTime)

				log.SetOutput(file)
			}
			log.Printf("%s", d.Body)
		}
	}()

	<-forever
	file.Close()
}
