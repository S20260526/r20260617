package infra

import (
	"bufio"
	"context"
	"fmt"
	"internal/app"
	"log/slog"
	"os"
	"os/exec"
)

type coprocess struct {
	ipc app.Ipc
	cmd *exec.Cmd
}

func NewPython3(ctx context.Context, path string, ipc app.Ipc) (app.Coprocess, error) {
	_, err := os.Stat(path)

	if err != nil {
		return nil, fmt.Errorf("script file %s not found", path)
	}

	return newCoprocess(ctx, ipc, "python3", path)
}

func newCoprocess(ctx context.Context, ipc app.Ipc, name string, args ...string) (app.Coprocess, error) {
	allargs := append(args, ipc.SocketName())

	cmd := exec.CommandContext(ctx, name, allargs...)

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	go func() {
		scnr := bufio.NewScanner(stderr)

		for scnr.Scan() {

			s := scnr.Text()

			if len(s) > 0 {
				slog.Warn(
					"coprocess",
					"where", name,
					"when", "scan stderr",
					"what", s,
				)
			}
		}
	}()

	return &coprocess{ipc, cmd}, nil
}

func (c *coprocess) Call(ctx context.Context, request *app.CoprocessRequest) (*app.CoprocessResponse, error) {
	return c.ipc.Call(ctx, request)
}

func (c *coprocess) Wait() error {
	return c.cmd.Wait()
}
