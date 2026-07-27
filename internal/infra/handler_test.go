package infra

import (
	"app"
	"testing"
)

func TestUseCase(t *testing.T) {
	hh := NewHandler(app.NewHelloService())

	if hh.Get() != "Hello" {
		t.Error("Hello not met")
	}

	wh := NewHandler(app.NewWorldService())

	if wh.Get() != "World" {
		t.Error("World not met")
	}
}
