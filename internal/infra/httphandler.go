package infra

import (
	"domain"
	"fmt"
	"log/slog"
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
	slog.Info("request", "from", r.RemoteAddr, "to", r.Host, "URI", r.RequestURI)

	w.Write([]byte(fmt.Sprintf("%s, %s!", h.uc.Hello(), h.uc.World())))
}

func NewHandler(uc UseCase) Handler {
	return Handler{uc: uc}
}
