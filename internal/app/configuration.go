package app

import (
	"context"
	"fmt"
	"slices"
)

type HostColonPort struct {
	Host string
	Port int
}

func (h HostColonPort) String() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type PushingConfiguration struct {
	Host  HostColonPort
	Queue string
}

type PullingConfiguration struct {
	Host  []HostColonPort
	Queue string
}

type RegistratorConfiguration struct {
	Driver string
	Dsn    string
}

type Configuration struct {
	InputPort   int
	Storage     HostColonPort
	Pushing     PushingConfiguration
	Pulling     PullingConfiguration
	ScriptFile  string
	Registrator RegistratorConfiguration
}

func (c *Configuration) Clone() Configuration {
	out := *c

	out.Pulling.Host = slices.Clone(c.Pulling.Host)

	return out
}

type ConfigurationWatcher interface {
	Watch(context.Context) (Configuration, <-chan Configuration, error)
}
