package financialconsumer

import (
	"github.com/streadway/amqp"
	"log"
)

const (
	connectionURL = "amqp://guest:guest@localhost:5672/"
	queueName     = "financial-info"
)

type FinancialConsumer struct {
	Conn     *amqp.Connection
	Ch       *amqp.Channel
	Queue    amqp.Queue
	Messages <-chan amqp.Delivery
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func New() *FinancialConsumer {
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

	messagesCh, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to register a consumer")

	return &FinancialConsumer{
		Conn:     conn,
		Ch:       ch,
		Queue:    q,
		Messages: messagesCh,
	}
}
