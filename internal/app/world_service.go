package app

import (
	"domain"
)

type WorldService struct {
}

func NewWorldService() *WorldService {
	return &WorldService{}
}

func (s *WorldService) Get() domain.Term {
	return domain.NewWorld()
}
