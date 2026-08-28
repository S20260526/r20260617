package main

import (
	"errors"
	"log/slog"
	"time"

	"infra"

	amqp "github.com/rabbitmq/amqp091-go"
)

func loopCycle() {
	slog.Info("rabbitmq", "state", "dial")

	conn, e := amqp.Dial(
		infra.Config.Get(
			"rabbitmq.url",
			"amqp://guest:guest@rabbitmq:5672",
		),
	)

	if e != nil {
		slog.Info("rabbitmq", "connect error", e)

		return
	}

	defer conn.Close()

	chnl, e := conn.Channel()

	if e != nil {
		slog.Info("rabbitmq", "chan error", e)

		return
	}

	defer chnl.Close()

	_, e = chnl.QueueDeclare(
		"world",
		true,  // Durable
		false, // Delete if unused
		false, // Exclusive
		true,  // No-wait
		nil,   // Arguments
	)

	if e != nil {
		slog.Info("rabbitmq", "queue error", e)

		return
	}

	for {
		e = chnl.Qos(
			1,     // prefetchCount
			0,     // prefetchSize
			false, // lobal
		)

		if e != nil {
			slog.Info("rabbitmq", "Qos error", e)

			break
		}

		delivery, e := chnl.Consume(
			"world",
			"",    // Consumer
			true,  // Auto-Ack
			false, // Exclusive,
			false, // No-local,
			false, // No-wait,
			nil,   // Args
		)

		for e == nil {
			slog.Info("rabbitmq", "state", "ready")

			select {
			case msg, ok := <-delivery:
				if !ok {
					e = errors.New("delivery unexpected nil")
				} else {
					slog.Info("rabbitmq", "msg", msg.Body)
				}
			}
		}

		slog.Info("rabbitmq", "consume error", e)
	}

	slog.Info("rabbitmq", "state", "relax")
	time.Sleep(time.Second * 1)
}

func main() {
	for {
		loopCycle()
	}
}
