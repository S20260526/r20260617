package infra

import (
	"app"
	"testing"
)

type FakeCounter struct {
}

func (f FakeCounter) Inc() {
}

type FakeMetrics struct {
}

func (f FakeMetrics) NewCounter(_, _ string) app.Counter {
	return FakeCounter{}
}

func TestUseCase(t *testing.T) {
	NewHandler(app.NewService(FakeMetrics{}))
}
