package gen

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codingeasygo/crud/sqlx"
	"github.com/codingeasygo/crud/testsql"
	"github.com/codingeasygo/util/xsql"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func nameConv(isTable bool, name string) string {
	if isTable {
		return ConvCamelCase(true, strings.TrimPrefix(name, "emall_"))
	}
	if name == "tid" || name == "uuid" || name == "i18n" || name == "qq" {
		return strings.ToUpper(name)
	} else if strings.HasSuffix(name, "_id") {
		return ConvCamelCase(false, strings.TrimSuffix(name, "_id")+"_ID")
	} else if strings.HasSuffix(name, "_ids") {
		return ConvCamelCase(false, strings.TrimSuffix(name, "_ids")+"_IDs")
	} else {
		return ConvCamelCase(false, name)
	}
}

var sharedPG *sqlx.DbQueryer

func getPG() *sqlx.DbQueryer {
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
	sharedPG = sqlx.NewDbQueryer(db)
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

var PgInit = `
package autogen

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/codingeasygo/crud"
	"github.com/codingeasygo/crud/gen"
	"github.com/codingeasygo/crud/sqlx"
	"github.com/codingeasygo/crud/testsql"

	_ "github.com/lib/pq"
)

func init() {
	db, err := sql.Open("postgres", "postgresql://dev:123@psql.loc:5432/crud?sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	sharedPG := sqlx.NewDbQueryer(db)
	_, _, err = sharedPG.Exec(context.Background(), testsql.PG_DROP)
	if err != nil {
		panic(err)
	}
	_, _, err = sharedPG.Exec(context.Background(), testsql.PG_LATEST)
	if err != nil {
		panic(err)
	}
	func() {
		defer func() {
			recover()
		}()
		reflect.ValueOf(GetQueryer).Call(nil)
	}()
	GetQueryer = func() crud.Queryer { return sharedPG }
	crud.Default.Verbose = true
	crud.Default.NameConv = gen.NameConvPG
	crud.Default.ParmConv = gen.ParmConvPG
}
`

var PgGen = AutoGen{
	TypeField: map[string]map[string]string{
		"crud_object": {
			"int_array":     "xsql.IntArray",
			"int64_array":   "xsql.Int64Array",
			"float64_array": "xsql.Float64Array",
			"string_array":  "xsql.StringArray",
			"map_array":     "xsql.MArray",
		},
	},
	FieldFilter: map[string]map[string]string{
		"crud_object": {
			FieldsOrder: "type,update_time,create_time",
		},
	},
	CodeAddInit: map[string]string{
		"crud_object": `
			if ARG.Level < 1 {
				ARG.Level = 1
			}
		`,
	},
	CodeSlice: CodeSlicePG,
	TableGenAdd: xsql.StringArray{
		"crud_object",
	},
	TableInclude: xsql.StringArray{},
	TableExclude: xsql.StringArray{},
	Queryer:      getPG,
	TableSQL:     TableSQLPG,
	ColumnSQL:    ColumnSQLPG,
	Schema:       "public",
	TypeMap:      TypeMapPG,
	NameConv:     nameConv,
	GetQueryer:   "GetQueryer",
	Out:          "./autogen/",
	OutPackage:   "autogen",
}

func TestPgGen(t *testing.T) {
	var err error
	defer func() {
		if err == nil {
			os.RemoveAll(PgGen.Out)
		}
	}()
	os.MkdirAll(PgGen.Out, os.ModePerm)
	ioutil.WriteFile(filepath.Join(PgGen.Out, "auto_test.go"), []byte(PgInit), os.ModePerm)
	err = PgGen.Generate()
	if err != nil {
		t.Error(err)
		return
	}
	pwd, _ := os.Getwd()
	tester := exec.Command("go", "test", "-v")
	tester.Dir = filepath.Join(pwd, "autogen")
	tester.Stderr = os.Stderr
	tester.Stdout = os.Stdout
	err = tester.Run()
	if err != nil {
		t.Error(err)
		return
	}
}

var sharedSQLITE *sqlx.DbQueryer

func getSQLITE() *sqlx.DbQueryer {
	if sharedSQLITE != nil {
		return sharedSQLITE
	}
	db, err := sql.Open("sqlite3", "file:crud.sqlite?cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	sharedSQLITE = sqlx.NewDbQueryer(db)
	_, _, err = sharedSQLITE.Exec(context.Background(), testsql.SQLITE_DROP)
	if err != nil {
		panic(err)
	}
	_, _, err = sharedSQLITE.Exec(context.Background(), testsql.SQLITE_LATEST)
	if err != nil {
		panic(err)
	}
	return sharedSQLITE
}

const SqliteInit = `
package autogen

import (
	"context"
	"reflect"
	"database/sql"

	"github.com/codingeasygo/crud"
	"github.com/codingeasygo/crud/gen"
	"github.com/codingeasygo/crud/sqlx"
	"github.com/codingeasygo/crud/testsql"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	db, err := sql.Open("sqlite3", "file:crud.sqlite?cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	sharedSQLITE := sqlx.NewDbQueryer(db)
	_, _, err = sharedSQLITE.Exec(context.Background(), testsql.SQLITE_DROP)
	if err != nil {
		panic(err)
	}
	_, _, err = sharedSQLITE.Exec(context.Background(), testsql.SQLITE_LATEST)
	if err != nil {
		panic(err)
	}
	func() {
		defer func() {
			recover()
		}()
		reflect.ValueOf(GetQueryer).Call(nil)
	}()
	GetQueryer = func() crud.Queryer { return sharedSQLITE }
	crud.Default.Verbose = true
	crud.Default.NameConv = gen.NameConvSQLITE
	crud.Default.ParmConv = gen.ParmConvSQLITE
}
`

var SqliteGen = AutoGen{
	TypeField: map[string]map[string]string{
		"crud_object": {
			"int_array":     "xsql.IntArray",
			"int64_array":   "xsql.Int64Array",
			"float64_value": "decimal.Decimal",
			"float64_ptr":   "decimal.Decimal",
			"float64_array": "xsql.Float64Array",
			"string_array":  "xsql.StringArray",
			"map_array":     "xsql.MArray",
		},
	},
	FieldFilter: map[string]map[string]string{
		"crud_object": {
			FieldsOrder: "type,update_time,create_time",
		},
	},
	CodeAddInit: map[string]string{
		"crud_object": `
			if ARG.Level < 1 {
				ARG.Level = 1
			}
		`,
	},
	CodeSlice: CodeSliceSQLITE,
	Comments: map[string]map[string]string{
		"crud_object": {
			"type":   `simple type in, A=1:test a, B=2:test b, C=3:test c`,
			"status": `simple status in, Normal=100, Disabled=200, Removed=-1`,
		},
	},
	TableGenAdd: xsql.StringArray{
		"crud_object",
	},
	TableInclude: xsql.StringArray{},
	TableExclude: xsql.StringArray{},
	Queryer:      getSQLITE,
	TableSQL:     TableSQLSQLITE,
	ColumnSQL:    ColumnSQLSQLITE,
	Schema:       "",
	TypeMap:      TypeMapSQLITE,
	NameConv:     nameConv,
	GetQueryer:   "GetQueryer",
	Out:          "./autogen/",
	OutPackage:   "autogen",
}

func TestSqliteGen(t *testing.T) {
	var err error
	defer func() {
		if err == nil {
			os.RemoveAll(PgGen.Out)
		}
	}()
	os.MkdirAll(PgGen.Out, os.ModePerm)
	ioutil.WriteFile(filepath.Join(SqliteGen.Out, "auto_test.go"), []byte(SqliteInit), os.ModePerm)
	err = SqliteGen.Generate()
	if err != nil {
		t.Error(err)
		return
	}
	pwd, _ := os.Getwd()
	tester := exec.Command("go", "test", "-v")
	tester.Dir = filepath.Join(pwd, "autogen")
	tester.Stderr = os.Stderr
	tester.Stdout = os.Stdout
	err = tester.Run()
	if err != nil {
		t.Error(err)
		return
	}
}
