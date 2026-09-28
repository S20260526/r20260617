package app

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

var fail = errors.New("fail")
var tstmp = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

type mq struct {
	fail  bool
	trace string
	url   string
	last  []byte
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
	q.last = slices.Clone(msg)

	if q.fail {
		return fail
	}

	return nil
}

type ms struct {
	fail  bool
	trace string
	last  string
}

func (s *ms) Create(ctx context.Context, blob []byte) (string, error) {
	s.trace += "C"

	if s.fail {
		return "", fail
	}

	s.last = string(blob)

	return "id1", nil
}

func (s *ms) Read(ctx context.Context, key string) ([]byte, error) {
	s.trace += "R"

	return nil, fail
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

	if f.Handle(context.Background(), tstmp, []byte("1234")) != nil {
		t.Fatal()
	}

	if s.trace != "C" {
		t.Error()
	}

	if q.trace != "COPX" || q.url != "host:5672" || !slices.Equal(q.last, d) {
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

	if f.Handle(context.Background(), tstmp, []byte("1234")) != fail {
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

	if f.Handle(context.Background(), tstmp, []byte("1234")) != fail {
		t.Fatal()
	}

	if s.trace != "CD" {
		t.Error()
	}

	if q.trace != "COPXDX" || q.url != "host:5672" {
		t.Error()
	}
}
