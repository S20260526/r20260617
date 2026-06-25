package app

import (
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

func TestService(t *testing.T) {
	s := NewService()

	h := domain.NewHello()
	w := domain.NewWorld()

	cmp(s.Hello(), &h, t)
	cmp(s.World(), &w, t)
}
