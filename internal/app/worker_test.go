package app

import (
	"context"
	"internal/app/msgqueue"
	"testing"
)

func TestWorkerOK(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultYes}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		EventsTable: "events",
		Registrator: r,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || q.url != "host:5672" {
		t.Error()
	}

	if s.trace != "R" {
		t.Error()
	}

	if c.trace != "C" || c.payload != "1234" {
		t.Error()
	}

	if r.trace != "P" || r.table != "events" ||
		r.recent.Timestamp != tstmp || r.recent.Id != "id1" {
		t.Error()
	}

	if q.in.acknowledged != "A" {
		t.Error()
	}
}

func TestWorkerPullFail(t *testing.T) {
	q := &mockQueue{payload: "", fail: true}
	s := &mockStorage{}
	c := &mockCoprocess{}
	r := &mockRegistry{}

	w := Worker{
		Puller: msgqueue.NewPuller(
			[]string{"host1:5672", "host2:5672", "host3:5672"}, q,
		),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COGXD" || q.url != "host1:5672" {
		t.Error()
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COGXDCOGXDCOGXD" || q.url != "host3:5672" {
		t.Error()
	}

	if s.trace != "" || c.trace != "" || r.trace != "" {
		t.Error()
	}
}

func TestWorkerUnmarshalFail(t *testing.T) {
	q := &mockQueue{payload: "[1234"}
	s := &mockStorage{}
	c := &mockCoprocess{}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "" || c.trace != "" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestWorkerStorageFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{fail: true}
	c := &mockCoprocess{}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestWorkerCoprocessFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{failed: true}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestWorkerCoprocessResultNo(t *testing.T) {

	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultNo}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "RD" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "A" {
		t.Error()
	}
}

func TestWorkerCoprocessResultFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultFail}
	r := &mockRegistry{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestWorkerRegistratorFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultYes}
	r := &mockRegistry{failed: true}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "P" ||
		r.recent.Timestamp != tstmp || r.recent.Id != "id1" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
}
