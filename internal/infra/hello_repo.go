package infra

import (
	"context"
	"log/slog"

	"domain"

	vlk "github.com/valkey-io/valkey-go"
)

type HelloRepo struct {
	client vlk.Client
}

func NewHelloRepo() (*HelloRepo, error) {
	cli, e := vlk.NewClient(
		vlk.ClientOption{
			InitAddress: []string{
				"vlk-master:6379",
				"vlk-replica1:6379",
				"vlk-replica2:6379",
			},
			Sentinel: vlk.SentinelOption{
				MasterSet: "valkey-master",
			},
		},
	)

	return &HelloRepo{
		client: cli,
	}, e
}

func (r *HelloRepo) Done() {
	r.client.Close()
}

func (r *HelloRepo) Get() domain.Term {
	const key = "hello"

	gcmd := r.client.B().Get().Key(key).Build()

	rslt := r.client.Do(context.Background(), gcmd)

	if e := rslt.Error(); e != nil {
		if e == vlk.Nil {
			slog.Info("hello", "valkey", "cache missing")

			pcmd := r.client.B().Set().Key(key).Value("").Build()
			e = r.client.Do(context.Background(), pcmd).Error()
		}

		if e != nil {
			slog.Info("hello", "valkey error", e)
		}
	}

	return domain.NewHello()
}
