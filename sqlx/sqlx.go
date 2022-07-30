package sqlx

import (
	"context"
	"database/sql"

	"github.com/codingeasygo/crud"
)

var Shared *DbQueryer

var Pool = func() *DbQueryer {
	return Shared
}

func Bootstrap(driverName, dataSourceName string) (db *sql.DB, err error) {
	db, err = sql.Open(driverName, dataSourceName)
	if err == nil {
		Shared = NewDbQueryer(db)
	}
	return
}

type TxQueryer struct {
	*sql.Tx
	ErrNoRows   error
	NotAffected bool
}

func NewTxQueryer(tx *sql.Tx) (queryer *TxQueryer) {
	queryer = &TxQueryer{Tx: tx, ErrNoRows: crud.ErrNoRows}
	return
}

func (t *TxQueryer) getErrNoRows() (err error) {
	if t.ErrNoRows == nil {
		err = crud.ErrNoRows
	} else {
		err = t.ErrNoRows
	}
	return
}

func (t *TxQueryer) Exec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	res, err := t.Tx.ExecContext(ctx, query, args...)
	if err == nil {
		insertId, _ = res.LastInsertId() //ignore error for some driver is not supported
	}
	if err == nil {
		affected, err = res.RowsAffected()
	}
	if t.NotAffected && err == nil && affected < 1 {
		affected = 1
	}
	return
}

func (t *TxQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	insertId, affected, err := t.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = t.getErrNoRows()
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
	ErrNoRows   error
	NotAffected bool
}

func NewDbQueryer(db *sql.DB) (queryer *DbQueryer) {
	queryer = &DbQueryer{DB: db, ErrNoRows: crud.ErrNoRows}
	return
}

func (d *DbQueryer) getErrNoRows() (err error) {
	if d.ErrNoRows == nil {
		err = crud.ErrNoRows
	} else {
		err = d.ErrNoRows
	}
	return
}

func (d *DbQueryer) Begin(ctx context.Context) (tx *TxQueryer, err error) {
	raw, err := d.DB.BeginTx(ctx, nil)
	if err == nil {
		tx = NewTxQueryer(raw)
		tx.ErrNoRows = d.ErrNoRows
		tx.NotAffected = d.NotAffected
	}
	return
}

func (d *DbQueryer) Exec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	res, err := d.DB.ExecContext(ctx, query, args...)
	if err == nil {
		insertId, _ = res.LastInsertId() //ignore error for some driver is not supported
	}
	if err == nil {
		affected, err = res.RowsAffected()
	}
	if d.NotAffected && err == nil && affected < 1 {
		affected = 1
	}
	return
}

func (d *DbQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	insertId, affected, err := d.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = d.getErrNoRows()
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
