package infra

import (
	"context"
	"crypto/tls"
	"log/slog"

	ft "fubotorp"

	"google.golang.org/grpc"
)

type server struct {
	ft.UnimplementedServiceServer
	h Handler
}

func (s *server) Get(ctx context.Context, r *ft.Rqst) (*ft.Rsps, error) {
	return &ft.Rsps{Message: s.h.Get()}, nil
}

func StartGrpcServer(h Handler) error {
	tc, e := LoadTlsConfig()

	if e != nil {
		return e
	}

	tc.NextProtos = []string{"h2"}

	port := Config.Get("grpcport", "8081")

	c8n, e := tls.Listen("tcp", ":" + port, tc)

	if e != nil {
		return e
	}

	s := grpc.NewServer()

	ft.RegisterServiceServer(s, &server{h: h})

	slog.Info("grpc", "state", "starting")

	e = s.Serve(c8n)

	slog.Info("grpc", "state", "exit")

	return e
}
