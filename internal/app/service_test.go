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

func TestContext(t *testing.T) {
	s := NewService()

	ctx := s.CreateContext(context.Background())

	if s.GetTraceId(ctx) == "" {
		t.Fatal("got empty TraceId")
	}
}

func TestService(t *testing.T) {
	s := NewService()

	h := domain.NewHello()
	w := domain.NewWorld()

	ctx := s.CreateContext(context.Background())

	cmp(s.Hello(ctx), &h, t)
	cmp(s.World(ctx), &w, t)
}
