package pgx

import (
	"context"

	"github.com/codingeasygo/crud"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Shared *PgQueryer

var Pool = func() *PgQueryer {
	return Shared
}

func Bootstrap(connString string) (pool *pgxpool.Pool, err error) {
	pool, err = pgxpool.Connect(context.Background(), connString)
	if err == nil {
		Shared = NewPgQueryer(pool)
	}
	return
}

var ErrNoRows = pgx.ErrNoRows
var ErrTxClosed = pgx.ErrTxClosed
var ErrTxCommitRollback = pgx.ErrTxCommitRollback

type Row struct {
	SQL string
	pgx.Row
}

func (r Row) Scan(dest ...interface{}) (err error) {
	defer func() {
		xerr := r.Row.Scan(dest...)
		if err == nil {
			err = xerr
		}
	}()
	err = mockerCheck("Row.Scan", r.SQL)
	return
}

type Rows struct {
	SQL string
	pgx.Rows
}

func (r *Rows) Scan(dest ...interface{}) error {
	if err := mockerCheck("Rows.Scan", r.SQL); err != nil {
		return err
	}
	return r.Rows.Scan(dest...)
}

func (r *Rows) Values() ([]interface{}, error) {
	if err := mockerCheck("Rows.Values", r.SQL); err != nil {
		return nil, err
	}
	return r.Rows.Values()
}

func (r *Rows) Close() (err error) {
	r.Rows.Close()
	return
}

type BatchResults struct {
	pgx.BatchResults
}

func (b *BatchResults) Exec() (pgconn.CommandTag, error) {
	if err := mockerCheck("BatchResult.Exec", ""); err != nil {
		return nil, err
	}
	return b.BatchResults.Exec()
}

func (b *BatchResults) Query() (rows *Rows, err error) {
	if err := mockerCheck("BatchResult.Query", ""); err != nil {
		return nil, err
	}
	raw, err := b.BatchResults.Query()
	if err == nil {
		rows = &Rows{Rows: raw}
	}
	return
}

func (b *BatchResults) QueryRow() *Row {
	return &Row{Row: b.BatchResults.QueryRow()}
}

func (b *BatchResults) Close() error {
	if err := mockerCheck("BatchResult.Close", ""); err != nil {
		return err
	}
	return b.BatchResults.Close()
}

type Tx struct {
	pgx.Tx
}

// Begin starts a pseudo nested transaction.
func (t *Tx) Begin(ctx context.Context) (tx *Tx, err error) {
	if err := mockerCheck("Tx.Begin", ""); err != nil {
		return nil, err
	}
	raw, err := t.Tx.Begin(ctx)
	if err == nil {
		tx = &Tx{Tx: raw}
	}
	return
}

func (t *Tx) Commit(ctx context.Context) error {
	if err := mockerCheck("Tx.Commit", ""); err != nil {
		t.Tx.Rollback(ctx)
		return err
	}
	return t.Tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	if err := mockerCheck("Tx.Rollback", ""); err != nil {
		t.Tx.Rollback(ctx)
		return err
	}
	return t.Tx.Rollback(ctx)
}

func (t *Tx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if err := mockerCheck("Tx.CopyFrom", ""); err != nil {
		return 0, err
	}
	return t.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (t *Tx) SendBatch(ctx context.Context, b *pgx.Batch) *BatchResults {
	return &BatchResults{
		BatchResults: t.Tx.SendBatch(ctx, b),
	}
}

func (t *Tx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	if err := mockerCheck("Tx.Prepare", sql); err != nil {
		return nil, err
	}
	return t.Tx.Prepare(ctx, name, sql)
}

func (t *Tx) Exec(ctx context.Context, sql string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Tx.Exec", sql); err != nil {
		return 0, 0, err
	}
	res, err := t.Tx.Exec(ctx, sql, args...)
	if err == nil {
		affected = res.RowsAffected()
	}
	return
}

func (t *Tx) ExecRow(ctx context.Context, sql string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Tx.ExecRow", sql); err != nil {
		return 0, err
	}
	insertId, affected, err := t.Exec(ctx, sql, args...)
	if err == nil && affected < 1 {
		err = pgx.ErrNoRows
	}
	return
}

func (t *Tx) Query(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Tx.Query", sql); err != nil {
		return nil, err
	}
	raw, err := t.Tx.Query(ctx, sql, args...)
	if err == nil {
		rows = &Rows{SQL: sql, Rows: raw}
	}
	return
}

