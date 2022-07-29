package sqlx

import (
	"context"
	"database/sql"

	"github.com/codingeasygo/crud"
)

type TxQueryer struct {
	*sql.Tx
	ErrNoRows error
}

func NewTxQueryer(tx *sql.Tx) (queryer *TxQueryer) {
	queryer = &TxQueryer{Tx: tx, ErrNoRows: crud.ErrNoRows}
	return
}

func (t *TxQueryer) Exec(ctx context.Context, query string, args ...interface{}) (affected int64, err error) {
	res, err := t.Tx.ExecContext(ctx, query, args...)
	if err == nil {
		affected, err = res.RowsAffected()
	}
	return
}

func (t *TxQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (err error) {
	affected, err := t.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = t.ErrNoRows
	}
	return
}

func (t *TxQueryer) Query(ctx context.Context, query string, args ...interface{}) (rows crud.Rows, err error) {
	rows, err = t.Tx.QueryContext(ctx, query, args...)
	return
}

func (t *TxQueryer) QueryRow(ctx context.Context, query string, args ...interface{}) (row crud.Row) {
	row = t.Tx.QueryRowContext(ctx, query, args...)
	return
}

type DbQueryer struct {
	*sql.DB
	ErrNoRows error
}

func NewDbQueryer(db *sql.DB) (queryer *DbQueryer) {
	queryer = &DbQueryer{DB: db, ErrNoRows: crud.ErrNoRows}
	db.Begin()
	return
}

func (d *DbQueryer) Begin(ctx context.Context) (tx *TxQueryer, err error) {
	raw, err := d.DB.BeginTx(ctx, nil)
	if err == nil {
		tx = NewTxQueryer(raw)
	}
	return
}

func (d *DbQueryer) Exec(ctx context.Context, query string, args ...interface{}) (affected int64, err error) {
	res, err := d.DB.ExecContext(ctx, query, args...)
	if err == nil {
		affected, err = res.RowsAffected()
	}
	return
}

func (d *DbQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (err error) {
	affected, err := d.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = d.ErrNoRows
	}
	return
}

func (d *DbQueryer) Query(ctx context.Context, query string, args ...interface{}) (rows crud.Rows, err error) {
	rows, err = d.DB.QueryContext(ctx, query, args...)
	return
}

func (d *DbQueryer) QueryRow(ctx context.Context, query string, args ...interface{}) (row crud.Row) {
	row = d.DB.QueryRowContext(ctx, query, args...)
	return
}
