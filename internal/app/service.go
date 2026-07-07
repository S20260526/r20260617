package app

import (
	"context"
	"domain"
	"fmt"
	"log/slog"
	"math/rand"
)

const traceIdKey = "TraceId"

type Counter interface {
	Inc()
}

type Metrics interface {
	NewCounter(name, human_readable string) Counter
}

type Service struct {
	hello_cnt, world_cnt Counter
}

func NewService(m Metrics) *Service {
	return &Service{
		hello_cnt: m.NewCounter("hello_cnt", "hello instance allocated count"),
		world_cnt: m.NewCounter("world_cnt", "world instance allocated count"),
	}
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

	s.hello_cnt.Inc()

	return &h
}

func (s *Service) World(ctx context.Context) *domain.World {
	w := domain.NewWorld()

	slog.Info("create", "traceid", s.GetTraceId(ctx), "class", w.String(), "address", fmt.Sprintf("%p", &w))

	s.world_cnt.Inc()

	return &w
}
