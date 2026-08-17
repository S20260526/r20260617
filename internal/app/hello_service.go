package app

import (
	"domain"
)

type HelloService struct {
	repo domain.Repo
}

func NewHelloService(repo domain.Repo) *HelloService {
	return &HelloService{repo}
}

func (s *HelloService) Get() domain.Term {
	return s.repo.Get()
}
