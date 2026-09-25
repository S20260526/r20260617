package app

import (
	c "context"
	"internal/grpcipc"
)

type CoprocessRequest = grpcipc.Request
type CoprocessResponse = grpcipc.Response

type Ipc interface {
	SocketName() string
	Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error)
}

type Coprocess interface {
	Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error)
	Wait() error
}
