package sqlx

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codingeasygo/crud"
	"github.com/codingeasygo/crud/gen"
	"github.com/codingeasygo/crud/testsql"
	"github.com/codingeasygo/util/xsql"

	_ "github.com/lib/pq"
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

func init() {
	_, err := Bootstrap("postgres", "postgresql://dev:123@psql.loc:5432/crud?sslmode=disable")
	if err != nil {
		panic(err)
	}
	_, _, err = Pool().Exec(context.Background(), testsql.PG_DROP)
	if err != nil {
		panic(err)
	}
	_, _, err = Pool().Exec(context.Background(), testsql.PG_LATEST)
	if err != nil {
		panic(err)
	}
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
	Queryer:      func() interface{} { return Pool() },
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

func TestQueryer(t *testing.T) {
	rows, err := Pool().Query(context.Background(), "select 1")
	if err != nil {
		t.Error(err)
		return
	}
	rows.Close()
	Pool().QueryRow(context.Background(), "select 1")

	_, err = Pool().ExecRow(context.Background(), "update crud_object set status=0 where 1=0")
	if err != crud.ErrNoRows {
		t.Error(err)
		return
	}

	tx, err := Pool().Begin(context.Background())
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
	tx.QueryRow(context.Background(), "select 1")
	Pool().ErrNoRows = nil
	Pool().getErrNoRows()
	tx.ErrNoRows = nil
	tx.getErrNoRows()
}
