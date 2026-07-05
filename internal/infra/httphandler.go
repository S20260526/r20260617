package infra

import (
	"context"
	"domain"
	"fmt"
	"log/slog"
	"net/http"
)

type UseCase interface {
	CreateContext(ctx context.Context) context.Context
	GetTraceId(ctx context.Context) string

	Hello(ctx context.Context) *domain.Hello
	World(ctx context.Context) *domain.World
}

type Handler struct {
	uc UseCase
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/hcheck" {
		return
	}

	ctx := h.uc.CreateContext(r.Context())

	slog.Info("request", "traceid", h.uc.GetTraceId(ctx), "from", r.RemoteAddr, "to", r.Host, "URI", r.RequestURI)

	w.Write([]byte(fmt.Sprintf("%s, %s!", h.uc.Hello(ctx), h.uc.World(ctx))))
}

func NewHandler(uc UseCase) Handler {
	return Handler{uc: uc}
}
