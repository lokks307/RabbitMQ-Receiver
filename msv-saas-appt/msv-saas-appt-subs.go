/*
Exchange type: direct
Exchange Name: msv-saas-appt-subs-logs
Routing key: msv-saas-appt-subs-key

id: c2-test
pw: c2-test-pw
*/

/*
sudo docker build -t msv-saas-appt-subs:0.0.1 .

sudo docker run -d --name msv-saas-appt-subs \
-v ~/logs/msv-saas-appt-subs-logs:/app/logs/msv-saas-appt-subs-logs \
msv-saas-appt-subs:0.0.1
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
	exchangeName := "msv-saas-appt-subs-logs"
	routingKey := "msv-saas-appt-subs-key"

	currentTime := time.Now().Format("2006-01-02")
	logFilePath := fmt.Sprintf("logs/%s/%s-%s.txt", exchangeName, exchangeName, currentTime)

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
		"msv-saas-appt-subs-queue", // name
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
				file.Close()
				currentTime = time.Now().Format("2006-01-02")

				logFilePath := fmt.Sprintf("logs/%s/%s-%s.txt", exchangeName, exchangeName, currentTime)

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