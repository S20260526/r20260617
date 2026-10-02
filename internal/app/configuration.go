package app

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"time"
)

type HostColonPort struct {
	Host string
	Port int
}

func (h HostColonPort) String() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

var schemeRe = regexp.MustCompile(`^[[:lower:]]+://$`)

func (h HostColonPort) Url(scheme string) string {
	if !schemeRe.MatchString(scheme) {
		slog.Warn(
			"app",
			"where", "config",
			"when", "HostPortUrl",
			"what", "invalid scheme",
		)

		return ""
	}

	return fmt.Sprintf("%s%s:%d", scheme, h.Host, h.Port)
}

func HostPortUrls(in []HostColonPort, scheme string) []string {
	var out []string

	if !schemeRe.MatchString(scheme) {
		slog.Warn(
			"app",
			"where", "config",
			"when", "HostPortUrl",
			"what", "invalid scheme",
		)

		return out
	}

	for _, it := range in {
		out = append(out, scheme+it.String())
	}

	return out
}

type PushingConfiguration struct {
	Host  HostColonPort
	Queue string
}

type PullingConfiguration struct {
	Host       []HostColonPort
	RetryTO    time.Duration
	Queue, DLQ string
}

type RegistratorConfiguration struct {
	Driver string
	Dsn    string
	Table  string
}

type Scripting struct {
	SocketDir string
	File      string
}

type Configuration struct {
	InputPort   int
	Storage     HostColonPort
	Pushing     PushingConfiguration
	Pulling     PullingConfiguration
	Scripting   Scripting
	Registrator RegistratorConfiguration
	MetricsPort int
}

func (c *Configuration) Clone() Configuration {
	out := *c

	out.Pulling.Host = slices.Clone(c.Pulling.Host)

	return out
}

type ConfigurationWatcher interface {
	Watch(context.Context) (Configuration, <-chan Configuration, error)
}
