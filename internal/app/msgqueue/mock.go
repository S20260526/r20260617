package msgqueue

import (
	"context"
	"errors"
)

var connectFailed = errors.New("ECONN")
var channelOpenFailed = errors.New("ECHAN")
var publishFailed = errors.New("EPBSH")
var consumeFailed = errors.New("EPULL")

var urlU = "U"
var urlsUVW = []string{"U", "V", "W"}

type mockIncoming struct {
	data string
}

func (m mockIncoming) GetData() []byte {
	return []byte(m.data)
}

func (m mockIncoming) Acknowledge() {
}

func (m mockIncoming) Reject() {
}

type mockQueue struct {
	failWith error
	inData   string
	trace    string
}

func (q *mockQueue) Connect(url string) error {
	q.trace += "c" + url + " "

	if q.failWith == connectFailed {
		q.failWith = nil

		return connectFailed
	}

	return nil
}

func (q *mockQueue) OpenChannel() error {
	q.trace += "o "

	if q.failWith == channelOpenFailed {
		q.failWith = nil

		return channelOpenFailed
	}

	return nil
}

func (q *mockQueue) CloseChannel() {
	q.trace += "X "
}

func (q *mockQueue) Disconnect() {
	q.trace += "D "
}

func (q *mockQueue) Publish(ctx context.Context, b []byte) error {
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

func (q *mockQueue) Consume(ctx context.Context) (Incoming, error) {
	q.trace += "C "

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if q.failWith == consumeFailed {
		q.failWith = nil

		return nil, consumeFailed
	}

	return mockIncoming{q.inData}, nil
}
