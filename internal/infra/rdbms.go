package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"internal/app"
	"regexp"
)

type RDBMS struct {
	db *sql.DB
}

func NewRDBMS(ctx context.Context, drivername, dns string) (*RDBMS, error) {
	db, err := sql.Open(drivername, dns)

	if err != nil {
		return nil, err
	}

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return &RDBMS{db: db}, nil
}

var tableNameRe = regexp.MustCompile(`^[[:lower:]][[:lower:][:digit:]_]*$`)
var TableNameIsInvalid = errors.New("table name go invalid symbols")

func (p *RDBMS) Put(ctx context.Context, table string, e app.Event) error {
	if !tableNameRe.MatchString(table) {
		return TableNameIsInvalid
	}

	_, err := p.db.ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO %s (t, id) VALUES ($1, $2)", table),
		e.T, e.Id,
	)

	return err
}
