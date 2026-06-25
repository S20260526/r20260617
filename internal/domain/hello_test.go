package domain

import (
	"testing"
)

func TestHello(t *testing.T) {
	h := NewHello()
	a := h.String()
	e := "Hello"

	if a != e {
		t.Fatalf("got \"%s\", \"%s\" expected", a, e)
	}
}
