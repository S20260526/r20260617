package infra

import (
	c "context"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"internal/msgqueue"
	"log/slog"
)

type RMQIncoming struct {
	dlvr *amqp.Delivery
}

func (r RMQIncoming) GetData() []byte {
	return r.dlvr.Body
}

func (r RMQIncoming) Acknowledge() {
	if err := r.dlvr.Ack(true); err != nil {
		slog.Warn(
			"infra",
			"where", "RabbitMQ",
			"when", "Ack",
			"what", err,
		)
	}
}

func (r RMQIncoming) Reject() {
	if err := r.dlvr.Nack(false, true); err != nil {
		slog.Warn(
			"infra",
			"where", "RabbitMQ",
			"when", "Nack",
			"what", err,
		)
	}
}

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

func (q *RMQQueue) Consume(_ c.Context) (msgqueue.Incoming, error) {
	return nil, errors.New("penguins do fly, but only by mistake")
}

func (q *RMQPull) Consume(ctx c.Context) (msgqueue.Incoming, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case dlvr, ok := <-q.dlvr:
		if !ok {
			return nil, errors.New("unexpected void delivery")
		}

		return &RMQIncoming{&dlvr}, nil
	}
}
