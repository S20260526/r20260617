package app

import (
	"context"
	"internal/app/msgqueue"
	"internal/grpcipc"
	"testing"
)

func (s *ms) Read(ctx context.Context, key string) ([]byte, error) {
	s.trace += "R"

	if s.fail {
		return nil, fail
	}

	return []byte(s.payload), nil
}

type mc struct {
	trace   string
	payload string
	failed  bool
	result  CoprocessResult
}

func (c *mc) Call(ctx context.Context, request *CoprocessRequest) (*CoprocessResponse, error) {
	c.trace += "C"
	c.payload = string(request.Payload)

	if c.failed {
		return nil, fail
	}

	return &CoprocessResponse{Result: grpcipc.Result(c.result)}, nil
}

func (c *mc) Wait() error {
	return fail
}

func TestBackOK(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mq{payload: string(om)}
	s := &ms{payload: "1234"}
	c := &mc{result: CoprocessResultYes}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != nil {
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

	if q.in.acknowledged != "A" {
		t.Error()
	}
}

func TestBackPullFail(t *testing.T) {
	q := &mq{payload: "", fail: true}
	s := &ms{}
	c := &mc{}

	b := Back{
		Puller: msgqueue.NewPuller(
			[]string{"host1:5672", "host2:5672", "host3:5672"}, q,
		),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COGXD" || q.url != "host1:5672" {
		t.Error()
	}

	if b.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if b.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COGXDCOGXDCOGXD" || q.url != "host3:5672" {
		t.Error()
	}

	if s.trace != "" || c.trace != "" {
		t.Error()
	}
}

func TestBackUnmarshalFail(t *testing.T) {
	q := &mq{payload: "[1234"}
	s := &ms{}
	c := &mc{}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) == nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "" || c.trace != "" || q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestBackStorageFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mq{payload: string(om)}
	s := &ms{fail: true}
	c := &mc{}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "" || q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestBackCoprocessFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mq{payload: string(om)}
	s := &ms{payload: "1234"}
	c := &mc{failed: true}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != fail {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || q.in.acknowledged != "R" {
		t.Error()
	}
}

func TestBackCoprocessResultNo(t *testing.T) {

	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mq{payload: string(om)}
	s := &ms{payload: "1234"}
	c := &mc{result: CoprocessResultNo}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "RD" || c.trace != "C" || q.in.acknowledged != "A" {
		t.Error(q.trace, s.trace, c.trace)
	}
}

func TestBackCoprocessResultFail(t *testing.T) {
	om, _ := Order{Timestamp: tstmp, BlobId: "id1"}.Marshal()

	q := &mq{payload: string(om)}
	s := &ms{payload: "1234"}
	c := &mc{result: CoprocessResultFail}

	b := Back{
		Puller:    msgqueue.NewPuller([]string{"host1:5672"}, q),
		Storage:   s,
		Coprocess: c,
	}

	if b.Pull(context.Background()) != nil {
		t.Fatal()
	}

	if q.trace != "COG" || s.trace != "R" || c.trace != "C" || q.in.acknowledged != "R" {
		t.Error(q.trace, s.trace, c.trace)
	}
}
