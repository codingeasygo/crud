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
	Query(sql string, args ...interface{}) (rows Rows, err error)
}

type CrudQueryer interface {
	CrudQuery(sql string, args ...interface{}) (rows Rows, err error)
}
