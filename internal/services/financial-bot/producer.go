package financialbot

import (
	"github.com/streadway/amqp"
	"log"
)

const (
	connectionURL = "amqp://guest:guest@localhost:5672/"
	queueName     = "financial-info"
)

func init() {
	setupProducer()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func setupProducer() {
	conn, err := amqp.Dial(connectionURL)
	failOnError(err, "failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "failed to open a channel")

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to declare a queue")

	financialBot.conn = conn
	financialBot.ch = ch
	financialBot.queue = q
}
