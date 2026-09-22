package msgqueue

import (
	"testing"
)

var url = "U"

func TestPushOptimistic(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	p.Charge("1234")

	if p.Push() != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}
}

func TestPushConnectFailed(t *testing.T) {
	q := &mq{failWith: connectFailed}
	p := NewPusher(url, q)

	p.Charge("1234")

	if p.Push() != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	if p.Push() != nil || q.trace != "cU cU o p1234 " {
		t.Fatal()
	}
}

func TestPushOpenChannelFailed(t *testing.T) {
	q := &mq{failWith: channelOpenFailed}
	p := NewPusher(url, q)

	p.Charge("1234")

	if p.Push() != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	if p.Push() != nil || q.trace != "cU o D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushPublishFailed(t *testing.T) {
	q := &mq{failWith: publishFailed}
	p := NewPusher(url, q)

	p.Charge("1234")

	if p.Push() != publishFailed || q.trace != "cU o p1234 X D " {
		t.Fatal(q.trace)
	}

	if p.Push() != nil || q.trace != "cU o p1234 X D cU o p1234 " {
		t.Fatal()
	}
}

func TestPushCharge(t *testing.T) {
	q := &mq{}
	p := NewPusher(url, q)

	if p.Push() != notCharged || q.trace != "" {
		t.Fatal()
	}

	p.Charge("1234")

	if p.Push() != nil || q.trace != "cU o p1234 " {
		t.Fatal()
	}

	if p.Push() != notCharged || q.trace != "cU o p1234 " {
		t.Fatal()
	}

	p.Charge("ABCD")

	if p.Push() != nil || q.trace != "cU o p1234 pABCD " {
		t.Fatal()
	}

	p.Charge("EFGH")

	q.failWith = publishFailed

	if p.Push() != publishFailed || q.trace != "cU o p1234 pABCD pEFGH X D " {
		t.Fatal()
	}

	if p.Push() != nil || q.trace != "cU o p1234 pABCD pEFGH X D cU o pEFGH " {
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

	p.Charge("1234")
	p.Push()

	p.Cleanup()

	if q.trace != "cU o p1234 X D " {
		t.Fatal()
	}

	if p.Push() != notCharged || q.trace != "cU o p1234 X D " {
		t.Fatal()
	}
}
