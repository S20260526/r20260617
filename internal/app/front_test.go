package app

import (
	"context"
	"internal/app/msgqueue"
	"testing"
)

func TestFrontOK(t *testing.T) {
	s := &mockStorage{key: "id1"}
	q := &mockQueue{}
	m := &mockMetrics{}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
		Metrics: m,
	}

	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	if f.Push(context.Background(), tstmp, []byte("1234")) != nil {
		t.Fatal()
	}

	if s.trace != "C" {
		t.Error()
	}

	if q.trace != "COP" || q.url != "host:5672" || q.payload != string(om) {
		t.Error()
	}

	if m.in != 1 || m.out != 1 || m.err != 0 {
		t.Error()
	}
}

func TestFrontStoreFail(t *testing.T) {
	s := &mockStorage{fail: true}
	q := &mockQueue{}
	m := &mockMetrics{}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
		Metrics: m,
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

	if m.in != 1 || m.out != 0 || m.err != 1 {
		t.Error()
	}
}

func TestFrontQueueFail(t *testing.T) {
	s := &mockStorage{}
	q := &mockQueue{fail: true}
	m := &mockMetrics{}

	f := Front{
		Pusher:  msgqueue.NewPusher("host:5672", q),
		Storage: s,
		Metrics: m,
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

	if m.in != 1 || m.out != 0 || m.err != 1 {
		t.Error()
	}
}
