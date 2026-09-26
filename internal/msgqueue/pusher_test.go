package msgqueue

import (
	c "context"
	"testing"
)

var url = "U"

func TestPushOptimistic(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}
}

func TestPushConnectFailed(t *testing.T) {
	q := &mq{failWith: connectFailed}
	p := NewPusher(url, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cU cU o p1234 " {
		t.Fatal()
	}
}

func TestPushOpenChannelFailed(t *testing.T) {
	q := &mq{failWith: channelOpenFailed}
	p := NewPusher(url, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cU o D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushPublishFailed(t *testing.T) {
	q := &mq{failWith: publishFailed}
	p := NewPusher(url, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != publishFailed || q.trace != "cU o p1234 X D " {
		t.Fatal(q.trace)
	}

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 X D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushCharge(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	if p.Push(c.Background()) != notCharged || q.trace != "" {
		t.Fatal()
	}

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}

	if p.Push(c.Background()) != notCharged || q.trace != "cU o p1234 " {
		t.Fatal()
	}

	p.Charge([]byte("ABCD"))

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 pABCD " {
		t.Fatal()
	}

	p.Charge([]byte("EFGH"))

	q.failWith = publishFailed

	if p.Push(c.Background()) != publishFailed || q.trace != "cU o p1234 pABCD pEFGH X D " {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 pABCD pEFGH X D cU o pEFGH " {
		t.Fatal()
	}
}

func TestPushCleanup(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	p.Cleanup()

	if q.trace != "" {
		t.Fatal()
	}

	p.Charge([]byte("1234"))
	p.Push(c.Background())

	p.Cleanup()

	if q.trace != "cU o p1234 X D " {
		t.Fatal()
	}

	if p.Push(c.Background()) != notCharged || q.trace != "cU o p1234 X D " {
		t.Fatal()
	}
}

func TestPushContext(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	ctx, cancel := c.WithCancel(c.Background())

	p.Charge([]byte("1234"))

	cancel()

	if p.Push(ctx) != c.Canceled || q.trace != "cU o p1234 X D " { // FIXME m.b. don't cleanup
		t.Fatal()
	}
}

func TestPushSetUrl(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	if p.Url() != url {
		t.Fatal()
	}

	p.SetUrl("V")

	if p.Url() != "V" {
		t.Fatal()
	}

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != nil {
		t.Fatal()
	}

	p.Charge([]byte("5678"))

	p.SetUrl("W")

	if p.Url() != "W" {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cV o p1234 X D cW o p5678 " {
		t.Fatal()
	}
}
