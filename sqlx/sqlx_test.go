package sqlx

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

	"github.com/codingeasygo/crud"
	"github.com/codingeasygo/crud/gen"
	"github.com/codingeasygo/crud/testsql"
	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xsql"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func nameConv(isTable bool, name string) string {
	if isTable {
		return gen.ConvCamelCase(true, strings.TrimPrefix(name, "emall_"))
	}
	if name == "tid" || name == "uuid" || name == "i18n" || name == "qq" {
		return strings.ToUpper(name)
	} else if strings.HasSuffix(name, "_id") {
		return gen.ConvCamelCase(false, strings.TrimSuffix(name, "_id")+"_ID")
	} else if strings.HasSuffix(name, "_ids") {
		return gen.ConvCamelCase(false, strings.TrimSuffix(name, "_ids")+"_IDs")
	} else {
		return gen.ConvCamelCase(false, name)
	}
}

var sharedSQLITE *DbQueryer

func getSQLITE() *DbQueryer {
	if sharedSQLITE != nil {
		return sharedSQLITE
	}
	db, err := sql.Open("sqlite3", "file:crud.sqlite?cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	sharedSQLITE = NewDbQueryer(db)
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

var sharedPG *DbQueryer

func getPG() *DbQueryer {
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
	sharedPG = NewDbQueryer(db)
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

var PgGen = gen.AutoGen{
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
			gen.FieldsOrder: "type,update_time,create_time",
		},
	},
	CodeAddInit: map[string]string{
		"crud_object": `
			if ARG.Level < 1 {
				ARG.Level = 1
			}
		`,
	},
	CodeSlice: gen.CodeSlicePG,
	TableGenAdd: xsql.StringArray{
		"crud_object",
	},
	TableInclude: xsql.StringArray{},
	TableExclude: xsql.StringArray{},
	Queryer:      getPG,
	TableSQL:     gen.TableSQLPG,
	ColumnSQL:    gen.ColumnSQLPG,
	Schema:       "public",
	TypeMap:      gen.TypeMapPG,
	NameConv:     nameConv,
	Out:          "./autogen/",
	OutPackage:   "autogen",
}

func TestPgGen(t *testing.T) {
	defer os.RemoveAll(PgGen.Out)
	os.MkdirAll(PgGen.Out, os.ModePerm)
	ioutil.WriteFile(filepath.Join(PgGen.Out, "auto_test.go"), []byte(PgInit), os.ModePerm)
	err := PgGen.Generate()
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

func TestQueryerPG(t *testing.T) {
	{ //exec test
		var newID int64
		err := getPG().QueryRowContext(
			context.Background(),
			`insert into crud_object(title,time_value,update_time,create_time,status) values($1,$2,$3,$4,$5) returning tid`,
			"title", time.Now(), time.Now(), time.Now(), 100,
		).Scan(&newID)
		if err != nil || newID < 1 {
			t.Error(err)
			return
		}
		_, affected, err := getPG().Exec(context.Background(), `update crud_object set title=$1 where tid=$2`, "titl2", newID)
		if err != nil || affected < 1 {
			t.Error(err)
			return
		}
	}
	{
		rows, err := getPG().Query(context.Background(), "select 1")
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		getPG().QueryRow(context.Background(), "select 1").Scan(converter.Int(0))
		_, err = getPG().ExecRow(context.Background(), "update crud_object set status=0 where 1=0")
		if err != crud.ErrNoRows {
			t.Error(err)
			return
		}

		tx, err := getPG().Begin(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
		defer tx.Rollback()

		_, err = tx.ExecRow(context.Background(), "update crud_object set status=0 where 1=0")
		if err != crud.ErrNoRows {
			t.Error(err)
			return
		}
		rows, err = tx.Query(context.Background(), "select 1")
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		tx.QueryRow(context.Background(), "select 1").Scan(converter.Int(0))
		getPG().ErrNoRows = nil
		getPG().getErrNoRows()
		tx.ErrNoRows = nil
		tx.getErrNoRows()
	}
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

var SqliteGen = gen.AutoGen{
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
			gen.FieldsOrder: "type,update_time,create_time",
		},
	},
	CodeAddInit: map[string]string{
		"crud_object": `
			if ARG.Level < 1 {
				ARG.Level = 1
			}
		`,
	},
	CodeSlice: gen.CodeSliceSQLITE,
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
	TableSQL:     gen.TableSQLSQLITE,
	ColumnSQL:    gen.ColumnSQLSQLITE,
	Schema:       "",
	TypeMap:      gen.TypeMapSQLITE,
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

func TestQueryerSQLITE(t *testing.T) {
	{ //exec test
		var newID int64
		err := getSQLITE().QueryRowContext(
			context.Background(),
			`insert into crud_object(title,time_value,update_time,create_time,status) values($1,$2,$3,$4,$5) returning tid`,
			"title", time.Now(), time.Now(), time.Now(), 100,
		).Scan(&newID)
		if err != nil || newID < 1 {
			t.Error(err)
			return
		}
		_, affected, err := getSQLITE().Exec(context.Background(), `update crud_object set title=$1 where tid=$2`, "titl2", newID)
		if err != nil || affected < 1 {
			t.Error(err)
			return
		}
	}
	{ //normal test
		rows, err := getSQLITE().Query(context.Background(), "select 1")
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		getSQLITE().QueryRow(context.Background(), "select 1").Scan(converter.Int(0))

		_, err = getSQLITE().ExecRow(context.Background(), "update crud_object set status=0 where 1=0")
		if err != crud.ErrNoRows {
			t.Error(err)
			return
		}

		tx, err := getSQLITE().Begin(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
		defer tx.Rollback()

		_, err = tx.ExecRow(context.Background(), "update crud_object set status=0 where 1=0")
		if err != crud.ErrNoRows {
			t.Error(err)
			return
		}
		rows, err = tx.Query(context.Background(), "select 1")
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		tx.QueryRow(context.Background(), "select 1").Scan(converter.Int(0))
		getSQLITE().ErrNoRows = nil
		getSQLITE().getErrNoRows()
		tx.ErrNoRows = nil
		tx.getErrNoRows()
	}
}
