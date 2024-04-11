/*
Exchange type: direct
Exchange Name: careplanner-admin-api-logs
Routing key: careplanner-admin-api-key

id: c2-test
pw: c2-test-pw
*/

/*
sudo docker build -t careplanner-admin-api-receiver:0.0.1 .

sudo docker run -d --name careplanner-admin-api-receiver \
-v ~/logs/careplanner-admin-api-logs:/apps/logs/careplanner-admin-api-logs \
careplanner-admin-api-receiver:0.0.1
*/
package main
import (
	"log"
	"os"
	"fmt"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func main(){
	currentTime := time.Now()
	logFilePath := fmt.Sprintf("logs/careplanner-admin-api-logs/careplanner-admin-log-%s.txt", currentTime.Format("2006-01-02"))

	file, err := os.OpenFile(logFilePath, os.O_CREATE | os.O_APPEND| os.O_WRONLY, 0666)
	failOnError(err, "error opening log file")
	defer file.Close()
	log.SetOutput(file)

	conn, err := amqp.Dial("amqp://c2-test:c2-test-pw@ec2-43-203-234-165.ap-northeast-2.compute.amazonaws.com:5672/")
	failOnError(err, "error connecting rabbitmq")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "error making rabbitmq channel")
	defer ch.Close()

	exchangeName := "careplanner-admin-api-logs"
	routingKey := "careplanner-admin-api-key"
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
		"careplanner-admin-api-queue", // name
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
		for d:= range msgs{
			log.Printf("%s", d.Body)
		}
	}()
	
	<-forever
}