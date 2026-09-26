package msgqueue

import (
	"context"
)

type Puller struct {
	url   []string
	curr  int
	queue ConsumingQueue

	ready bool
}

func copyUrl(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)

	return out
}

func NewPuller(url []string, queue ConsumingQueue) *Puller {
	return &Puller{url: copyUrl(url), curr: 0, queue: queue, ready: false}
}

func (p *Puller) Url() []string {
	return copyUrl(p.url)
}

func (p *Puller) SetUrl(url []string) {
	p.Cleanup()

	p.url = copyUrl(url)
}

func (p *Puller) Pull(ctx context.Context) (Incoming, error) {
	var err error

	if !p.ready {
		err = p.queue.Connect(p.url[p.curr])

		if err == nil {
			err = p.queue.OpenChannel()

			if err != nil {
				p.queue.Disconnect()
			} else {
				p.ready = true
			}
		}
	}

	if p.ready {
		var d Incoming

		d, err = p.queue.Consume(ctx)

		if err == nil {
			return d, nil
		}

		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}

	p.curr++

	if p.curr == len(p.url) {
		p.curr = 0
	}

	return nil, err
}

func (p *Puller) Cleanup() {
	if p.ready {
		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}
}
