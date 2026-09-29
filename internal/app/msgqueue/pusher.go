package msgqueue

import (
	"context"
)

type Pusher struct {
	url   string
	queue PublishingQueue

	ready bool
}

func NewPusher(url string, q PublishingQueue) *Pusher {
	return &Pusher{url: url, queue: q, ready: false}
}

func (p *Pusher) Push(ctx context.Context, payload []byte) error {
	if !p.ready {
		err := p.queue.Connect(p.url)

		if err != nil {
			return err
		}
		err = p.queue.OpenChannel()

		if err != nil {
			p.queue.Disconnect()

			return err
		}

		p.ready = true
	}

	err := p.queue.Publish(ctx, payload)

	if err != nil {
		p.Cleanup()

		return err
	}

	return nil
}

func (p *Pusher) Cleanup() {
	if p.ready {
		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}
}
