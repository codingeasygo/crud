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

type Row struct {
	SQL string
	*sql.Row
}

func (r Row) Scan(dest ...interface{}) (err error) {
	defer func() {
		xerr := r.Row.Scan(dest...)
		if err == nil {
			err = xerr
		}
	}()
	err = mockerCheck("Rows.Scan", r.SQL)
	return
}

type Rows struct {
	SQL string
	*sql.Rows
}

func (r *Rows) Scan(dest ...interface{}) error {
	if err := mockerCheck("Rows.Scan", r.SQL); err != nil {
		return err
	}
	return r.Rows.Scan(dest...)
}

func (r *Rows) Close() (err error) {
	err = r.Rows.Close()
	return
}

type TxQueryer struct {
	*sql.Tx
	ErrNoRows error
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

func (t *TxQueryer) Commit() error {
	if err := mockerCheck("Tx.Commit", ""); err != nil {
		t.Tx.Rollback()
		return err
	}
	return t.Tx.Commit()
}

func (t *TxQueryer) Rollback() error {
	if err := mockerCheck("Tx.Rollback", ""); err != nil {
		t.Tx.Rollback()
		return err
	}
	return t.Tx.Rollback()
}

func (t *TxQueryer) Exec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Tx.Exec", ""); err != nil {
		return 0, 0, err
	}
	res, err := t.Tx.ExecContext(ctx, query, args...)
	if err == nil {
		insertId, _ = res.LastInsertId() //ignore error for some driver is not supported
	}
	if err == nil {
		affected, err = res.RowsAffected()
	}
	return
}

func (t *TxQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Tx.Exec", ""); err != nil {
		return 0, err
	}
	insertId, affected, err := t.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = t.getErrNoRows()
	}
	return
}

func (t *TxQueryer) Query(ctx context.Context, query string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Tx.Query", ""); err != nil {
		return nil, err
	}
	raw, err := t.Tx.QueryContext(ctx, query, args...)
	if err == nil {
		rows = &Rows{Rows: raw, SQL: query}
	}
	return
}

func (t *TxQueryer) QueryRow(ctx context.Context, query string, args ...interface{}) (row crud.Row) {
	raw := t.Tx.QueryRowContext(ctx, query, args...)
	row = &Row{Row: raw, SQL: query}
	return
}

type DbQueryer struct {
	*sql.DB
	ErrNoRows error
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
	if err := mockerCheck("Pool.Begin", ""); err != nil {
		return nil, err
	}
	raw, err := d.DB.BeginTx(ctx, nil)
	if err == nil {
		tx = NewTxQueryer(raw)
		tx.ErrNoRows = d.ErrNoRows
	}
	return
}

func (d *DbQueryer) Exec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	if err := mockerCheck("Pool.Exec", ""); err != nil {
		return 0, 0, err
	}
	res, err := d.DB.ExecContext(ctx, query, args...)
	if err == nil {
		insertId, _ = res.LastInsertId() //ignore error for some driver is not supported
	}
	if err == nil {
		affected, err = res.RowsAffected()
	}
	return
}

func (d *DbQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	if err := mockerCheck("Pool.Exec", ""); err != nil {
		return 0, err
	}
	insertId, affected, err := d.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = d.getErrNoRows()
	}
	return
}

func (d *DbQueryer) Query(ctx context.Context, query string, args ...interface{}) (rows crud.Rows, err error) {
	if err := mockerCheck("Pool.Query", ""); err != nil {
		return nil, err
	}
	raw, err := d.DB.QueryContext(ctx, query, args...)
	if err == nil {
		rows = &Rows{Rows: raw, SQL: query}
	}
	return
}

func (d *DbQueryer) QueryRow(ctx context.Context, query string, args ...interface{}) (row crud.Row) {
	raw := d.DB.QueryRowContext(ctx, query, args...)
	row = &Row{Row: raw, SQL: query}
	return
}
