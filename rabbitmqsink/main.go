package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func loopCycle(url string, w io.Writer) {
	slog.Info("rabbitmq", "state", "dial")

	conn, e := amqp.Dial(url)

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

	encoder := json.NewEncoder(w)

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

		if e == nil {
			slog.Info("rabbitmq", "state", "ready")
		}

		for e == nil {

			select {
			case msg, ok := <-delivery:
				if !ok {
					e = errors.New("delivery unexpected nil")
				} else {
					var buf [1024 * 1024]byte

					rand.Read(buf[:])

					e := encoder.Encode(
						struct {
							Stamp string `json:"stamp"`
							Blob  string `json:"blob"`
						}{
							Stamp: string(msg.Body),
							Blob:  base64.StdEncoding.EncodeToString(buf[:]),
						},
					)

					if e != nil {
						slog.Info("rabbitmq", "sink error", e)
					}
				}
			}
		}

		slog.Info("rabbitmq", "consume error", e)
	}

	slog.Info("rabbitmq", "state", "relax")
}

func main() {
	var w io.Writer

	f, e := os.Create("/sink.txt")

	if e == nil {
		w = f
	} else {
		slog.Info("sinkfile", "error", e)
	}

	for {
		for _, url := range []string{
			"amqp://guest:guest@rmq1:5672",
			"amqp://guest:guest@rmq2:5672",
			"amqp://guest:guest@rmq3:5672",
		} {
			loopCycle(url, w)

			time.Sleep(time.Second * 1)
		}
	}
}
