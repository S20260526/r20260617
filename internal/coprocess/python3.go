package coprocess

import (
	"context"
	"internal/app"
)

func NewPython3(ctx context.Context, path string, ipc app.Ipc) (*app.Coprocess, error) {
	return app.NewCoprocess(ctx, ipc, "python3", path)
}
