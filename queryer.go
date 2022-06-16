package crud

type Scanner interface {
	Scan(v interface{})
}

type Rows interface {
	Scan(dest ...interface{}) (err error)
	Next() bool
	Close()
}

type Row interface {
	Scan(dest ...interface{}) (err error)
}

type Queryer interface {
	Exec(sql string, args ...interface{}) (affected int64, err error)
	ExecRow(sql string, args ...interface{}) (err error)
	Query(sql string, args ...interface{}) (rows Rows, err error)
	QueryRow(sql string, args ...interface{}) (row Row)
}

type CrudQueryer interface {
	CrudExec(sql string, args ...interface{}) (affected int64, err error)
	CrudExecRow(sql string, args ...interface{}) (err error)
	CrudQuery(sql string, args ...interface{}) (rows Rows, err error)
	CrudQueryRow(sql string, args ...interface{}) (row Row)
}
