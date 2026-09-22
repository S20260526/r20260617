package msgqueue

import (
	"testing"
)

var urls = []string{"U", "V", "W"}

func TestPullOptimistic(t *testing.T) {
	q := &mq{inData: "1234"}

	p := NewPuller(urls, q)

	d, err := p.Pull()

	if d != "1234" || err != nil || q.trace != "cU o C " {
		t.Fatal()
	}

	q.inData = "ABCD"

	d, err = p.Pull()

	if d != "ABCD" || err != nil || q.trace != "cU o C C " {
		t.Fatal()
	}
}

func TestPullConnectFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: connectFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull()

	if d != "" || err != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	q.failWith = connectFailed

	d, err = p.Pull()

	if d != "" || err != connectFailed || q.trace != "cU cV " {
		t.Fatal()
	}

	q.failWith = connectFailed

	d, err = p.Pull()

	if d != "" || err != connectFailed || q.trace != "cU cV cW " {
		t.Fatal()
	}

	d, err = p.Pull()

	if d != "1234" || err != nil || q.trace != "cU cV cW cU o C " {
		t.Fatal()
	}
}

func TestPullOpenChannelFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: channelOpenFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull()

	if d != "" || err != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	d, err = p.Pull()

	if d != "1234" || err != nil || q.trace != "cU o D cV o C " {
		t.Fatal()
	}
}

func TestPullConsumeFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: consumeFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull()

	if d != "" || err != consumeFailed || q.trace != "cU o C X D " {
		t.Fatal()
	}

	d, err = p.Pull()

	if d != "1234" || err != nil || q.trace != "cU o C X D cV o C " {
		t.Fatal()
	}
}

func TestPullCleanup(t *testing.T) {
	q := &mq{inData: "1234"}

	p := NewPuller(urls, q)

	p.Cleanup()

	if q.trace != "" {
		t.Fatal()
	}

	p.Pull()
	p.Cleanup()

	if q.trace != "cU o C X D " {
		t.Fatal()
	}

	_, err := p.Pull()

	if err != nil || q.trace != "cU o C X D cU o C " {
		t.Fatal()
	}
}
