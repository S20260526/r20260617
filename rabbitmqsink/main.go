package main

import (
	"errors"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	for {
		slog.Info("sink", "state", "dial")

		conn, e := amqp.Dial("amqp://guest:guest@rabbitmq:5672")

		if e != nil {
			slog.Info("sink", "connect error", e)

			continue
		}

		chnl, e := conn.Channel()

		if e != nil {
			slog.Info("sink", "chan error", e)

			conn.Close()

			continue
		}

		for {
			e = chnl.Qos(
				1,     // prefetchCount
				0,     // prefetchSize
				false, // lobal
			)

			if e != nil {
				slog.Info("sink", "Qos error", e)

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
				slog.Info("sink", "state", "ready")

				select {
				case msg, ok := <-delivery:
					if !ok {
						e = errors.New("delivery unexpected nil")
					} else {
						slog.Info("sink", "msg", msg.Body)
					}
				}
			}

			slog.Info("sink", "consume error", e)
		}

		chnl.Close()
		conn.Close()

		slog.Info("sink", "state", "relax")
		time.Sleep(time.Second * 1)
	}
}
