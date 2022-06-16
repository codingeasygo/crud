package crud

type Scanner interface {
	Scan(v interface{})
}

type Rows interface {
	Scan(dest ...interface{}) (err error)
	Next() bool
	Close()
}

type Queryer interface {
	Exec(sql string, args ...interface{}) (affected int64, err error)
	Query(sql string, args ...interface{}) (rows Rows, err error)
}

type CrudQueryer interface {
	CrudExec(sql string, args ...interface{}) (affected int64, err error)
	CrudQuery(sql string, args ...interface{}) (rows Rows, err error)
}
