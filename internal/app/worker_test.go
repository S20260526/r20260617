package app

import (
	"context"
	"internal/app/msgqueue"
	"testing"
	"time"
)

func TestWorkerOK(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultYes}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		EventsTable: "events",
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || q.url != "host:5672" {
		t.Error()
	}

	if s.trace != "R" || s.key != "id1" {
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

	if m.in != 1 || m.yes != 1 || m.no != 0 || m.fail != 0 || m.err != 0 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}

func TestWorkerPullFail(t *testing.T) {
	q := &mockQueue{payload: "", fail: true}
	s := &mockStorage{}
	c := &mockCoprocess{}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller: msgqueue.NewPuller(
			[]string{"host1:5672", "host2:5672", "host3:5672"}, q,
		),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
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

	if m.in != 3 || m.yes != 0 || m.no != 0 || m.fail != 0 || m.err != 3 {
		t.Error()
	}

	t0 := time.Time{}

	if m.beginTime != t0 {
		t.Error()
	}
}

func TestWorkerUnmarshalFail(t *testing.T) {
	q := &mockQueue{payload: "[1234"}
	s := &mockStorage{}
	c := &mockCoprocess{}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "" || c.trace != "" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}

	if m.in != 1 || m.yes != 0 || m.no != 0 || m.fail != 0 || m.err != 1 {
		t.Error()
	}

	t0 := time.Time{}

	if m.beginTime != t0 {
		t.Error()
	}
}

func TestWorkerStorageFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{fail: true}
	c := &mockCoprocess{}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}

	if m.in != 1 || m.yes != 0 || m.no != 0 || m.fail != 0 || m.err != 1 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}

func TestWorkerCoprocessFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{failed: true}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}
	if m.in != 1 || m.yes != 0 || m.no != 0 || m.fail != 0 || m.err != 1 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}

func TestWorkerCoprocessResultNo(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultNo}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "RD" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "A" {
		t.Error()
	}

	if m.in != 1 || m.yes != 0 || m.no != 1 || m.fail != 0 || m.err != 0 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}

func TestWorkerCoprocessResultFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultFail}
	r := &mockRegistry{}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "" ||
		q.in.acknowledged != "R" {
		t.Error()
	}

	if m.in != 1 || m.yes != 0 || m.no != 0 || m.fail != 1 || m.err != 0 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}

func TestWorkerRegistratorFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mockQueue{payload: string(om)}
	s := &mockStorage{payload: "1234"}
	c := &mockCoprocess{result: CoprocessResultYes}
	r := &mockRegistry{failed: true}
	m := &mockMetrics{}

	w := Worker{
		Puller:      msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:     s,
		Coprocess:   c,
		Registrator: r,
		Metrics:     m,
	}

	if w.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || r.trace != "P" ||
		r.recent.Timestamp != tstmp || r.recent.Id != "id1" ||
		q.in.acknowledged != "R" {
		t.Error()
	}

	if m.in != 1 || m.yes != 1 || m.no != 0 || m.fail != 0 || m.err != 1 {
		t.Error()
	}

	if m.beginTime != tstmp {
		t.Error()
	}
}
