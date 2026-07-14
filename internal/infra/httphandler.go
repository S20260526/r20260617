package infra

import (
	"domain"
	"net/http"
)

type UseCase interface {
	Get() domain.Term
}

type Handler struct {
	uc UseCase
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.uc.Get().String()))
}

func NewHandler(uc UseCase) Handler {
	return Handler{uc: uc}
}
