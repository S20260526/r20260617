package app

import (
	"context"
	"domain"
	"fmt"
	"log/slog"
	"math/rand"
)

const traceIdKey = "TraceId"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceIdKey, fmt.Sprintf("%x", rand.Uint64()))
}

func (s *Service) GetTraceId(ctx context.Context) string {
	return ctx.Value(traceIdKey).(string)
}

func (s *Service) Hello(ctx context.Context) *domain.Hello {
	h := domain.NewHello()

	slog.Info("create", "traceid", s.GetTraceId(ctx), "class", h.String(), "address", fmt.Sprintf("%p", &h))

	return &h
}

func (s *Service) World(ctx context.Context) *domain.World {
	w := domain.NewWorld()

	slog.Info("create", "traceid", s.GetTraceId(ctx), "class", w.String(), "address", fmt.Sprintf("%p", &w))

	return &w
}
