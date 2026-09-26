package msgqueue

import (
	c "context"
	"slices"
	"testing"
)

var urls = []string{"U", "V", "W"}

func TestPullOptimistic(t *testing.T) {
	q := &mq{inData: "1234"}

	p := NewPuller(urls, q)

	d, err := p.Pull(c.Background())

	if string(d.GetData()) != "1234" || err != nil || q.trace != "cU o C " {
		t.Fatal()
	}

	q.inData = "ABCD"

	d, err = p.Pull(c.Background())

	if string(d.GetData()) != "ABCD" || err != nil || q.trace != "cU o C C " {
		t.Fatal()
	}
}

func TestPullConnectFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: connectFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull(c.Background())

	if d != nil || err != connectFailed || q.trace != "cU " {
		t.Fatal()
	}

	q.failWith = connectFailed

	d, err = p.Pull(c.Background())

	if d != nil || err != connectFailed || q.trace != "cU cV " {
		t.Fatal()
	}

	q.failWith = connectFailed

	d, err = p.Pull(c.Background())

	if d != nil || err != connectFailed || q.trace != "cU cV cW " {
		t.Fatal()
	}

	d, err = p.Pull(c.Background())

	if string(d.GetData()) != "1234" || err != nil || q.trace != "cU cV cW cU o C " {
		t.Fatal()
	}
}

func TestPullOpenChannelFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: channelOpenFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull(c.Background())

	if d != nil || err != channelOpenFailed || q.trace != "cU o D " {
		t.Fatal()
	}

	d, err = p.Pull(c.Background())

	if string(d.GetData()) != "1234" || err != nil || q.trace != "cU o D cV o C " {
		t.Fatal()
	}
}

func TestPullConsumeFailed(t *testing.T) {
	q := &mq{inData: "1234", failWith: consumeFailed}

	p := NewPuller(urls, q)

	d, err := p.Pull(c.Background())

	if d != nil || err != consumeFailed || q.trace != "cU o C X D " {
		t.Fatal()
	}

	d, err = p.Pull(c.Background())

	if string(d.GetData()) != "1234" || err != nil || q.trace != "cU o C X D cV o C " {
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

	p.Pull(c.Background())
	p.Cleanup()

	if q.trace != "cU o C X D " {
		t.Fatal()
	}

	_, err := p.Pull(c.Background())

	if err != nil || q.trace != "cU o C X D cU o C " {
		t.Fatal()
	}
}

func TestPullContext(t *testing.T) {
	q := &mq{inData: "1234"}

	p := NewPuller(urls, q)

	ctx, cancel := c.WithCancel(c.Background())

	cancel()

	d, err := p.Pull(ctx)

	if d != nil || err != c.Canceled || q.trace != "cU o C X D " {
		t.Fatal()
	}
}

func TestPullUrl(t *testing.T) {
	q := &mq{inData: "1234"}

	p := NewPuller(urls, q)

	if !slices.Equal(p.Url(), urls) {
		t.Fatal()
	}

	p.SetUrl([]string{"A", "B", "C"})

	if !slices.Equal(p.Url(), []string{"A", "B", "C"}) {
		t.Fatal()
	}

	d, err := p.Pull(c.Background())

	if string(d.GetData()) != "1234" || err != nil || q.trace != "cA o C " {
		t.Fatal()
	}

	p.SetUrl([]string{"U", "V", "W"})

	if !slices.Equal(p.Url(), []string{"U", "V", "W"}) {
		t.Fatal()
	}

	q.inData = "5678"

	d, err = p.Pull(c.Background())

	if string(d.GetData()) != "5678" || err != nil || q.trace != "cA o C X D cU o C " {
		t.Fatal(q.trace)
	}
}

func TestPullUrlsImmutable(t *testing.T) {
	q := &mq{}

	u := []string{"A", "B", "C"}

	p := NewPuller(u, q)

	u[0] = "Q"

	if !slices.Equal(p.Url(), []string{"A", "B", "C"}) {
		t.Fatal()
	}

	p.Url()[0] = "Q"

	if !slices.Equal(p.Url(), []string{"A", "B", "C"}) {
		t.Fatal()
	}

	p.SetUrl(u)

	u[0] = "A"

	if !slices.Equal(p.Url(), []string{"Q", "B", "C"}) {
		t.Fatal()
	}

}
