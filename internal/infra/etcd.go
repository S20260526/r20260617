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
		Dsn: "host=localhost dbname=testdb " +
			"sslmode=disable user=postgres password=1234",
	},
}

const configKeyPrefix = "root."

type configItem interface {
	fromString(string) error
}

type configItemMap map[string]configItem

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

type configurationWatcher struct {
	config  app.Configuration
	itemMap configItemMap
}

func (etcd *Etcd) Watch(ctx context.Context) (app.Configuration, <-chan app.Configuration, error) {
	watcher := &configurationWatcher{
		config: defaultConfig,
	}

	c := &watcher.config

	watcher.itemMap = configItemMap{
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

	resp, err := etcd.client.Get(ctx, configKeyPrefix, clientv3.WithPrefix())

	if err != nil {
		return watcher.config.Clone(), nil, err
	}

	for _, kv := range resp.Kvs {
		updateConfig(kv.Key, kv.Value, watcher.itemMap)
	}

	wchan := etcd.client.Watch(
		clientv3.WithRequireLeader(ctx), "root.",
		clientv3.WithPrefix(),
	)

	ochan := make(chan app.Configuration)

	go func(w *configurationWatcher) {
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-wchan:
				if !ok {
					return
				}

				w.processEvents(resp.Events)

				ochan <- w.config.Clone()
			}
		}
	}(watcher)

	return watcher.config.Clone(), ochan, nil
}

func (watcher *configurationWatcher) processEvents(events []*clientv3.Event) {
	for _, ev := range events {
		if ev.Type == clientv3.EventTypePut {
			err := updateConfig(ev.Kv.Key, ev.Kv.Value, watcher.itemMap)

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

func updateConfig(key []byte, value []byte, cfgIM configItemMap) error {
	skey := string(key)

	if !strings.HasPrefix(skey, configKeyPrefix) {
		return nil
	}

	it := cfgIM[strings.TrimPrefix(skey, configKeyPrefix)]

	if it != nil {
		return it.fromString(string(value))
	}

	return nil
}

type stringCI struct {
	dst *string
}

func (ci stringCI) fromString(src string) error {
	*ci.dst = src

	return nil
}

var hostPortNotMatch = errors.New("host:port pattern not matched")

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
