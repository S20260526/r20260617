package infra

import (
	"app"
	"domain"

	"testing"
)

type MockHelloRepo struct {
}

func (r MockHelloRepo) Get() domain.Term {
	return domain.NewHello()
}

type MockWorldQueue struct {
}

func (q MockWorldQueue) Put(_ string) {
}

func TestUseCase(t *testing.T) {
	hh := NewHandler(app.NewHelloService(MockHelloRepo{}))

	if hh.Get() != "Hello" {
		t.Error("Hello not met")
	}

	wh := NewHandler(app.NewWorldService(MockWorldQueue{}))

	if wh.Get() != "World" {
		t.Error("World not met")
	}
}
