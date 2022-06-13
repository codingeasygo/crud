package gen

// import (
// 	"context"
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/codingeasygo/crud"
// 	"github.com/jackc/pgx/v4/pgxpool"
// )

// var Pool func() *pgxpool.Pool

// func init() {
// 	pool, err := pgxpool.Connect(context.Background(), "postgresql://dev:123@psql.loc:5432/crud")
// 	if err != nil {
// 		panic(err)
// 	}
// 	Pool = func() *pgxpool.Pool {
// 		return pool
// 	}
// 	Pool().Exec(context.Background(), DROP)
// 	Pool().Exec(context.Background(), LATEST)
// }

// type PoolWrapper struct {
// }

// func NewPoolWrapper() (pool *PoolWrapper) {
// 	pool = &PoolWrapper{}
// 	return
// }

// func (c *PoolWrapper) Query(sql string, args ...interface{}) (rows crud.Rows, err error) {
// 	rows, err = Pool().Query(context.Background(), sql, args...)
// 	return
// }

// func TestGen(t *testing.T) {
// 	tables, err := Query(NewPoolWrapper(), SQLTablePG, SQLColumnPG, "public")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if len(tables) < 1 {
// 		t.Error("table is not found")
// 		return
// 	}
// 	// fmt.Println(JSON(tables))
// 	gen := NewGen(TypeMapPG, tables)
// 	gen.NameConv = func(isTable bool, name string) string {
// 		return ConvCamelCase(isTable, strings.TrimPrefix(name, "crud_"))
// 	}
// 	err = gen.GenerateByTemplate(StructTmpl, os.Stdout)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// }
