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
	"time"
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
		Queue:     "working",
		Heartbeat: time.Second * 30,
	},
	Pulling: app.PullingConfiguration{
		Host: []app.HostColonPort{
			app.HostColonPort{
				Host: "localhost", Port: 5672,
			},
		},
		RetryTO: time.Second,
		Queue:   "working",
		DLQ:     "dlq",
	},
	Scripting: app.Scripting{
		SocketDir: "/tmp",
		File:      "/dev/null/main.py",
	},
	Registrator: app.RegistratorConfiguration{
		Driver: "postgres",
		Dsn: "host=localhost dbname=testdb " +
			"sslmode=disable user=postgres password=1234",
		Table: "events",
	},
	MetricsPort: 8099,
}

const configKeyPrefix = "root."

type configItem interface {
	fromString(string) error
}

type configItemMap map[string]configItem

type Etcd struct {
	client *clientv3.Client
}

func NewEtcd(url string) (*Etcd, error) {
	client, err := clientv3.NewFromURL(url)

	if err != nil {
		return nil, err
	}

	return &Etcd{client: client}, nil
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
		"pushing.heartbeat":  durationCI{&c.Pushing.Heartbeat},
		"pulling.host":       hostColonPortArrayCI{&c.Pulling.Host},
		"pulling.queue":      stringCI{&c.Pulling.Queue},
		"pulling.dlq":        stringCI{&c.Pulling.DLQ},
		"pulling.retry.t":    durationCI{&c.Pulling.RetryTO},
		"script.socket.d":    stringCI{&c.Scripting.SocketDir},
		"script.file":        stringCI{&c.Scripting.File},
		"registrator.driver": stringCI{&c.Registrator.Driver},
		"registrator.dsn":    stringCI{&c.Registrator.Dsn},
		"registrator.table":  stringCI{&c.Registrator.Table},
		"metrics.port":       portCI{&c.MetricsPort},
	}

	getCtx, getCtxCancel := context.WithTimeout(ctx, time.Second*5)

	defer getCtxCancel()

	resp, err := etcd.client.Get(getCtx, configKeyPrefix, clientv3.WithPrefix())

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
		defer close(ochan)

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

func (etcd *Etcd) Close() {
	err := etcd.client.Close()

	if err != nil {
		slog.Warn(
			"infra",
			"where", "config",
			"when", "close",
			"what", err,
		)
	}
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
					"what", err,
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

type durationCI struct {
	dst *time.Duration
}

func (ci durationCI) fromString(src string) error {
	dur, err := time.ParseDuration(src)

	if err != nil {
		return err
	}

	*ci.dst = dur

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
