package infra

import (
	"context"
	"errors"
	"go.etcd.io/etcd/client/v3"
	"internal/app"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type Etcd struct {
	client *clientv3.Client
}

func NewEtcd(url string) *Etcd {
	client, err := clientv3.NewFromURL(url)

	if err != nil {
		slog.Warn(
			"infra",
			"where", "config",
			"when", "create",
			"what", err,
		)

		return nil
	}

	return &Etcd{client: client}
}

type configItem interface {
	fromString(string) error
}

type stringCI struct {
	dst *string
}

func (ci stringCI) fromString(src string) error {
	*ci.dst = src

	return nil
}

type portCI struct {
	dst *int
}

func (ci portCI) fromString(src string) error {
	port, err := strconv.Atoi(src)

	if err != nil || port < 1 || port > 65535 {
		return hostPortNotMatch
	}

	*ci.dst = port

	return nil
}

type hostColonPortCI struct {
	dst *app.HostColonPort
}

var hostColonPortPattern = regexp.MustCompile(
	"^([[:alpha:]][[:alpha:][:digit:]._-]*):([[:digit:]]{1,5})$",
)

var hostPortNotMatch = errors.New("host:port pattern not matched")

func (ci hostColonPortCI) fromString(src string) error {
	match := hostColonPortPattern.FindStringSubmatch(src)

	if match == nil {
		return hostPortNotMatch
	}

	port, err := strconv.Atoi(match[2])

	if err != nil || port < 1 || port > 65535 {
		return hostPortNotMatch
	}

	ci.dst.Host = match[1]
	ci.dst.Port = port

	return nil
}

type hostColonPortArrayCI struct {
	dst *[]app.HostColonPort
}

func (ci hostColonPortArrayCI) fromString(src string) error {
	if src == "" {
		*ci.dst = []app.HostColonPort{}

		return nil
	}

	spltd := strings.Split(src, ",")

	buf := make([]app.HostColonPort, len(spltd))

	for i, s := range spltd {
		var tmpdst app.HostColonPort

		err := hostColonPortCI{&tmpdst}.fromString(s)

		if err != nil {
			return hostPortNotMatch
		}

		buf[i] = tmpdst
	}

	*ci.dst = buf

	return nil
}

const prefix = "root."

func updateConfig(key []byte, value []byte, items map[string]configItem) {
	skey := string(key)

	if !strings.HasPrefix(skey, prefix) {
		return
	}

	it := items[strings.TrimPrefix(skey, prefix)]

	if it != nil {
		err := it.fromString(string(value))

		if err != nil {
			slog.Warn(
				"infra",
				"where", "config",
				"what", skey,
			)
		}
	}
}

var defaultConfig = app.Configuration{
	InputPort: 8089,
	Storage: app.HostColonPort{
		Host: "localhost", Port: 9333,
	},
	Pushing: app.PushingConfiguration{
		Host: app.HostColonPort{
			Host: "localhost", Port: 5672,
		},
		Queue: "working",
	},
	Pulling: app.PullingConfiguration{
		Host: []app.HostColonPort{
			app.HostColonPort{
				Host: "localhost", Port: 5672,
			},
		},
		Queue: "working",
	},
	ScriptDir: ".",
	Registrator: app.RegistratorConfiguration{
		Driver: "postgres",
		Dsn:    "host=localhost dbname=testdb sslmode=disable user=postgres password=1234",
	},
}

func (e *Etcd) Watch(ctx context.Context) (app.Configuration, <-chan app.Configuration, error) {
	const prefix = "root."

	config := defaultConfig

	items := map[string]configItem{
		"input.port":         portCI{&config.InputPort},
		"storage":            hostColonPortCI{&config.Storage},
		"pushing.host":       hostColonPortCI{&config.Pushing.Host},
		"pushing.queue":      stringCI{&config.Pushing.Queue},
		"pulling.host":       hostColonPortArrayCI{&config.Pulling.Host},
		"pulling.queue":      stringCI{&config.Pulling.Queue},
		"script.dir":         stringCI{&config.ScriptDir},
		"registrator.driver": stringCI{&config.Registrator.Driver},
		"registrator.dsn":    stringCI{&config.Registrator.Dsn},
	}

	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())

	if err != nil {
		return app.Configuration{}, nil, err
	}

	for _, kv := range resp.Kvs {
		updateConfig(kv.Key, kv.Value, items)
	}

	wchan := e.client.Watch(
		clientv3.WithRequireLeader(ctx), "root.",
		clientv3.WithPrefix(),
	)

	ochan := make(chan app.Configuration)

	go func(cfg app.Configuration) {
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-wchan:
				if !ok {
					return
				}

				for _, ev := range resp.Events {
					if ev.Type == clientv3.EventTypePut {
						updateConfig(ev.Kv.Key, ev.Kv.Value, items)
					}
				}

				ochan <- config
			}
		}
	}(config)

	return config, ochan, nil
}
