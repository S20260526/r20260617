package app

import (
	"domain"

	"time"
)

type WorldService struct {
	mqueue domain.MQueue
}

func NewWorldService(mqueue domain.MQueue) *WorldService {
	return &WorldService{mqueue}
}

func (s *WorldService) Get() domain.Term {
	if s.mqueue != nil {
		s.mqueue.Put(time.Now().Format(time.StampMilli))
	}

	return domain.NewWorld()
}
