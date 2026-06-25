package infra

import (
	"app"
	"testing"
)

func TestUseCase(t *testing.T) {
	NewHandler(app.NewService())
}
