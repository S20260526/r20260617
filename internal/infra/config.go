package infra

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.etcd.io/etcd/client/v3"
)

var NoSuchKey = errors.New("key not found")

type ConfigGetter struct {
	client *clientv3.Client
}

var Config = func() *ConfigGetter {
	cli, e := clientv3.NewFromURL("http://config:2379")

	if e != nil {
		slog.Info("config", "error", e)

		return nil
	}

	return &ConfigGetter{client: cli}
}()

func (cg *ConfigGetter) get(key, def string) (string, error) {
	if cg == nil {
		return def, NoSuchKey
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	rsp, e := cg.client.Get(ctx, key)

	if e != nil {
		return def, e
	}

	if len(rsp.Kvs) == 0 {
		return def, NoSuchKey
	}

	return string(rsp.Kvs[0].Value), nil
}

func (cg *ConfigGetter) Get(key, def string) string {
	ret, e := cg.get(key, def)

	slog.Info("config", "op", "GET", "KEY", key, "VAL", ret, "ERR", e)

	return ret
}
