package crud

import (
	"context"
	"database/sql"
	"time"

	"github.com/codingeasygo/crud/testsql"

	_ "github.com/lib/pq"
)

var sharedPG *TestDbQueryer

func getPG() *TestDbQueryer {
	if sharedPG != nil {
		return sharedPG
	}
	db, err := sql.Open("postgres", "postgresql://dev:123@psql.loc:5432/crud?sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	sharedPG = NewTestDbQueryer(db)
	_, _, err = sharedPG.Exec(context.Background(), testsql.PG_DROP)
	if err != nil {
		panic(err)
	}
	_, _, err = sharedPG.Exec(context.Background(), testsql.PG_LATEST)
	if err != nil {
		panic(err)
	}
	return sharedPG
}

func clearPG() {
	queryer := getPG()
	queryer.Exec(context.Background(), testsql.PG_CLEAR)
}

type TestDbQueryer struct {
	*sql.DB
	ErrNoRows error
}

func NewTestDbQueryer(db *sql.DB) (queryer *TestDbQueryer) {
	queryer = &TestDbQueryer{DB: db, ErrNoRows: ErrNoRows}
	return
}

func (d *TestDbQueryer) getErrNoRows() (err error) {
	if d.ErrNoRows == nil {
		err = ErrNoRows
	} else {
		err = d.ErrNoRows
	}
	return
}

func (d *TestDbQueryer) Exec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	res, err := d.DB.ExecContext(ctx, query, args...)
	if err == nil {
		insertId, _ = res.LastInsertId() //ignore error for some driver is not supported
	}
	if err == nil {
		affected, err = res.RowsAffected()
	}
	return
}

func (d *TestDbQueryer) ExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	insertId, affected, err := d.Exec(ctx, query, args...)
	if err == nil && affected < 1 {
		err = d.getErrNoRows()
	}
	return
}

func (d *TestDbQueryer) Query(ctx context.Context, query string, args ...interface{}) (rows Rows, err error) {
	rows, err = d.DB.QueryContext(ctx, query, args...)
	return
}

func (d *TestDbQueryer) QueryRow(ctx context.Context, query string, args ...interface{}) (row Row) {
	row = d.DB.QueryRowContext(ctx, query, args...)
	return
}

type TestCrudQueryer struct {
	Queryer Queryer
}

func (t *TestCrudQueryer) CrudExec(ctx context.Context, query string, args ...interface{}) (insertId, affected int64, err error) {
	insertId, affected, err = t.Queryer.Exec(ctx, query, args...)
	return
}
func (t *TestCrudQueryer) CrudExecRow(ctx context.Context, query string, args ...interface{}) (insertId int64, err error) {
	insertId, err = t.Queryer.ExecRow(ctx, query, args...)
	return
}
func (t *TestCrudQueryer) CrudQuery(ctx context.Context, query string, args ...interface{}) (rows Rows, err error) {
	rows, err = t.Queryer.Query(ctx, query, args...)
	return
}
func (t *TestCrudQueryer) CrudQueryRow(ctx context.Context, query string, args ...interface{}) (row Row) {
	row = t.Queryer.QueryRow(ctx, query, args...)
	return
}
