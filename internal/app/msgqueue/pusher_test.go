package msgqueue

import (
	c "context"
	"testing"
)

func TestPushOptimistic(t *testing.T) {
	q := &mockQueue{}
	p := NewPusher(urlU, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}
}

func TestPushConnectFailed(t *testing.T) {
	q := &mockQueue{failWith: connectFailed}
	p := NewPusher(urlU, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cU cU o p1234 " {
		t.Fatal()
	}
}

func TestPushOpenChannelFailed(t *testing.T) {
	q := &mockQueue{failWith: channelOpenFailed}
	p := NewPusher(urlU, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	if p.Push(c.Background()) != nil || q.trace != "cU o D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushPublishFailed(t *testing.T) {
	q := &mockQueue{failWith: publishFailed}
	p := NewPusher(urlU, q)

	p.Charge([]byte("1234"))

	if p.Push(c.Background()) != publishFailed || q.trace != "cU o p1234 X D " {
		t.Fatal(q.trace)
	}

	if p.Push(c.Background()) != nil || q.trace != "cU o p1234 X D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushCharge(t *testing.T) {
	q := &mockQueue{}
	p := NewPusher(urlU, q)

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
	q := &mockQueue{}
	p := NewPusher(urlU, q)

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
	q := &mockQueue{}
	p := NewPusher(urlU, q)

	ctx, cancel := c.WithCancel(c.Background())

	p.Charge([]byte("1234"))

	cancel()

	if p.Push(ctx) != c.Canceled || q.trace != "cU o p1234 X D " { // FIXME m.b. don't cleanup
		t.Fatal()
	}
}
