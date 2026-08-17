package app

import (
	"domain"
	"testing"
)

func cmp[U comparable](dt domain.Term, u *U, t *testing.T) {
	u1 := dt.(U)

	if &u1 == u {
		t.Logf("note: got a singleton of %#v@%p", *u, u)
	}

	if u1 != *u {
		t.Fatal("instances not equal")
	}
}

type MockHelloRepo struct {
}

func (r MockHelloRepo) Get() domain.Term {
	return domain.NewHello()
}

func TestHello(t *testing.T) {
	s := NewHelloService(MockHelloRepo{})
	h := domain.NewHello()

	cmp(s.Get(), &h, t)
}

func TestWorld(t *testing.T) {
	s := NewWorldService()
	w := domain.NewWorld()

	cmp(s.Get(), &w, t)
}
