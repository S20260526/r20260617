package msgqueue

import (
	c "context"
	"testing"
)

func TestPushOptimistic(t *testing.T) {
	q := &mockQueue{}
	p := NewPusher(urlU, q)

	if p.Push(c.Background(), []byte("1234")) != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}
}

func TestPushConnectFailed(t *testing.T) {
	q := &mockQueue{failWith: connectFailed}
	p := NewPusher(urlU, q)

	if p.Push(c.Background(), []byte("1234")) != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	if p.Push(c.Background(), []byte("1234")) != nil || q.trace != "cU cU o p1234 " {
		t.Fatal()
	}
}

func TestPushOpenChannelFailed(t *testing.T) {
	q := &mockQueue{failWith: channelOpenFailed}
	p := NewPusher(urlU, q)

	if p.Push(c.Background(), []byte("1234")) != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	if p.Push(c.Background(), []byte("1234")) != nil || q.trace != "cU o D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushPublishFailed(t *testing.T) {
	q := &mockQueue{failWith: publishFailed}
	p := NewPusher(urlU, q)

	if p.Push(c.Background(), []byte("1234")) != publishFailed || q.trace != "cU o p1234 X D " {
		t.Fatal(q.trace)
	}

	if p.Push(c.Background(), []byte("1234")) != nil || q.trace != "cU o p1234 X D cU o p1234 " {
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

	p.Push(c.Background(), []byte("1234"))

	p.Cleanup()

	if q.trace != "cU o p1234 X D " {
		t.Fatal()
	}
}

func TestPushContext(t *testing.T) {
	q := &mockQueue{}
	p := NewPusher(urlU, q)

	ctx, cancel := c.WithCancel(c.Background())

	cancel()

	if p.Push(ctx, []byte("1234")) != c.Canceled || q.trace != "cU o p1234 X D " { // FIXME m.b. don't cleanup
		t.Fatal()
	}
}
