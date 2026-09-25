package app

import (
	c "context"
	"internal/grpcipc"
	"os/exec"
)

type CoprocessRequest = grpcipc.Request
type CoprocessResponse = grpcipc.Response

type Ipc interface {
	SocketName() string
	Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error)
}

type Coprocess struct {
	ipc Ipc
	cmd *exec.Cmd
}

func NewCoprocess(ctx c.Context, ipc Ipc, name string, args ...string) (*Coprocess, error) {
	allargs := append(args, ipc.SocketName())

	cmd := exec.CommandContext(ctx, name, allargs...)

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return &Coprocess{ipc, cmd}, nil
}

func (c *Coprocess) Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error) {
	return c.ipc.Call(ctx, request)
}

func (c *Coprocess) Wait() error {
	return c.cmd.Wait()
}
