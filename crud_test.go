package crud

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	Default.Verbose = true
}

type PoolQueryer struct {
	ExecErr      error
	ExecAffected int64
	QueryMode    string
	QueryErr     error
	ScanErr      error
}

func NewPoolQueryer() (pool *PoolQueryer) {
	pool = &PoolQueryer{
		ExecAffected: 1,
	}
	return
}

func (t *PoolQueryer) Exec(sql string, args ...interface{}) (affected int64, err error) {
	affected = t.ExecAffected
	err = t.ExecErr
	return
}

func (t *PoolQueryer) ExecRow(sql string, args ...interface{}) (err error) {
	err = t.ExecErr
	return
}

func (t *PoolQueryer) Query(sql string, args ...interface{}) (rows Rows, err error) {
	if sql == "no" || t.QueryMode == "no" {
		rows = &TestRows{
			Err:   t.ScanErr,
			Index: -1,
		}
		err = t.QueryErr
		return
	}
	if sql == "int64" || t.QueryMode == "int64" {
		rows = &TestRows{
			Err:   t.ScanErr,
			Index: -1,
			Values: [][]interface{}{
				{int64(1)},
				{int64(2)},
				{int64(3)},
			},
		}
		err = t.QueryErr
		return
	}
	title := "test"
	rows = &TestRows{
		Err:   t.ScanErr,
		Index: -1,
		Values: [][]interface{}{
			ScanArgs(&Simple{
				TID:        1,
				UserID:     100,
				Type:       "test",
				Title:      &title,
				Image:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, "#all"),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				Image:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, "#all"),
			ScanArgs(&Simple{
				TID:        2,
				UserID:     0,
				Type:       "test",
				Title:      &title,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     SimpleStatusNormal,
			}, "#all"),
		},
	}
	err = t.QueryErr
	return
}

func (t *PoolQueryer) QueryRow(sql string, args ...interface{}) (row Row) {
	rows, _ := t.Query(sql, args...)
	if !rows.Next() {
		rows.(*TestRows).Err = Default.ErrNoRows
	}
	row = rows
	return
}

type PoolCrudQueryer struct {
	Queryer *PoolQueryer
}

func NewPoolCrudQueryer() (pool *PoolCrudQueryer) {
	pool = &PoolCrudQueryer{
		Queryer: NewPoolQueryer(),
	}
	return
}

func (p *PoolCrudQueryer) CrudExec(sql string, args ...interface{}) (affected int64, err error) {
	affected, err = p.Queryer.Exec(sql, args...)
	return
}

func (p *PoolCrudQueryer) CrudExecRow(sql string, args ...interface{}) (err error) {
	err = p.Queryer.ExecRow(sql, args...)
	return
}

func (p *PoolCrudQueryer) CrudQuery(sql string, args ...interface{}) (rows Rows, err error) {
	rows, err = p.Queryer.Query(sql, args...)
	return
}

func (p *PoolCrudQueryer) CrudQueryRow(sql string, args ...interface{}) (row Row) {
	row = p.Queryer.QueryRow(sql, args...)
	return
}

type TestRows struct {
	Values [][]interface{}
	Index  int
	Err    error
}

func (t *TestRows) Scan(dests ...interface{}) (err error) {
	if len(dests) < 1 {
		err = fmt.Errorf("dest is empty")
		return
	}
	err = t.Err
	if t.Index < len(t.Values) {
		values := t.Values[t.Index]
		for i, dest := range dests {
			destValue := reflect.Indirect(reflect.ValueOf(dest))
			src := reflect.ValueOf(values[i])
			if destValue.Type() == src.Type() {
				destValue.Set(reflect.ValueOf(values[i]))
			} else {
				destValue.Set(reflect.Indirect(reflect.ValueOf(values[i])))
			}
		}
	}
	return
}

func (t *TestRows) Next() bool {
	t.Index++
	return t.Index < len(t.Values)
}

func (t *TestRows) Close() {

}

const (
	SimpleTypeMaterial  = "material"
	SimpleTypeProduct   = "product"
	SimpleStatusNormal  = 100
	SimpleStatusRemoved = -1
)

var (
	SimpleTypeAll   = []string{SimpleTypeMaterial, SimpleTypeProduct}
	SimpleStatusAll = []int{SimpleStatusNormal, SimpleStatusRemoved}
)

type Simple struct {
	_          string    `table:"crud_simple"`
	TID        int64     `json:"tid"`
	UserID     int64     `json:"user_id"`
	Type       string    `json:"type"`
	Title      *string   `json:"title"`
	Image      *string   `json:"image"`
	Data       string    `json:"data" conv:"::text"`
	UpdateTime time.Time `json:"update_time"`
	CreateTime time.Time `json:"create_time"`
	Status     int       `json:"status"`
	XX         string    `json:"-"`
}

func TestJsonString(t *testing.T) {
	jsonString(t)
	jsonString(t.Error)
}