func (t *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) crud.Row {
	return &Row{
		SQL: sql,
		Row: t.Tx.QueryRow(ctx, sql, args...),
	}
}

func (t *Tx) CrudExec(ctx context.Context, sql string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Tx.Exec", sql); err != nil {
		return 0, 0, err
	}
	insertId, affected, err = t.Exec(ctx, sql, args...)
	return
}

func (t *Tx) CrudExecRow(ctx context.Context, sql string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Tx.ExecRow", sql); err != nil {
		return 0, err
	}
	insertId, err = t.ExecRow(ctx, sql, args...)
	return
}

func (t *Tx) CrudQuery(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Tx.Query", sql); err != nil {
		return nil, err
	}
	rows, err = t.Query(ctx, sql, args...)
	return
}

func (t *Tx) CrudQueryRow(ctx context.Context, sql string, args ...interface{}) (row crud.Row) {
	row = t.QueryRow(ctx, sql, args...)
	return
}

type PgQueryer struct {
	*pgxpool.Pool
}

func NewPgQueryer(pool *pgxpool.Pool) (queryer *PgQueryer) {
	queryer = &PgQueryer{Pool: pool}
	return
}

func (p *PgQueryer) Exec(ctx context.Context, sql string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Pool.Exec", sql); err != nil {
		return 0, 0, err
	}
	res, err := p.Pool.Exec(ctx, sql, args...)
	if err == nil {
		affected = res.RowsAffected()
	}
	return
}

func (p *PgQueryer) ExecRow(ctx context.Context, sql string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Pool.ExecRow", sql); err != nil {
		return 0, err
	}
	insertId, affected, err := p.Exec(ctx, sql, args...)
	if err == nil && affected < 1 {
		err = pgx.ErrNoRows
	}
	return
}

func (p *PgQueryer) Query(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Pool.Query", sql); err != nil {
		return nil, err
	}
	raw, err := p.Pool.Query(ctx, sql, args...)
	if err == nil {
		rows = &Rows{SQL: sql, Rows: raw}
	}
	return
}

func (p *PgQueryer) QueryRow(ctx context.Context, sql string, args ...interface{}) crud.Row {
	return &Row{
		SQL: sql,
		Row: p.Pool.QueryRow(ctx, sql, args...),
	}
}

func (p *PgQueryer) CrudExec(ctx context.Context, sql string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Pool.Exec", sql); err != nil {
		return 0, 0, err
	}
	insertId, affected, err = p.Exec(ctx, sql, args...)
	return
}

func (p *PgQueryer) CrudExecRow(ctx context.Context, sql string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Pool.ExecRow", sql); err != nil {
		return 0, err
	}
	insertId, err = p.ExecRow(ctx, sql, args...)
	return
}

func (p *PgQueryer) CrudQuery(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Pool.Query", sql); err != nil {
		return nil, err
	}
	rows, err = p.Query(ctx, sql, args...)
	return
}

func (p *PgQueryer) CrudQueryRow(ctx context.Context, sql string, args ...interface{}) (row crud.Row) {
	row = p.QueryRow(ctx, sql, args...)
	return
}

func (p *PgQueryer) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if err := mockerCheck("Pool.CopyFrom", ""); err != nil {
		return 0, err
	}
	return p.Pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *PgQueryer) SendBatch(ctx context.Context, b *pgx.Batch) *BatchResults {
	return &BatchResults{
		BatchResults: p.Pool.SendBatch(ctx, b),
	}
}

func (p *PgQueryer) Begin(ctx context.Context) (tx *Tx, err error) {
	if err := mockerCheck("Pool.Begin", ""); err != nil {
		return nil, err
	}
	raw, err := p.Pool.Begin(ctx)
	if err == nil {
		tx = &Tx{Tx: raw}
	}
	return
}

func Exec(ctx context.Context, sql string, args ...interface{}) (insertId, affected int64, err error) {
	return Shared.Exec(ctx, sql, args...)
}

func ExecRow(ctx context.Context, sql string, args ...interface{}) (insertId int64, err error) {
	return Shared.ExecRow(ctx, sql, args...)
}

func QueryRow(ctx context.Context, sql string, args ...interface{}) crud.Row {
	return Shared.QueryRow(ctx, sql, args...)
}

func Query(ctx context.Context, sql string, args ...interface{}) (rows crud.Rows, err error) {
	return Shared.Query(ctx, sql, args...)
}

func Begin(ctx context.Context) (tx *Tx, err error) {
	return Shared.Begin(ctx)
}
