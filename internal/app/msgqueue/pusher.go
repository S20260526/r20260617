package msgqueue

import (
	"context"
	"errors"
)

var notCharged = errors.New("Not charged")

type Pusher struct {
	url   string
	queue PublishingQueue

	ready bool

	payload []byte
}

func NewPusher(url string, q PublishingQueue) *Pusher {
	return &Pusher{url: url, queue: q, ready: false, payload: nil}
}

func (p *Pusher) Url() string {
	return p.url
}

func (p *Pusher) SetUrl(url string) {
	p.Cleanup()

	p.url = url
}

func (p *Pusher) Charge(b []byte) {
	p.payload = b
}

func (p *Pusher) Push(ctx context.Context) error {
	if p.payload == nil {
		return notCharged
	}

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

	err := p.queue.Publish(ctx, p.payload)

	if err != nil {
		p.Cleanup()

		return err
	}

	p.payload = nil

	return nil
}

func (p *Pusher) Cleanup() {
	if p.ready {
		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}
}
