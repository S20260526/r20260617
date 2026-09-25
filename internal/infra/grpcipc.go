package infra

import (
	"context"
	"internal/app"
	"internal/grpcipc"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log/slog"
)

type GrpcIpcClient struct {
	socketpath string
	client     grpcipc.IpcClient
}

func NewGrpcIpcClient(socketpath string) *GrpcIpcClient {
	g := &GrpcIpcClient{socketpath: socketpath}

	conn, err := grpc.NewClient(
		"unix://"+g.socketpath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		slog.Warn(
			"infra",
			"where", "gRPC-IPC",
			"when", "Client create",
			"what", err,
		)

		os.Exit(1)
	}

	g.client = grpcipc.NewIpcClient(conn)

	return g
}

func (g *GrpcIpcClient) Process(ctx context.Context, rqst *app.CoprocessRequest) (*app.CoprocessResponse, error) {
	return g.client.Process(ctx, rqst)
}
