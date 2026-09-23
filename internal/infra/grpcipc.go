package infra

import (
	"context"
	"errors"
	"internal/grpcipc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log/slog"
)

type GrpcIpcRequest = grpcipc.Request
type GrpcIpcResponse = grpcipc.Response

type GrpcIpcClient struct {
	socketpath string
	client     grpcipc.IpcClient
}

func (g *GrpcIpcClient) tryConnect() grpcipc.IpcClient {
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

		return nil
	}

	return grpcipc.NewIpcClient(conn)
}

func NewGrpcIpcClient(socketpath string) *GrpcIpcClient {
	g := &GrpcIpcClient{socketpath: socketpath}

	g.tryConnect()

	return g
}

func (g *GrpcIpcClient) Process(ctx context.Context, rqst *GrpcIpcRequest) (*GrpcIpcResponse, error) {
	if g.client == nil {
		g.tryConnect()
	}

	if g.client == nil {
		return nil, errors.New("not connected")
	}

	return g.client.Process(ctx, rqst)
}
