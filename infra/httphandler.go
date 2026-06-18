package infra

import (
	"domain"
	"fmt"
	"net/http"
)

type UseCase interface {
	Hello() *domain.Hello
	World() *domain.World
}

type Handler struct {
	uc UseCase
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%s, %s!", h.uc.Hello(), h.uc.World())))
}

func NewHandler(uc UseCase) Handler {
	return Handler{uc: uc}
}
