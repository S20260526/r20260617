package msgqueue

import (
	"context"
	"errors"
)

var connectFailed = errors.New("ECONN")
var channelOpenFailed = errors.New("ECHAN")
var publishFailed = errors.New("EPBSH")
var consumeFailed = errors.New("EPULL")

type mi struct {
	data string
}

func (m mi) GetData() []byte {
	return []byte(m.data)
}

func (m mi) Acknowledge() {
}

func (m mi) Reject() {
}

type mq struct {
	failWith error
	inData   string
	trace    string
}

func (q *mq) Connect(url string) error {
	q.trace += "c" + url + " "

	if q.failWith == connectFailed {
		q.failWith = nil

		return connectFailed
	}

	return nil
}

func (q *mq) OpenChannel() error {
	q.trace += "o "

	if q.failWith == channelOpenFailed {
		q.failWith = nil

		return channelOpenFailed
	}

	return nil
}

func (q *mq) CloseChannel() {
	q.trace += "X "
}

func (q *mq) Disconnect() {
	q.trace += "D "
}

func (q *mq) Publish(ctx context.Context, b []byte) error {
	q.trace += "p" + string(b) + " "

	if err := ctx.Err(); err != nil {
		return err
	}

	if q.failWith == publishFailed {
		q.failWith = nil

		return publishFailed
	}

	return nil
}

func (q *mq) Consume(ctx context.Context) (Incoming, error) {
	q.trace += "C "

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if q.failWith == consumeFailed {
		q.failWith = nil

		return nil, consumeFailed
	}

	return mi{q.inData}, nil
}
