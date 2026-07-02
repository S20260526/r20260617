package app

import (
	"domain"
	"fmt"
	"log/slog"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Hello() *domain.Hello {
	h := domain.NewHello()

	slog.Info("create", "class", h.String(), "address", fmt.Sprintf("%p", &h))

	return &h
}

func (s *Service) World() *domain.World {
	w := domain.NewWorld()

	slog.Info("create", "class", w.String(), "address", fmt.Sprintf("%p", &w))

	return &w
}
