package app

import (
	"domain"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Hello() *domain.Hello {
	h := domain.NewHello()

	return &h
}

func (s *Service) World() *domain.World {
	w := domain.NewWorld()

	return &w
}
