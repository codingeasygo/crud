package crud

import (
	"context"
	"fmt"
)

var ErrNoRows = fmt.Errorf("no rows")

type Scanner interface {
	Scan(v interface{})
}

type Rows interface {
	Scan(dest ...interface{}) (err error)
	Next() bool
	Close() error
}

type Row interface {
	Scan(dest ...interface{}) (err error)
}

type Queryer interface {
	Exec(ctx context.Context, query string, args ...interface{}) (affected int64, err error)
	ExecRow(ctx context.Context, query string, args ...interface{}) (err error)
	Query(ctx context.Context, query string, args ...interface{}) (rows Rows, err error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (row Row)
}

type CrudQueryer interface {
	CrudExec(ctx context.Context, query string, args ...interface{}) (affected int64, err error)
	CrudExecRow(ctx context.Context, query string, args ...interface{}) (err error)
	CrudQuery(ctx context.Context, query string, args ...interface{}) (rows Rows, err error)
	CrudQueryRow(ctx context.Context, query string, args ...interface{}) (row Row)
}
