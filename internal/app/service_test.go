package app

import (
	"context"
	"domain"
	"testing"
)

func cmp[U comparable](u *U, v *U, t *testing.T) {
	if u == v {
		t.Logf("note: got a singleton of %#v@%p", u, v)
	}

	if *u != *v {
		t.Fatal("instances not equal")
	}
}

type FakeCounter struct {
	name, human_readable string
	n                    int
}

func (c *FakeCounter) Inc() {
	c.n += 1
}

type FakeMetrics struct {
	c []*FakeCounter
}

func (m *FakeMetrics) NewCounter(name, human_readable string) Counter {
	c := &FakeCounter{
		name:           name,
		human_readable: human_readable,
		n:              0,
	}

	m.c = append(m.c, c)

	return c
}

func TestContext(t *testing.T) {
	m := &FakeMetrics{}
	s := NewService(m)

	ctx := s.CreateContext(context.Background())

	if s.GetTraceId(ctx) == "" {
		t.Fatal("got empty TraceId")
	}
}

func TestService(t *testing.T) {
	m := &FakeMetrics{}
	s := NewService(m)

	h := domain.NewHello()
	w := domain.NewWorld()

	ctx := s.CreateContext(context.Background())

	cmp(s.Hello(ctx), &h, t)
	cmp(s.World(ctx), &w, t)

	if len(m.c) != 2 {
		t.Fatalf("metrics got %d counters, 2 expected", len(m.c))
	}

	for i, c := range m.c {
		t.Logf("metric %d: name: \"%s\", human readable text: \"%s\"", i, c.name, c.human_readable)
		if c.n != 1 {
			t.Errorf("metrics \"%s\" got value %d, 1 expected", c.name, c.n)
		}
	}
}