func TestQueryer(t *testing.T) {
	var err error
	var row Row

	_, err = Default.queryerQuery(NewPoolQueryer(), "test", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.queryerQuery(NewPoolCrudQueryer(), "test", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Default.queryerQuery("xxx", "test", []interface{}{})
	}()

	row = Default.queryerQueryRow(NewPoolQueryer(), "test", []interface{}{})
	if row == nil {
		t.Error(err)
		return
	}
	row = Default.queryerQueryRow(NewPoolCrudQueryer(), "test", []interface{}{})
	if row == nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Default.queryerQueryRow("xxx", "test", []interface{}{})
	}()

	_, err = Default.queryerExec(NewPoolQueryer(), "test", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.queryerExec(NewPoolCrudQueryer(), "test", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Default.queryerExec("xxx", "test", []interface{}{})
	}()
}

func TestNewValue(t *testing.T) {
	{
		value := NewValue(&Simple{})
		if _, ok := value.Interface().(*Simple); !ok {
			t.Error("error")
			return
		}
		fmt.Println(value)
	}
	{
		src := []interface{}{string(""), int64(0)}
		value := NewValue(src).Interface().([]interface{})
		fmt.Println(src)
		fmt.Println(value)
		for i, s := range src {
			d := value[i]
			if reflect.TypeOf(s) != reflect.ValueOf(d).Elem().Type() {
				t.Error("error")
				return
			}
		}
		FilterFieldCall("test", value, "title,tid", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			fmt.Println(value)
		})
	}
	{
		var strVal string
		var intVal int64
		src := []interface{}{&strVal, &intVal}
		value := NewValue(src).Interface().([]interface{})
		fmt.Println(src)
		fmt.Println(value)
		for i, s := range src {
			d := value[i]
			if d == nil || reflect.TypeOf(s) != reflect.ValueOf(d).Elem().Type() {
				t.Error("error")
				return
			}
			fmt.Printf("%v,%v\n", d == nil, reflect.TypeOf(d))
		}
		FilterFieldCall("test", value, "title,tid", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			fmt.Printf("%v\n", reflect.TypeOf(value))
		})
	}
	{
		src := []interface{}{TableName("xx"), string(""), int64(0)}
		value := NewValue(src).Interface().([]interface{})
		if len(value)+1 != len(src) {
			t.Error("error")
			return
		}
	}
}

func TestFilterField(t *testing.T) {
	table := ""
	title := "test"
	simple := &Simple{
		UserID: 100,
		Type:   "test",
		Title:  &title,
		Image:  &title,
		Status: SimpleStatusNormal,
	}
	{
		table = Table(simple)
		if table != "crud_simple" {
			t.Error("error")
			return
		}
	}
	{
		v := MetaWith("crud_simple", int64(0))
		table = Table(v)
		if table != "crud_simple" {
			t.Error("error")
			return
		}
		table, fields := QueryField(v, "tid#all")
		if table != "crud_simple" || len(fields) != 1 {
			t.Error("error")
			return
		}
	}
	{
		v := MetaWith(simple, int64(0))
		table = Table(v)
		if table != "crud_simple" {
			t.Error("error")
			return
		}
		table, fields := QueryField(v, "tid#all")
		if table != "crud_simple" || len(fields) != 1 {
			t.Error("error")
			return
		}
	}
	{
		table, fields := QueryField(simple, "t.tid#all")
		if table != "crud_simple t" || len(fields) != 1 || fields[0] != "t.tid" {
			t.Error("error")
			return
		}
		table, fields = QueryField(MetaWith(simple, int64(0)), "t.tid#all")
		if table != "crud_simple t" || len(fields) != 1 || fields[0] != "t.tid" {
			t.Error("error")
			return
		}
		table, fields = QueryField(int64(0), "t.tid#all")
		if table != "" || len(fields) != 1 || fields[0] != "t.tid" {
			t.Error("error")
			return
		}
	}
	{
		table = FilterFieldCall("test", simple, "#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		})
		if table != "crud_simple" {
			t.Error("error")
			return
		}
	}
	{
		table = FilterFieldCall("test", int64(11), "tid#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); !ok || v != 11 {
				panic("error")
			}
		})
		if table != "" {
			t.Error("error")
			return
		}
		table = FilterFieldCall("test", int64(11), "count(tid)#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); !ok || v != 11 {
				panic("error")
			}
		})
		if table != "" {
			t.Error("error")
			return
		}
	}
	{
		table = FilterFieldCall("test", []interface{}{int64(11)}, "tid#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); !ok || v != 11 {
				panic("error")
			}
		})
		if table != "" {
			t.Error("error")
			return
		}
		table = FilterFieldCall("test", []interface{}{TableName("crud_simple"), int64(11), string("abc")}, "tid,title#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); ok && v != 11 {
				panic("error")
			}
			if v, ok := value.(string); ok && v != "abc" {
				panic("error")
			}
		})
		if table != "crud_simple" {
			t.Error("error")
			return
		}
		table = FilterFieldCall("test", []interface{}{TableName("crud_simple"), int64(11), string("abc")}, "count(tid),title#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); ok && v != 11 {
				panic("error")
			}
			if v, ok := value.(string); ok && v != "abc" {
				panic("error")
			}
		})
		if table != "crud_simple" {
			t.Error("error")
			return
		}
	}
	{ //error
		func() {
			defer func() {
				recover()
			}()
			FilterFieldCall("test", []interface{}{TableName("crud_simple"), int64(11), string("abc")}, "count(tid)", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			})
		}()
	}
}

type IsNilArray []string

type IsZeroArray []string

