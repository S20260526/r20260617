package app

import (
	"context"
	"time"
)

type Event struct {
	T  time.Time
	Id string
}

type Registrator interface {
	Put(ctx context.Context, table string, event Event) error
}
