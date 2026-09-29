package app

import (
	"context"
	"internal/app/msgqueue"
	"testing"
)

func TestJanitorOK(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}

	j := Janitor{
		Puller:  msgqueue.NewPuller([]string{"host:5672"}, q),
		Storage: s,
	}

	if j.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || q.url != "host:5672" {
		t.Error()
	}

	if s.trace != "D" || s.key != "id1" {
		t.Error()
	}

	if q.in.acknowledged != "A" {
		t.Error()
	}
}

func TestJanitorPullFail(t *testing.T) {
	q := &mockQueue{fail: true}
	s := &mockStorage{}

	j := Janitor{
		Puller: msgqueue.NewPuller(
			[]string{
				"host1:5672",
				"host2:5672",
				"host3:5672",
			}, q,
		),
		Storage: s,
	}

	if j.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COGXD" || q.url != "host1:5672" {
		t.Error()
	}

	if j.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COGXDCOGXD" || q.url != "host2:5672" {
		t.Error()
	}

	if j.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COGXDCOGXDCOGXD" || q.url != "host3:5672" {
		t.Error()
	}

	if s.trace != "" {
		t.Error()
	}
}

func TestJanitorUnmarshallFail(t *testing.T) {
	q := &mockQueue{payload: "ABCD"}
	s := &mockStorage{payload: "1234"}

	j := Janitor{
		Puller:  msgqueue.NewPuller([]string{"host:5672"}, q),
		Storage: s,
	}

	if j.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "" || q.in.acknowledged != "A" {
		t.Error()
	}
}

func TestJanitorStorageFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{fail: true}

	j := Janitor{
		Puller:  msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage: s,
	}

	if j.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "D" || q.in.acknowledged != "A" {
		t.Error()
	}
}
