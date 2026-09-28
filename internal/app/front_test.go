package app

import (
	"context"
	"errors"
	"internal/app/msgqueue"
	"testing"
	"time"
)

var fail = errors.New("fail")
var tstmp = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

type mi struct {
	payload      string
	acknowledged string
}

func (i *mi) GetData() []byte {
	return []byte(i.payload)
}

func (i *mi) Acknowledge() {
	i.acknowledged = "A"
}

func (i *mi) Reject() {
	i.acknowledged = "R"
}

type mq struct {
	fail    bool
	trace   string
	url     string
	payload string
	in      *mi
}

func (q *mq) Connect(url string) error {
	q.trace += "C"
	q.url = url

	return nil
}

func (q *mq) OpenChannel() error {
	q.trace += "O"

	return nil
}
func (q *mq) CloseChannel() {
	q.trace += "X"
}

func (q *mq) Disconnect() {
	q.trace += "D"
}

func (q *mq) Publish(ctx context.Context, msg []byte) error {
	q.trace += "P"
	q.payload = string(msg)

	if q.fail {
		return fail
	}

	return nil
}

func (q *mq) Consume(ctx context.Context) (msgqueue.Incoming, error) {
	q.trace += "G"

	if q.fail {
		return nil, fail
	}

	q.in = &mi{payload: q.payload}

	return q.in, nil
}

type ms struct {
	fail    bool
	trace   string
	payload string
}

func (s *ms) Create(ctx context.Context, blob []byte) (string, error) {
	s.trace += "C"

	if s.fail {
		return "", fail
	}

	s.payload = string(blob)

	return "id1", nil
}

func (s *ms) Delete(ctx context.Context, key string) error {
	s.trace += "D"

	return nil
}

func TestFrontOK(t *testing.T) {
	s := &ms{}
	q := &mq{}

	f := Front{
		BrokerUrl: HostColonPort{Host: "host", Port: 5672},
		Storage:   s,
		Queue:     q,
	}

	d, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	if f.Push(context.Background(), tstmp, []byte("1234")) != nil {
		t.Fatal()
	}

	if s.trace != "C" {
		t.Error()
	}

	if q.trace != "COPX" || q.url != "host:5672" || q.payload != string(d) {
		t.Error()
	}
}

func TestFrontStoreFail(t *testing.T) {
	s := &ms{fail: true}
	q := &mq{}

	f := Front{
		BrokerUrl: HostColonPort{Host: "host", Port: 5672},
		Storage:   s,
		Queue:     q,
	}

	if f.Push(context.Background(), tstmp, []byte("1234")) != fail {
		t.Fatal()
	}

	if s.trace != "C" {
		t.Error()
	}

	if q.trace != "" {
		t.Error()
	}
}

func TestFrontQueueFail(t *testing.T) {
	s := &ms{}
	q := &mq{fail: true}

	f := Front{
		BrokerUrl: HostColonPort{Host: "host", Port: 5672},
		Storage:   s,
		Queue:     q,
	}

	if f.Push(context.Background(), tstmp, []byte("1234")) != fail {
		t.Fatal()
	}

	if s.trace != "CD" {
		t.Error()
	}

	if q.trace != "COPXDX" || q.url != "host:5672" {
		t.Error()
	}
}
