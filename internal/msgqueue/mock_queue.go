package msgqueue

import (
	"errors"
)

var connectFailed = errors.New("ECONN")
var channelOpenFailed = errors.New("ECHAN")
var publishFailed = errors.New("EPBSH")
var consumeFailed = errors.New("EPULL")

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

func (q *mq) Publish(b []byte) error {
	q.trace += "p" + string(b) + " "

	if q.failWith == publishFailed {
		q.failWith = nil

		return publishFailed
	}

	return nil
}

func (q *mq) Consume() ([]byte, error) {
	q.trace += "C "

	if q.failWith == consumeFailed {
		q.failWith = nil

		return []byte{}, consumeFailed
	}

	return []byte(q.inData), nil
}