func (s IsNilArray) IsNil() bool {
	return s == nil
}

func (s IsZeroArray) IsZero() bool {
	return len(s) < 1
}

func TestFilterFormatCall(t *testing.T) {
	{
		var emptyStr string
		var nilStr *string
		var nilEmptyStr *string = &emptyStr
		FilterFormatCall("1,2,3", []interface{}{"1", "2", "3"}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", ""}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", nil}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", nilStr}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", nilEmptyStr}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3#all", []interface{}{"1", "2", ""}, func(format string, arg interface{}) {
			if format == "3" && arg.(string) != "" {
				t.Error("error")
			}
			if format != "3" && arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", IsNilArray(nil)}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
		FilterFormatCall("1,2,3", []interface{}{"1", "2", IsZeroArray(nil)}, func(format string, arg interface{}) {
			if arg.(string) != format {
				t.Error("error")
				return
			}
		})
	}
	{ //error
		func() {
			defer func() {
				recover()
			}()
			FilterFormatCall("x", []interface{}{1, 2}, func(format string, arg interface{}) {
			})
		}()
	}
}

func TestInsert(t *testing.T) {
	var err error
	title := "test"
	simple := &Simple{
		UserID: 100,
		Type:   "test",
		Title:  &title,
		Image:  &title,
		Status: SimpleStatusNormal,
	}
	simple.CreateTime = time.Now()
	simple.UpdateTime = time.Now()
	insertSQL, insertArg := InsertSQL(simple, "", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL)
	fmt.Printf("   --->%v\n", insertArg)
	insertSQL, insertArg = Default.InsertSQL(simple, "", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL)
	fmt.Printf("   --->%v\n", insertArg)
	insertSQL1, insertArg1 := InsertSQL(simple, "^update_time,create_time,status", "crud_simple", "returning tid")
	fmt.Printf("insert\n")
	fmt.Printf("   --->%v\n", insertSQL1)
	fmt.Printf("   --->%v\n", insertArg1)
	if strings.Contains(insertSQL1, "update_time") {
		err = fmt.Errorf("error")
		t.Errorf("%v,%v", err, insertSQL1)
		return
	}
	table, fileds, param, args := InsertArgs(simple, "")
	if table != "crud_simple" || len(fileds) < 1 || len(param) < 1 || len(args) < 1 {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	table, fileds, param, args = Default.InsertArgs(simple, "")
	if table != "crud_simple" || len(fileds) < 1 || len(param) < 1 || len(args) < 1 {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	fileds, param, args = AppendInsert(fileds, param, args, true, "xx=$%v", "1")
	if len(fileds) < 1 || len(param) < 1 || len(args) < 1 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}

	fileds, param, args = AppendInsertf(nil, nil, nil, "x1=$%v,x2=$%v,x3=$%v", "1", "2", "")
	if len(fileds) != 2 || len(param) != 2 || len(args) != 2 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}
	fmt.Println("AppendInsertf-->", fileds, param, args)

	fileds, param, args = AppendInsertf(nil, nil, nil, "x1=$%v,x2=$%v,x3=$%v#all", "1", "2", "")
	if len(fileds) != 3 || len(param) != 3 || len(args) != 3 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}
	fmt.Println("AppendInsertf-->", fileds, param, args)

	_, err = InsertFilter(NewPoolQueryer(), simple, "^tid#all", "returning", "tid#all")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = InsertFilter(NewPoolQueryer(), simple, "^tid#all", "", "")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.InsertFilter(NewPoolQueryer(), simple, "^tid#all", "", "")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdate(t *testing.T) {
	var err error
	title := "test"
	simple := &Simple{
		UserID: 100,
		Type:   "test",
		Title:  &title,
		Image:  &title,
		Status: SimpleStatusNormal,
	}
	simple.UpdateTime = time.Now()
	var updateSQL string
	var where []string
	var args []interface{}

	updateSQL, args = UpdateSQL(simple, "title,image,update_time,status#all", nil)

	if strings.Contains(updateSQL, "tid") {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}

	where, args = AppendWhere(where, args, simple.TID > 0, "tid=$%v", simple.TID)
	where, args = AppendWhere(where, args, simple.UserID > 0, "user_id=$%v", simple.UserID)

	updateSQL = JoinWhere(updateSQL, where, "and")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	updateSQL, args = UpdateSQL(simple, "", nil)
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)
	updateSQL, args = Default.UpdateSQL(simple, "", nil)
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", updateSQL)
	fmt.Printf("   --->%v\n", args)

	table, sets, args := UpdateArgs(simple, "", nil)
	if table != "crud_simple" || len(sets) < 1 || len(args) < 1 {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	table, sets, args = Default.UpdateArgs(simple, "", nil)
	if table != "crud_simple" || len(sets) < 1 || len(args) < 1 {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", table)
	fmt.Printf("   --->%v\n", sets)
	fmt.Printf("   --->%v\n", args)

	sets, args = AppendSet(sets, args, true, "xx=$%v", "1")
	fmt.Printf("update\n")
	fmt.Printf("   --->%v\n", sets)
	fmt.Printf("   --->%v\n", args)

	sets, args = AppendSetf(nil, nil, "x1=$%v,x2=$%v,x3=$%v", "1", "2", "")
	if len(sets) != 2 || len(args) != 2 {
		t.Error("error")
		return
	}
	fmt.Println("AppendSetf-->", sets, args)
	sets, args = AppendSetf(nil, nil, "x1=$%v,x2=$%v,x3=$%v#all", "1", "2", "")
	if len(sets) != 3 || len(args) != 3 {
		t.Error("error")
		return
	}
	fmt.Println("AppendSetf-->", sets, args)

	_, err = Update(NewPoolQueryer(), simple, sets, where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.Update(NewPoolQueryer(), simple, sets, where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateRow(NewPoolQueryer(), simple, sets, where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.UpdateRow(NewPoolQueryer(), simple, sets, where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = UpdateFilter(NewPoolQueryer(), simple, "title,image,update_time,status", where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.UpdateFilter(NewPoolQueryer(), simple, "title,image,update_time,status", where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateFilterRow(NewPoolQueryer(), simple, "title,image,update_time,status", where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.UpdateFilterRow(NewPoolQueryer(), simple, "title,image,update_time,status", where, "and", args)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = UpdateSimple(NewPoolQueryer(), simple, "title,image,update_time,status", "", nil)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = Default.UpdateSimple(NewPoolQueryer(), simple, "title,image,update_time,status", "", nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateSimpleRow(NewPoolQueryer(), simple, "title,image,update_time,status", "where tid=$1", []interface{}{1})
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.UpdateSimpleRow(NewPoolQueryer(), simple, "title,image,update_time,status", "where tid=$1", []interface{}{1})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestQueryField(t *testing.T) {
	simple := &Simple{}
	fmt.Println(QueryField(simple, "count(tid)#all"))
	fmt.Println(QueryField(int64(0), "count(tid)#all"))
}

func TestSearch(t *testing.T) {
	var err error
	var userID int64
	var key string
	simple := &Simple{}
	var where []string
	var args []interface{}
	querySQL := QuerySQL(simple, "#nil,zero")
	where, args = AppendWhere(where, args, true, "user_id=$%v", userID)
	where, args = AppendWhere(where, args, true, "(title like $%v or data like $%v)", "%"+key+"%")
	querySQL = JoinWhere(querySQL, where, "and")
	querySQL = JoinPage(querySQL, "order by tid", 0, 10)
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", querySQL)
	fmt.Printf("   --->%v\n", args)

	querySQL = Default.QuerySQL(simple, "#nil,zero")
	where, args = Default.AppendWhere(where, args, true, "user_id=$%v", userID)
	where, args = Default.AppendWhere(where, args, true, "(title like $%v or data like $%v)", "%"+key+"%")
	querySQL = Default.JoinWhere(querySQL, where, "and")
	querySQL = Default.JoinPage(querySQL, "order by tid", 0, 10)
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", querySQL)
	fmt.Printf("   --->%v\n", args)

	where, args = AppendWheref(nil, nil, "x1=$%v,x2=$%v,x3=$%v", "1", "2", "")
	if len(where) != 2 || len(args) != 2 {
		t.Error("error")
		return
	}
	fmt.Println("AppendWheref-->", where, args)
	where, args = AppendWheref(nil, nil, "x1=$%v,x2=$%v,x3=$%v#all", "1", "2", "")
	if len(where) != 3 || len(args) != 3 {
		t.Error("error")
		return
	}
	fmt.Println("AppendWheref-->", where, args)

	table, fileds := QueryField(simple, "#all")
	if table != "crud_simple" {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", fileds)
	if len(fileds) < 1 {
		err = fmt.Errorf("fail")
		t.Error(err)
		return
	}
	table, fileds = Default.QueryField(simple, "#all")
	if table != "crud_simple" {
		err = fmt.Errorf("table error")
		t.Error(err)
		return
	}
	fmt.Printf("query\n")
	fmt.Printf("   --->%v\n", fileds)
}

type UserIDx map[int64]string

func (u UserIDx) Scan(v interface{}) {
	s := v.(*Simple)
	u[s.TID] = fmt.Sprintf("%v-%v", s.Title, s.UserID)
}

func TestQuery(t *testing.T) {
	var err error
	var simpleList []*Simple
	var simple *Simple
	var userIDs0, userIDs1 []int64
	var images0, images1 []*string
	var simples0 map[int64]*Simple
	var simples1 = map[int64][]*Simple{}
	var userIDm0 map[int64]int64
	var userIDm1 map[int64][]int64
	var userIDx = UserIDx{}
	err = Query(
		NewPoolQueryer(), &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simpleList, &simple,
		&userIDs0, "user_id",
		&userIDs1, "user_id#all",
		&images0, "image",
		&images1, "image#all",
		&simples0, "tid",
		&simples1, "user_id",
		&userIDm0, "tid:user_id",
		&userIDm1, "user_id:tid",
		&userIDx,
		func(v *Simple) {
			// simples1[v.TID] = v
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
	if len(simpleList) < 1 || simple == nil || len(userIDs0) < 1 || len(userIDs1) < 1 ||
		len(images0) < 1 || len(images1) < 1 || len(simples0) < 1 || len(simples1) < 1 ||
		len(userIDm0) < 1 || len(userIDm1) < 1 || len(userIDx) < 1 {
		fmt.Printf(`
			len(simpleList)=%v, simple=%v, len(userIDs0)=%v, len(userIDs1)=%v,
			len(images0)=%v, len(images1)=%v, len(simples0)=%v, len(simples1)=%v,
			len(userIDm0)=%v, len(userIDm1)=%v, len(userIDx)=%v
			`,
			len(simpleList), simple == nil, len(userIDs0), len(userIDs1),
			len(images0), len(images1), len(simples0), len(simples1),
			len(userIDm0), len(userIDm1), len(userIDx),
		)
		err = fmt.Errorf("data error")
		t.Error(err)
		return
	}
	data, _ := json.MarshalIndent(map[string]interface{}{
		"simple":      simple,
		"list":        simpleList,
		"user_ids_0":  userIDs0,
		"user_ids_1":  userIDs1,
		"images_0":    images0,
		"images_1":    images1,
		"simples_0":   simples0,
		"simples_1":   simples1,
		"user_id_m_0": userIDm0,
		"user_id_m_1": userIDm1,
		"user_id_x":   userIDx,
	}, "", "\t")
	fmt.Println("-->\n", string(data))

	simple = nil
	err = QueryRow(
		NewPoolQueryer(), &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		&simple,
	)
	if err != nil || simple == nil {
		t.Error(err)
		return
	}

	simple = &Simple{}
	err = QueryRow(
		NewPoolQueryer(), &Simple{}, "#all",
		"sql", []interface{}{"arg"},
		simple,
	)
	if err != nil || simple == nil || simple.TID < 1 {
		t.Error(err)
		return
	}

	//
	//test int64
	var testIDs0 []int64
	var testIDs1 int64
	testIDs0 = nil
	err = Query(
		&PoolQueryer{}, int64(0), "",
		"int64", []interface{}{"arg"},
		&testIDs0,
	)
	if err != nil || len(testIDs0) != 3 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}
	testIDs0 = nil
	err = Default.Query(
		&PoolQueryer{}, int64(0), "",
		"int64", []interface{}{"arg"},
		&testIDs0,
	)
	if err != nil || len(testIDs0) != 3 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}
	err = QueryRow(
		&PoolQueryer{}, int64(0), "",
		"int64", []interface{}{"arg"},
		&testIDs1,
	)
	if err != nil || testIDs1 < 1 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}
	err = Default.QueryRow(
		&PoolQueryer{}, int64(0), "",
		"int64", []interface{}{"arg"},
		&testIDs1,
	)
	if err != nil || testIDs1 < 1 {
		err = fmt.Errorf("error")
		t.Error(err)
		return
	}

	//
	//
	simpleList = []*Simple{}
	err = QueryFilter(NewPoolQueryer(), &Simple{}, "#all", nil, "", nil, "", 0, 10, &simpleList)
	if err != nil {
		return
	}
	if len(simpleList) < 1 {
		err = fmt.Errorf("data error")
		t.Error(err)
		return
	}
	simpleList = []*Simple{}
	err = Default.QueryFilter(NewPoolQueryer(), &Simple{}, "#all", nil, "", nil, "", 0, 10, &simpleList)
	if err != nil {
		return
	}
	if len(simpleList) < 1 {
		err = fmt.Errorf("data error")
		t.Error(err)
		return
	}
	err = QueryFilterRow(NewPoolQueryer(), &Simple{}, "#all", nil, "", []interface{}{"arg"}, &simple)
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.QueryFilterRow(NewPoolQueryer(), &Simple{}, "#all", nil, "", []interface{}{"arg"}, &simple)
	if err != nil {
		t.Error(err)
		return
	}
	//
	simpleList = []*Simple{}
	err = QuerySimple(NewPoolQueryer(), &Simple{}, "#all", "", []interface{}{"arg"}, 0, 10, &simpleList)
	if err != nil {
		t.Error(err)
		return
	}
	if len(simpleList) < 1 {
		err = fmt.Errorf("data error")
		t.Error(err)
		return
	}
	simpleList = []*Simple{}
	err = Default.QuerySimple(NewPoolQueryer(), &Simple{}, "#all", "", []interface{}{"arg"}, 0, 10, &simpleList)
	if err != nil {
		t.Error(err)
		return
	}
	if len(simpleList) < 1 {
		err = fmt.Errorf("data error")
		t.Error(err)
		return
	}
	err = QuerySimpleRow(NewPoolQueryer(), &Simple{}, "#all", "", []interface{}{"arg"}, &simple)
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.QuerySimpleRow(NewPoolQueryer(), &Simple{}, "#all", "", []interface{}{"arg"}, &simple)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCount(t *testing.T) {
	var err error
	simple := &Simple{}
	{
		if countSQL := CountSQL(simple, ""); !strings.Contains(countSQL, "count(*)") {
			t.Errorf("err:%v", countSQL)
			return
		}
		if countSQL := CountSQL(simple, "*"); !strings.Contains(countSQL, "count(*)") {
			t.Errorf("err:%v", countSQL)
			return
		}
		if countSQL := CountSQL(simple, "count(*)"); !strings.Contains(countSQL, "count(*)") {
			t.Errorf("err:%v", countSQL)
			return
		}
		if countSQL := CountSQL(simple, "count(*)#all"); !strings.Contains(countSQL, "count(*)") {
			t.Errorf("err:%v", countSQL)
			return
		}
		if countSQL := CountSQL(simple, "count(tid)#all"); !strings.Contains(countSQL, "count(tid)") {
			t.Errorf("err:%v", countSQL)
			return
		}
		if countSQL := Default.CountSQL(simple, "count(tid)#all"); !strings.Contains(countSQL, "count(tid)") {
			t.Errorf("err:%v", countSQL)
			return
		}
	}
	{
		var countVal int64
		err = Count(NewPoolQueryer(), simple, "tid#all", "select count(*) from crud_simple", nil, &countVal, "tid")
		if err != nil || countVal < 1 {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
		err = Default.Count(NewPoolQueryer(), simple, "tid#all", "select count(*) from crud_simple", nil, &countVal, "tid")
		if err != nil || countVal < 1 {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
	{
		var countVal int64
		err = CountFilter(NewPoolQueryer(), simple, "count(tid)#all", nil, "", nil, &countVal, "tid")
		if err != nil || countVal < 1 {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
		err = Default.CountFilter(NewPoolQueryer(), simple, "count(tid)#all", nil, "", nil, &countVal, "tid")
		if err != nil || countVal < 1 {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
	{
		var countVal int64
		err = CountSimple(NewPoolQueryer(), simple, "count(tid)#all", "where tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
		err = Default.CountSimple(NewPoolQueryer(), simple, "count(tid)#all", "where tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
	{
		var countVal *int64
		err = CountSimple(NewPoolQueryer(), MetaWith(simple, countVal), "count(tid)#all", "where tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(*countVal)
	}
	{
		var typeVal *string
		err = QuerySimpleRow(NewPoolQueryer(), MetaWith(simple, int64(0), int64(0), typeVal), "tid,user_id,type#all", "where tid>$1", []interface{}{1}, &typeVal, "type")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(*typeVal)
	}
	{
		var countVal int64
		err = QuerySimpleRow(NewPoolQueryer(), MetaWith(simple, countVal), "tid#all", "where tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
	{
		var typeVal string
		err = QuerySimpleRow(NewPoolQueryer(), MetaWith(simple, int64(0), int64(0), typeVal), "tid,user_id,type#all", "where tid>$1", []interface{}{1}, &typeVal, "type")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(typeVal)
	}
	{
		var countVal int64
		err = CountSimple(NewPoolQueryer(), MetaWith(simple, countVal), "count(tid)#all", "where tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
	{
		var countVal int64
		err = CountSimple(NewPoolQueryer(), MetaWith(simple, countVal), "s.count(tid)#all", "where s.tid>$1", []interface{}{1}, &countVal, "tid")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(countVal)
	}
}

func TestScan(t *testing.T) {
	var err error
	{
		rows, _ := (NewPoolQueryer()).Query("")
		err = Scan(
			rows, &Simple{}, "#all",
			func(v *Simple) {
				// simples1[v.TID] = v
			},
		)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		row := (NewPoolQueryer()).QueryRow("")
		err = ScanRow(
			row, &Simple{}, "#all",
			func(v *Simple) {
				// simples1[v.TID] = v
			},
		)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

type ListSimpleUnify struct {
	Model Simple `json:"model"`
	Where struct {
		UserID int64 `json:"user_id"`
		Type   int   `json:"type" filter:"#all"`
		Key    struct {
			Title string `json:"title" cmp:"title like $%v"`
			Data  string `json:"data" cmp:"data like $%v"`
		} `json:"key" join:"or"`
		Status []int `cmp:"status=any($%v)"`
		Empty  int
	} `json:"where" join:"and"`
	Page struct {
		Order  string `json:"order"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	} `json:"page"`
	Query struct {
		Enabled bool      `json:"enabled" scan:"-"`
		Simples []*Simple `json:"simples"`
		UserIDs []int64   `json:"user_ids" scan:"user_id#all"`
	} `json:"query" filter:"#all"`
	Count struct {
		Enabled bool  `json:"enabled" scan:"-"`
		All     int64 `json:"all" scan:"tid"`
		UserID  int64 `json:"user_id" scan:"user_id"`
	} `json:"count" filter:"count(tid),max(user_id)#all"`
}

type FindSimpleUnify struct {
	Model Simple `json:"model"`
	Where struct {
		UserID int64  `json:"user_id" cmp:"user_id >"`
		Type   int    `json:"type" filter:"#all"`
		Key    string `json:"key" cmp:"title like $%v or data like $%v"`
		Status []int  `json:"status" cmp:"any($%v)"`
		Empty  int    `json:"empty"`
	} `json:"where" join:"and"`
	QueryRow struct {
		Enabled bool    `json:"enabled" scan:"-"`
		Simple  *Simple `json:"simple"`
		UserID  int64   `json:"user_id" scan:"user_id#all"`
	} `json:"query" filter:"#all"`
}

func TestUnify(t *testing.T) {
	var err error
	list := &ListSimpleUnify{}
	list.Where.UserID = 100
	list.Where.Type = 10
	list.Where.Key.Title = "%a%"
	list.Where.Key.Data = "%a%"
	list.Where.Status = []int{10, 100}
	list.Query.Enabled = true
	list.Count.Enabled = true
	find := &FindSimpleUnify{}
	find.Where.UserID = 100
	find.Where.Type = 10
	find.Where.Key = "%a%"
	find.Where.Status = []int{10, 100}
	find.QueryRow.Enabled = true

	//
	where, args := AppendWhereUnify(nil, nil, list)
	fmt.Println("AppendWhereUnify-->", strings.Join(where, " and "), args)
	sql, args := JoinWhereUnify("select tid from crud_simple", nil, list)
	fmt.Println("JoinWhereUnify-->", sql, args)
	sql, args = Default.JoinWhereUnify("select tid from crud_simple", nil, list)
	fmt.Println("JoinWhereUnify-->", sql, args)
	sql = JoinPageUnify(sql, list)
	fmt.Println("JoinPageUnify-->", sql)
	sql = Default.JoinPageUnify(sql, list)
	fmt.Println("JoinPageUnify-->", sql)
	sql, args = QueryUnifySQL(list)
	fmt.Println("QueryUnifySQL-->", sql, args)
	sql, args = Default.QueryUnifySQL(list)
	fmt.Println("QueryUnifySQL-->", sql, args)
	modelValue, queryFilter, dests := ScanUnifyDest(list, "Query")
	if modelValue != &list.Model || queryFilter != "#all" || len(dests) != 3 {
		t.Errorf("%v,%v,%v", reflect.TypeOf(modelValue), queryFilter, len(dests))
		return
	}
	sql, args = CountUnifySQL(list)
	fmt.Println("CountUnifySQL-->", sql, args)
	sql, args = Default.CountUnifySQL(list)
	fmt.Println("CountUnifySQL-->", sql, args)
	modelValue, queryFilter, dests = CountUnifyDest(list)
	if len(modelValue.([]interface{})) != 3 || queryFilter != "count(tid),max(user_id)#all" || len(dests) != 4 {
		t.Errorf("%v,%v,%v", len(modelValue.([]interface{})), queryFilter, len(dests))
		return
	}

	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = QueryUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("QueryUnify-->", jsonString(list))
	if len(list.Query.UserIDs) != 3 || len(list.Query.Simples) != 3 {
		t.Error("error")
		return
	}
	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = Default.QueryUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("QueryUnify-->", jsonString(list))
	if len(list.Query.UserIDs) != 3 || len(list.Query.Simples) != 3 {
		t.Error("error")
		return
	}

	list.Query.Simples = nil
	list.Query.UserIDs = nil
	rows, _ := NewPoolQueryer().Query("select")
	err = ScanUnify(rows, list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ScanUnify-->", jsonString(list))
	if len(list.Query.UserIDs) != 3 || len(list.Query.Simples) != 3 {
		t.Error("error")
		return
	}

	find.QueryRow.Simple = nil
	find.QueryRow.UserID = 0
	err = QueryUnifyRow(NewPoolQueryer(), find)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("QueryUnifyRow-->", jsonString(find))
	if find.QueryRow.UserID < 1 || find.QueryRow.Simple == nil {
		t.Error("error")
		return
	}
	find.QueryRow.Simple = nil
	find.QueryRow.UserID = 0
	err = Default.QueryUnifyRow(NewPoolQueryer(), find)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("QueryUnifyRow-->", jsonString(list))
	if find.QueryRow.UserID < 1 || find.QueryRow.Simple == nil {
		t.Error("error")
		return
	}

	find.QueryRow.Simple = nil
	find.QueryRow.UserID = 0
	row := NewPoolQueryer().QueryRow("select")
	err = ScanUnifyRow(row, find)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ScanUnifyRow-->", jsonString(find))
	if find.QueryRow.UserID < 1 || find.QueryRow.Simple == nil {
		t.Error("error")
		return
	}

	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = CountUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("CountUnify-->", jsonString(list))
	if list.Count.All < 1 || list.Count.UserID < 1 {
		t.Error("error")
		return
	}
	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = Default.CountUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("CountUnify-->", jsonString(list))
	if list.Count.All < 1 || list.Count.UserID < 1 {
		t.Error("error")
		return
	}

	find.QueryRow.Simple = nil
	find.QueryRow.UserID = 0
	err = QueryUnifyRow(NewPoolQueryer(), find)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("QueryUnifyRow-->", jsonString(find))
	if find.QueryRow.UserID < 1 || find.QueryRow.Simple == nil {
		t.Error("error")
		return
	}

	find.QueryRow.Simple = nil
	find.QueryRow.UserID = 0
	err = ApplyUnify(NewPoolQueryer(), find)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ApplyUnify-->", jsonString(find))
	if find.QueryRow.UserID < 1 || find.QueryRow.Simple == nil {
		t.Error("error")
		return
	}
	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = ApplyUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ApplyUnify-->", jsonString(list))
	if len(list.Query.UserIDs) != 3 || len(list.Query.Simples) != 3 || list.Count.All < 1 || list.Count.UserID < 1 {
		t.Error("error")
		return
	}
	list.Query.Simples = nil
	list.Query.UserIDs = nil
	err = Default.ApplyUnify(NewPoolQueryer(), list)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ApplyUnify-->", jsonString(list))
	if len(list.Query.UserIDs) != 3 || len(list.Query.Simples) != 3 || list.Count.All < 1 || list.Count.UserID < 1 {
		t.Error("error")
		return
	}
}

func TestError(t *testing.T) {
	var err error
	list := &ListSimpleUnify{}
	list.Where.UserID = 100
	list.Where.Type = 10
	list.Where.Key.Title = "%a%"
	list.Where.Key.Data = "%a%"
	list.Where.Status = []int{10, 100}
	find := &FindSimpleUnify{}
	find.Where.UserID = 100
	find.Where.Type = 10
	find.Where.Key = "%a%"
	find.Where.Status = []int{10, 100}
	{ //insert error
		simple := &Simple{}
		pool := NewPoolQueryer()

		//
		pool.QueryMode = "no"
		_, err = InsertFilter(pool, simple, "^tid#all", "returning", "tid#all")
		if err != Default.ErrNoRows {
			t.Error(err)
			return
		}

		//
		pool.QueryErr = fmt.Errorf("error")
		_, err = InsertFilter(pool, simple, "^tid#all", "returning", "tid#all")
		if err == nil {
			t.Error(err)
			return
		}
	}
	{ //update error
		simple := &Simple{}
		pool := NewPoolQueryer()

		//
		pool.ExecAffected = 1
		pool.ExecErr = fmt.Errorf("error")
		_, sets, args := UpdateArgs(simple, "", nil)
		_, err = Update(pool, simple, sets, nil, "", args)
		if err == nil {
			t.Error("not error")
			return
		}
		_, err = UpdateFilter(pool, simple, "title,image,update_time,status", nil, "", nil)
		if err == nil {
			t.Error("not error")
			return
		}
		_, err = UpdateSimple(pool, simple, "title,image,update_time,status", "", nil)
		if err == nil {
			t.Error("not error")
			return
		}

		//
		pool.ExecAffected = 0
		pool.ExecErr = nil
		err = UpdateRow(pool, simple, sets, nil, "", args)
		if err != Default.ErrNoRows {
			t.Error("not error")
			return
		}
		err = UpdateFilterRow(pool, simple, "title,image,update_time,status", nil, "", nil)
		if err != Default.ErrNoRows {
			t.Error("not error")
			return
		}
		err = UpdateSimpleRow(pool, simple, "title,image,update_time,status", "", nil)
		if err == nil {
			t.Error("not error")
			return
		}
	}
	{ //query error
		var images []*string
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&images,
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&images, 1,
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&images, "",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&images, "not",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = QueryRow(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&images, "not",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = QueryRow(
			NewPoolQueryer(), &Simple{}, "#all",
			"no", []interface{}{"arg"},
			&images, "image",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		//
		var simples map[int64]*Simple
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples,
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, 1,
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, "",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, "not",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, "tid:not",
		)
		if err == nil {
			t.Error("not error")
			return
		}

		//
		err = Query(
			NewPoolQueryer(), &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			1,
		)
		if err == nil {
			t.Error("not error")
			return
		}

		//
		err = Query(
			&PoolQueryer{QueryErr: fmt.Errorf("xx")}, &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, "tid",
		)
		if err == nil {
			t.Error("not error")
			return
		}
		err = Query(
			&PoolQueryer{ScanErr: fmt.Errorf("xx")}, &Simple{}, "#all",
			"sql", []interface{}{"arg"},
			&simples, "tid",
		)
		if err == nil {
			t.Error("not error")
			return
		}

		//
		find.QueryRow.Simple = nil
		find.QueryRow.UserID = 0
		err = QueryUnify(&PoolQueryer{QueryErr: fmt.Errorf("xx")}, find)
		if err == nil {
			t.Error(err)
			return
		}
		find.QueryRow.Simple = nil
		find.QueryRow.UserID = 0
		err = QueryUnifyRow(&PoolQueryer{ScanErr: fmt.Errorf("xx")}, find)
		if err == nil {
			t.Error(err)
			return
		}
		func() {
			defer func() {
				recover()
			}()
			QueryUnifyRow(&PoolQueryer{ScanErr: fmt.Errorf("xx")}, list)
		}()
	}
	{ //count error
		var countValue int64
		simple := &Simple{}
		pool := NewPoolQueryer()

		//
		pool.QueryMode = "no"
		err = CountSimple(pool, simple, "count(tid)#all", "", nil, &countValue, "tid")
		if err != Default.ErrNoRows {
			t.Error(err)
			return
		}

		//
		pool.QueryErr = fmt.Errorf("error")
		err = CountSimple(pool, simple, "count(tid)#all", "", nil, &countValue, "tid")
		if err == nil {
			t.Error(err)
			return
		}
		list.Query.Simples = nil
		list.Query.UserIDs = nil
		err = CountUnify(pool, list)
		if err == nil {
			t.Error(err)
			return
		}
	}
	{ //dest error
		var countValue int64
		simple := &Simple{}
		pool := NewPoolQueryer()

		//
		err = CountSimple(pool, simple, "count(tid)#all", "", nil, &countValue, "title#all")
		if err == nil || !strings.Contains(err.Error(), "not supported on dests[0] to set") {
			t.Error(err)
			return
		}
		//
		err = CountSimple(pool, []interface{}{int64(0)}, "count(tid)#all", "", nil, &countValue, "none#all")
		if err == nil {
			t.Error(err)
			return
		}
	}
}
