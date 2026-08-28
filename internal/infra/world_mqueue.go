package infra

import (
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	queue chan string
}

func NewRabbitMQ() *RabbitMQ {
	q := &RabbitMQ{
		queue: make(chan string, 1024),
	}

	go q.mainLoop()

	return q
}

func (q *RabbitMQ) mainLoop() {
	s := ""

	for {
		conn := dial()
		chnl := makeChnl(conn)

		slog.Info("rabbitmq", "state", "connected")
		for {
			for s == "" {
				select {
				case s = <-q.queue:
				}
			}

			if !publish(chnl, s) {
				break
			}
			slog.Info("rabbitmq", "state", "sent")
			s = ""
		}

		if chnl != nil {
			chnl.Close()
		}

		if conn != nil {
			conn.Close()
		}

		slog.Info("rabbitmq", "state", "relax")
		time.Sleep(time.Second * 5)
	}
}

func dial() *amqp.Connection {
	slog.Info("rabbitmq", "state", "dial")

	conn, e := amqp.Dial(
		Config.Get(
			"rabbitmq.url",
			"amqp://guest:guest@rabbitmq:5672",
		),
	)

	if e == nil {
		return conn
	}

	slog.Info("rabbitmq", "connect error", e)

	return nil
}

func makeChnl(conn *amqp.Connection) *amqp.Channel {
	if conn != nil {
		chnl, e := conn.Channel()

		if e == nil {
			return chnl
		}

		slog.Info("rabbitmq", "chan error", e)
	}

	return nil
}

func publish(chnl *amqp.Channel, s string) bool {
	e := chnl.Publish(
		"",      // Exchange
		"world", // Key
		true,    // Mandatory
		false,   // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(s),
		},
	)

	if e != nil {
		slog.Info("rabbitmq", "publish error", e)

		return false
	}

	return true
}

func (q *RabbitMQ) Put(s string) {
	q.queue <- s
}
