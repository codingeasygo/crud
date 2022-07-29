package pgx

import (
	"context"

	"github.com/codingeasygo/crud"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Shared *PgQueryer

var Pool = func() *PgQueryer {
	return Shared
}

func Bootstrap(connString string) (err error) {
	pool, err := pgxpool.Connect(context.Background(), connString)
	if err == nil {
		Shared = NewPgQueryer(pool)
	}
	return
}

type PgRows struct {
	pgx.Rows
}

func (p *PgRows) Close() (err error) {
	p.Rows.Close()
	return
}

type TxQueryer struct {
	pgx.Tx
}

func NewTxQueryer(tx pgx.Tx) (queryer *TxQueryer) {
	queryer = &TxQueryer{Tx: tx}
	return
}

func (t *TxQueryer) Exec(ctx context.Context, sql string, args ...interface{}) (affected int64, err error) {
	res, err := t.Tx.Exec(ctx, sql, args...)
	if err == nil {
		affected = res.RowsAffected()
	}
	return
}

func (t *TxQueryer) ExecRow(ctx context.Context, sql string, args ...interface{}) (err error) {
	affected, err := t.Exec(ctx, sql, args...)
	if err == nil && affected < 1 {
		err = pgx.ErrNoRows
	}
	return
}

func (t *TxQueryer) Query(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	raw, err := t.Tx.Query(ctx, sql, args...)
	if err == nil {
		rows = &PgRows{Rows: raw}
	}
	return
}

func (t *TxQueryer) QueryRow(ctx context.Context, sql string, args ...interface{}) (row crud.Row) {
	row = t.Tx.QueryRow(ctx, sql, args...)
	return
}

type PgQueryer struct {
	*pgxpool.Pool
}

func NewPgQueryer(pool *pgxpool.Pool) (queryer *PgQueryer) {
	queryer = &PgQueryer{Pool: pool}
	return
}

func (d *PgQueryer) Begin(ctx context.Context) (tx *TxQueryer, err error) {
	raw, err := d.Pool.Begin(ctx)
	if err == nil {
		tx = NewTxQueryer(raw)
	}
	return
}

func (d *PgQueryer) Exec(ctx context.Context, sql string, args ...interface{}) (affected int64, err error) {
	res, err := d.Pool.Exec(ctx, sql, args...)
	if err == nil {
		affected = res.RowsAffected()
	}
	return
}

func (d *PgQueryer) ExecRow(ctx context.Context, sql string, args ...interface{}) (err error) {
	affected, err := d.Exec(ctx, sql, args...)
	if err == nil && affected < 1 {
		err = pgx.ErrNoRows
	}
	return
}

func (d *PgQueryer) Query(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	raw, err := d.Pool.Query(ctx, sql, args...)
	if err == nil {
		rows = &PgRows{Rows: raw}
	}
	return
}

func (d *PgQueryer) QueryRow(ctx context.Context, sql string, args ...interface{}) (row crud.Row) {
	row = d.Pool.QueryRow(ctx, sql, args...)
	return
}
