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

type configItemMap map[string]configItem

type configuration struct {
	config  app.Configuration
	itemMap configItemMap
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

func updateConfig(key []byte, value []byte, cfgIM configItemMap) error {
	skey := string(key)

	if !strings.HasPrefix(skey, prefix) {
		return nil
	}

	it := cfgIM[strings.TrimPrefix(skey, prefix)]

	if it != nil {
		return it.fromString(string(value))
	}

	return nil
}

func processEtcdEvents(events []*clientv3.Event, cfgIM configItemMap) {
	for _, ev := range events {
		if ev.Type == clientv3.EventTypePut {
			err := updateConfig(ev.Kv.Key, ev.Kv.Value, cfgIM)

			if err != nil {
				slog.Warn(
					"infra",
					"where", "config",
					"when", string(ev.Kv.Key),
				)
			}
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

	var cfg configuration

	cfg.config = defaultConfig

	c := &cfg.config

	cfg.itemMap = configItemMap{
		"input.port":         portCI{&c.InputPort},
		"storage":            hostColonPortCI{&c.Storage},
		"pushing.host":       hostColonPortCI{&c.Pushing.Host},
		"pushing.queue":      stringCI{&c.Pushing.Queue},
		"pulling.host":       hostColonPortArrayCI{&c.Pulling.Host},
		"pulling.queue":      stringCI{&c.Pulling.Queue},
		"script.dir":         stringCI{&c.ScriptDir},
		"registrator.driver": stringCI{&c.Registrator.Driver},
		"registrator.dsn":    stringCI{&c.Registrator.Dsn},
	}

	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())

	if err != nil {
		return cfg.config, nil, err
	}

	for _, kv := range resp.Kvs {
		updateConfig(kv.Key, kv.Value, cfg.itemMap)
	}

	wchan := e.client.Watch(
		clientv3.WithRequireLeader(ctx), "root.",
		clientv3.WithPrefix(),
	)

	ochan := make(chan app.Configuration)

	go func(cfg configuration) {
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-wchan:
				if !ok {
					return
				}

				processEtcdEvents(resp.Events, cfg.itemMap)

				ochan <- cfg.config
			}
		}
	}(cfg)

	return cfg.config, ochan, nil
}
