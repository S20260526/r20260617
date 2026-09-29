package app

import (
	"context"
	"testing"
	"internal/app/msgqueue"
)

func TestFrontOK(t *testing.T) {
	s := &mockStorage{key: "id1"}
	q := &mockQueue{}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
	}

	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	if f.Push(context.Background(), tstmp, []byte("1234")) != nil {
		t.Fatal()
	}

	if s.trace != "C" {
		t.Error()
	}

	if q.trace != "COP" || q.url != "host:5672" || q.payload != string(om) {
		t.Error(q.trace)
	}
}

func TestFrontStoreFail(t *testing.T) {
	s := &mockStorage{fail: true}
	q := &mockQueue{}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
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
	s := &mockStorage{}
	q := &mockQueue{fail: true}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
	}

	if f.Push(context.Background(), tstmp, []byte("1234")) != fail {
		t.Fatal()
	}

	if s.trace != "CD" {
		t.Error()
	}

	if q.trace != "COPXD" || q.url != "host:5672" {
		t.Error()
	}
}
