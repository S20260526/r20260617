package coprocess

import (
	"context"
	"os/exec"
)

type Ipc interface {
	SocketName() string
	Call(ctx context.Context, payload []byte) ([]byte, error)
}

type Coprocess struct {
	ipc Ipc
	cmd *exec.Cmd
}

func New(ctx context.Context, path string, ipc Ipc) (*Coprocess, error) {
	cmd := exec.CommandContext(ctx, "python3", path, ipc.SocketName())

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return &Coprocess{ipc, cmd}, nil
}

func (c *Coprocess) Call(ctx context.Context, payload []byte) ([]byte, error) {
	return c.ipc.Call(ctx, payload)
}

func (c *Coprocess) Wait() error {
	return c.cmd.Wait()
}
