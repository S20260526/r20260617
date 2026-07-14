package app

import (
	"domain"
)

type HelloService struct {
}

func NewHelloService() *HelloService {
	return &HelloService{}
}

func (s *HelloService) Get() domain.Term {
	return domain.NewHello()
}
