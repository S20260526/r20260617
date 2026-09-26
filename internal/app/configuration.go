package app

import (
	"context"
	"fmt"
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
	ScriptDir   string
	Registrator RegistratorConfiguration
}

type Configurator interface {
	Watch(context.Context) (Configuration, <-chan Configuration, error)
}
