package main

import (
	c "context"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	imq "internal/msgqueue"
	"log"

	"os"
	"time"
)

type RMQQueue struct {
	name string
	conn *amqp.Connection
	chnl *amqp.Channel
}

type RMQPush struct {
	RMQQueue
}

type RMQPull struct {
	RMQQueue
	dlvr <-chan amqp.Delivery
}

func NewRMQPush(name string) *RMQPush {
	p := &RMQPush{}
	p.name = name

	return p
}

func NewRMQPull(name string) *RMQPull {
	p := &RMQPull{}
	p.name = name

	return p
}

func (q *RMQQueue) Connect(url string) error {
	conn, err := amqp.Dial(url)

	q.conn = conn

	return err
}

func (q *RMQPush) OpenChannel() error {
	chnl, err := q.conn.Channel()

	q.chnl = chnl

	return err
}

func (q *RMQPull) OpenChannel() error {
	chnl, err := q.conn.Channel()

	if err != nil {
		return err
	}

	err = chnl.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	)

	if err == nil {
		dlvr, err := chnl.Consume(
			q.name,
			"",    // Consumer
			true,  // Auto-Ack
			false, // Exclusive,
			false, // No-local,
			false, // No-wait,
			nil,   // Args
		)

		if err == nil {
			q.chnl = chnl
			q.dlvr = dlvr

			return nil
		}
	}

	chnl.Close()

	return err
}

func (q *RMQPush) CloseChannel() {
	q.chnl.Close()
}

func (q *RMQPull) CloseChannel() {
	q.dlvr = nil

	q.chnl.Close()
}

func (q *RMQQueue) Disconnect() {
	q.conn.Close()
}

func (q *RMQQueue) Publish(ctx c.Context, msg []byte) error {
	return q.chnl.PublishWithContext(
		ctx,
		"",
		q.name, // key
		true,   // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "binary/octets",
			Body:        msg,
		},
	)
}

func (q *RMQQueue) Consume(_ c.Context) ([]byte, error) {
	return nil, errors.New("penguins do fly, but only by mistake")
}

func (q *RMQPull) Consume(ctx c.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case dlvr, ok := <-q.dlvr:
		if !ok {
			return nil, errors.New("unexpected void delivery")
		}

		return dlvr.Body, nil
	}
}

func usage() {
	log.Println("usage: rabbitmq URL [push|pull]")
	os.Exit(1)
}

func goPush(url string) {
	p := imq.NewPusher(url, NewRMQPush("world"))

	ctx := c.Background()

	var i uint8 = 0

	for {
		log.Println("PUSH", i)

		p.Charge([]byte{i})

		for err := p.Push(ctx); err != nil; err = p.Push(ctx) {
			log.Println("ERR:", err)

			time.Sleep(time.Second)
		}

		i++

		time.Sleep(time.Second)
	}
}

func goPull(url string) {
	p := imq.NewPuller([]string{url}, NewRMQPull("world"))

	ctx := c.Background()

	for {
		d, err := p.Pull(ctx)

		if err != nil {
			log.Println("ERR:", err)
		} else {
			log.Println("PULL:", d)
		}

		time.Sleep(time.Second)
	}
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	url := os.Args[1]
	mode := os.Args[2]

	switch mode {
	default:
		usage()
	case "push":
		goPush(url)
	case "pull":
		goPull(url)
	}
}
