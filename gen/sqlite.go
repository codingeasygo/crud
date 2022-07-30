package gen

import (
	"reflect"

	"github.com/codingeasygo/util/xsql"
)

const TableSQLSQLITE = `
select name,type,'' from sqlite_master
where type='table' and name <> 'sqlite_sequence'
order by name asc
`

const ColumnSQLSQLITE = `
select name,type,pk,"notnull",dflt_value,cid,type,'' from pragma_table_info($1)
`

var TypeMapSQLITE = map[string][]string{
	//int
	"integer": {"int64", "*int64"},
	"int8":    {"int64", "*int64"},
	"bigint":  {"int64", "*int64"},
	"int":     {"int", "*int"},
	"int2":    {"int", "*int"},
	"int4":    {"int", "*int"},
	//float
	"real":             {"float32", "*float32"},
	"numeric":          {"float64", "*float64"},
	"double":           {"float64", "*float64"},
	"double precision": {"float64", "*float64"},
	//string
	"character":         {"string", "*string"},
	"varchar":           {"string", "*string"},
	"varying character": {"string", "*string"},
	"text":              {"string", "*string"},
	"clob":              {"string", "*string"},
	//time
	"date":     {"xsql.Time", "xsql.Time"},
	"datetime": {"xsql.Time", "xsql.Time"},
	//bool
	"boolean": {"bool", "*bool"},
}

var CodeSliceSQLITE = map[string]string{
	"RowLock": "",
}

func NameConvSQLITE(on, name string, field reflect.StructField) string {
	return name
}

func ParmConvSQLITE(on, fieldName, fieldFunc string, field reflect.StructField, value interface{}) interface{} {
	if c, ok := value.(xsql.ArrayConverter); on == "where" && ok {
		return c.InArray()
	}
	return value
}
