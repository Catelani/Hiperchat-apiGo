package libs

import (
	"log"

	"github.com/streadway/amqp"
	"github.com/Mateus-pilo/go-whats-opt/hlp"
)

var connection *session

type session struct {
	*amqp.Connection
	*amqp.Channel
}

func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}


func ConnectionMqp() (session) {

	if connection != nil {
		return *connection
	}

	url := hlp.Config.GetString("AMQP_URL")
	conn, err := amqp.Dial(url)

	if err != nil {
		panic("could not establish connection with RabbitMQ:" + err.Error())
	}

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")


	if err != nil {
			panic("could not open RabbitMQ channel:" + err.Error())
	}

	err = ch.ExchangeDeclare("msgSend", "topic", true, false, false, false, nil)

	if err != nil {
			panic(err)
	}


	if err != nil {
		panic("error publishing a message to the queue:" + err.Error())
	}

	// We create a queue named Test
	_, err = ch.QueueDeclare("msgSend", true, false, false, false, nil)
	_, err = ch.QueueDeclare("msgReceive", true, false, false, false, nil)


	if err != nil {
		panic("error declaring the queue: " + err.Error())
	}

	connection = &session{conn,ch}
	return *connection;
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
  }
}
