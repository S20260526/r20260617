package msgqueue

type Puller struct {
	url   []string
	curr  int
	queue Queue

	ready bool
}

func NewPuller(url []string, queue Queue) *Puller {
	return &Puller{url: url, curr: 0, queue: queue, ready: false}
}

func (p *Puller) Pull() (string, error) {
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
		var d []byte

		d, err = p.queue.Consume()

		if err == nil {
			return string(d), nil
		}

		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}

	p.curr++

	if p.curr == len(p.url) {
		p.curr = 0
	}

	return "", err
}

func (p *Puller) Cleanup() {
	if p.ready {
		p.ready = false

		p.queue.CloseChannel()
		p.queue.Disconnect()
	}
}
