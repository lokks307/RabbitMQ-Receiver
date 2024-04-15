package main

/*
sudo docker build -t careplanner-mobile-api-receiver:0.0.1 .

sudo docker run -d --name careplanner-mobile-api \
-v ~/logs/careplanner-mobile-api-logs:/app/logs/careplanner-mobile-api-logs \
careplanner-mobile-api-receiver:0.0.1
*/

import(
	"fmt"
	"os"
	"log"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)
func failOnError(err error, msg string){
	if err != nil{
		log.Panicf(msg, err)
	}
}

func main(){
	currentTime := time.Now().Format("2006-01-02")

	logFilePath := fmt.Sprintf("logs/careplanner-mobile-api-logs/careplanner-mobile-logs-%s.txt",currentTime)
	file, err := os.OpenFile(logFilePath, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0666 )
	failOnError(err, "open file failed " + currentTime)
	log.SetOutput(file)

	conn, err := amqp.Dial("amqp://c2-test:c2-test-pw@ec2-43-203-234-165.ap-northeast-2.compute.amazonaws.com:5672/")
	failOnError(err, "rabbitmq connection failed")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "opening channel failed")
	defer ch.Close()


	exchangeName := "careplanner-mobile-api-logs"
	routingKey := "careplanner-mobile-api-key"
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
		"careplanner-mobile-api-queue", // name
		false,                // durable
		false,                // delete when unused
		true,                 // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	failOnError(err, "failed to declare a queue")

	err =ch.QueueBind(
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
		true, //auto ack
		false, // exclusive
		false, //no local
		false, //no wait
		nil,
	)
	failOnError(err, "consumption failed")

	var forever chan struct{}

	go func(){
		log.Printf("go routine started")
		for d:= range msgs{
			if currentTime != time.Now().Format("2006-01-02"){
				log.Printf("am i here?")
				file.Close()
				currentTime = time.Now().Format("2006-01-02")

				logFilePath := fmt.Sprintf("logs/careplanner-mobile-api-logs/careplanner-mobile-logs-%s.txt",currentTime)
				file, err = os.OpenFile(logFilePath, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0666 )
				failOnError(err, "open file rotation failed " + currentTime)

				log.SetOutput(file)
			} 
			log.Printf("%s", d.Body)
		}
	}()
	<-forever
	file.Close()

}