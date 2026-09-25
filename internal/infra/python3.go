package infra

import (
	"context"
	"internal/app"
	"os/exec"
)

type coprocess struct {
	ipc app.Ipc
	cmd *exec.Cmd
}

func NewPython3(ctx context.Context, path string, ipc app.Ipc) (app.Coprocess, error) {
	return newCoprocess(ctx, ipc, "python3", path)
}

func newCoprocess(ctx context.Context, ipc app.Ipc, name string, args ...string) (app.Coprocess, error) {
	allargs := append(args, ipc.SocketName())

	cmd := exec.CommandContext(ctx, name, allargs...)

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return &coprocess{ipc, cmd}, nil
}

func (c *coprocess) Call(ctx context.Context, request *app.CoprocessRequest) (*app.CoprocessResponse, error) {
	return c.ipc.Call(ctx, request)
}

func (c *coprocess) Wait() error {
	return c.cmd.Wait()
}
