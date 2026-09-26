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

type RMQConnection struct {
	conn *amqp.Connection
}

func NewRMQConnection() *RMQConnection {
	return &RMQConnection{}
}

type RMQQueue struct {
	conn *RMQConnection
	name string
	chnl *amqp.Channel
}

type RMQPublishing struct {
	RMQQueue
}

type RMQConsuming struct {
	RMQQueue
	dlvr <-chan amqp.Delivery
}

func (q *RMQConnection) connect(url string) error {
	conn, err := amqp.Dial(url)

	if err == nil {
		q.conn = conn
	}

	return err
}

func (q *RMQConnection) channel() (*amqp.Channel, error) {
	return q.conn.Channel()
}

func (q *RMQConnection) doClose() {
	q.conn.Close()
}

func NewRMQPublishing(conn *RMQConnection, name string) *RMQPublishing {
	return &RMQPublishing{RMQQueue{conn: conn, name: name}}
}

func NewRMQConsuming(conn *RMQConnection, name string) *RMQConsuming {
	return &RMQConsuming{RMQQueue{conn: conn, name: name}, nil}
}

func (q *RMQQueue) Connect(url string) error {
	return q.conn.connect(url)
}

func (q *RMQPublishing) OpenChannel() error {
	chnl, err := q.conn.channel()

	q.chnl = chnl

	return err
}

func (q *RMQConsuming) OpenChannel() error {
	chnl, err := q.conn.channel()

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
			false, // Auto-Ack
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

func (q *RMQPublishing) CloseChannel() {
	q.chnl.Close()
}

func (q *RMQConsuming) CloseChannel() {
	q.dlvr = nil

	q.chnl.Close()
}

func (q *RMQQueue) Disconnect() {
	q.conn.doClose()
}

func (q *RMQPublishing) Publish(ctx c.Context, msg []byte) error {
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

func (q *RMQConsuming) Consume(ctx c.Context) (msgqueue.Incoming, error) {
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
