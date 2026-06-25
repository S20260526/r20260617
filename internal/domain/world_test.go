package domain

import (
	"testing"
)

func TestWorld(t *testing.T) {
	w := NewWorld()
	a := w.String()
	e := "World"

	if a != e {
		t.Fatalf("got \"%s\", \"%s\" expected", a, e)
	}
}
