package crud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"strings"
	"testing"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xsql"
	"github.com/shopspring/decimal"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	Default.NameConv("", "", reflect.StructField{})
	Default.ParmConv("", "", "", reflect.StructField{}, nil)
	Default.getErrNoRows()
	Default.ErrNoRows = ErrNoRows
	Default.Verbose = true
	Default.NameConv = func(on, name string, field reflect.StructField) string {
		if on == "query" && strings.HasPrefix(field.Type.String(), "xsql.") && field.Type.String() != "xsql.Time" {
			return name + "::text"
		} else {
			return name
		}
	}
	Default.ParmConv = func(on, fieldName, fieldFunc string, field reflect.StructField, value interface{}) interface{} {
		if c, ok := value.(xsql.ArrayConverter); on == "where" && ok {
			return c.DbArray()
		}
		return value
	}
	go http.ListenAndServe(":6063", nil)
}

func TestJsonString(t *testing.T) {
	jsonString(t)
	jsonString(t.Error)
}

func TestBuildOrderby(t *testing.T) {
	if v := BuildOrderby("x", "+y"); len(v) > 0 {
		t.Error("error")
		return
	}
	if v := BuildOrderby("x", "+x"); v != "order by x asc" {
		t.Error(v)
		return
	}
	if v := BuildOrderby("x", "-x"); v != "order by x desc" {
		t.Error(v)
		return
	}
}

func TestQueryCall(t *testing.T) {
	clearPG()
	testQueryCall(t, getPG())
}

func testQueryCall(t *testing.T, queryer Queryer) {
	var err error
	var rows Rows

	rows, err = Default.queryerQuery(queryer, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	rows.Scan(converter.IntPtr(0))
	rows.Close()
	rows, err = Default.queryerQuery(&TestCrudQueryer{Queryer: queryer}, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	rows.Scan(converter.IntPtr(0))
	rows.Close()
	rows, err = Default.queryerQuery(func() interface{} { return queryer }, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	rows.Scan(converter.IntPtr(0))
	rows.Close()
	func() {
		defer func() {
			recover()
		}()
		Default.queryerQuery("xxx", context.Background(), "select 1", []interface{}{})
	}()

	err = Default.queryerQueryRow(queryer, context.Background(), "select 1", []interface{}{}).Scan(converter.IntPtr(0))
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.queryerQueryRow(&TestCrudQueryer{Queryer: queryer}, context.Background(), "select 1", []interface{}{}).Scan(converter.IntPtr(0))
	if err != nil {
		t.Error(err)
		return
	}
	err = Default.queryerQueryRow(func() interface{} { return queryer }, context.Background(), "select 1", []interface{}{}).Scan(converter.IntPtr(0))
	if err != nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Default.queryerQueryRow("xxx", context.Background(), "select 1", []interface{}{})
	}()

	_, _, err = Default.queryerExec(queryer, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	_, _, err = Default.queryerExec(&TestCrudQueryer{Queryer: queryer}, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	_, _, err = Default.queryerExec(func() interface{} { return queryer }, context.Background(), "select 1", []interface{}{})
	if err != nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		Default.queryerExec("xxx", context.Background(), "select 1", []interface{}{})
	}()
}

func TestNewValue(t *testing.T) {
	{
		value := NewValue(CrudObject{})
		if _, ok := value.Interface().(*CrudObject); !ok {
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
		var intVal int
		if v := NewValue(intVal); v.Type() != reflect.TypeOf(&intVal) {
			t.Error("error")
			return
		}
		var intPtr *int
		if v := NewValue(intPtr); v.Type() != reflect.TypeOf(&intPtr) {
			t.Error("error")
			return
		}
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
	object := &CrudObject{
		Type:   "test",
		Title:  "title",
		Image:  converter.StringPtr("image"),
		Data:   xsql.M{"abc": 1},
		Status: CrudObjectStatusNormal,
	}
	{
		table := Table(object)
		if table != "crud_object" {
			t.Error("error")
			return
		}
		table, fields := QueryField(object, "")
		if table != "crud_object" || len(fields) != 5 {
			t.Errorf("%v,%v", table, fields)
			return
		}
		table, fields = QueryField(object, "tid#all|data#all")
		if table != "crud_object" || len(fields) != 2 || fields[0] != "tid" || fields[1] != "data::text" {
			t.Errorf("%v", fields)
			return
		}
		table, fields = QueryField(object, "tid#all|")
		if table != "crud_object" || len(fields) != 6 {
			t.Errorf("%v", fields)
			return
		}
		table, fields = QueryField(object, "tid#all|*")
		if table != "crud_object" || len(fields) != 6 {
			t.Errorf("%v", fields)
			return
		}
		table, fields = QueryField(object, "*|tid#all")
		if table != "crud_object" || len(fields) != 6 {
			t.Errorf("%v", fields)
			return
		}
		table, fields = Default.QueryField(object, "*|tid#all")
		if table != "crud_object" || len(fields) != 6 {
			t.Errorf("%v", fields)
			return
		}
	}
	{
		v := MetaWith("crud_object", int64(0))
		table := Table(v)
		if table != "crud_object" {
			t.Error("error")
			return
		}
		table, fields := QueryField(v, "tid#all")
		if table != "crud_object" || len(fields) != 1 {
			t.Error("error")
			return
		}
	}
	{
		v := MetaWith(object, int64(0))
		table := Table(v)
		if table != "crud_object" {
			t.Error("error")
			return
		}
		table, fields := QueryField(v, "tid#all")
		if table != "crud_object" || len(fields) != 1 {
			t.Error("error")
			return
		}
	}
	{
		table, fields := QueryField(object, "t.tid#all")
		if table != "crud_object t" || len(fields) != 1 || fields[0] != "t.tid" {
			t.Error("error")
			return
		}
		table, fields = QueryField(MetaWith(object, int64(0)), "t.tid#all")
		if table != "crud_object t" || len(fields) != 1 || fields[0] != "t.tid" {
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
		table := FilterFieldCall("test", object, "#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		})
		if table != "crud_object" {
			t.Error("error")
			return
		}
	}
	{
		table := FilterFieldCall("test", int64(11), "tid#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
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
		table := FilterFieldCall("test", []interface{}{int64(11)}, "tid#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); !ok || v != 11 {
				panic("error")
			}
		})
		if table != "" {
			t.Error("error")
			return
		}
		table = FilterFieldCall("test", []interface{}{TableName("crud_object"), int64(11), string("abc")}, "tid,title#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); ok && v != 11 {
				panic("error")
			}
			if v, ok := value.(string); ok && v != "abc" {
				panic("error")
			}
		})
		if table != "crud_object" {
			t.Error("error")
			return
		}
		table = FilterFieldCall("test", []interface{}{TableName("crud_object"), int64(11), string("abc")}, "count(tid),title#all", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
			if v, ok := value.(int64); ok && v != 11 {
				panic("error")
			}
			if v, ok := value.(string); ok && v != "abc" {
				panic("error")
			}
		})
		if table != "crud_object" {
			t.Error("error")
			return
		}
	}
	{ //error
		func() {
			defer func() {
				recover()
			}()
			FilterFieldCall("test", []interface{}{TableName("crud_object"), int64(11), string("abc")}, "count(tid)", func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
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

func newTestObject() (object *CrudObject) {
	object = &CrudObject{
		Type:   "test",
		Title:  "title",
		Image:  converter.StringPtr("image"),
		Status: CrudObjectStatusNormal,
	}
	object.TimeValue = xsql.TimeNow()
	object.CreateTime = xsql.TimeNow()
	object.UpdateTime = xsql.TimeNow()
	return
}

func addTestMultiObject(queryer Queryer) (object *CrudObject, objects []*CrudObject, err error) {
	object = newTestObject()
	object.UserID = 100
	object.Type = CrudObjectTypeA
	object.Level = 1
	object.Title = "abc"
	object.IntValue = 1
	object.Int64Value = 1
	object.Float64Value = decimal.NewFromFloat(1)
	object.StringValue = fmt.Sprintf("%v", 1)
	object.IntPtr = &object.IntValue
	object.Int64Ptr = &object.Int64Value
	object.Float64Ptr = object.Float64Value
	object.StringPtr = &object.StringValue
	object.Status = CrudObjectStatusNormal
	_, err = InsertFilter(queryer, context.Background(), object, "^tid#all", "returning", "tid#all")
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		obj := newTestObject()
		object.UserID = int64(100 + i%3)
		obj.Type = CrudObjectTypeB
		obj.Level = 1000 + i
		obj.Title = fmt.Sprintf("abc-%v", i)
		obj.IntValue = i % 3
		obj.Int64Value = int64(i % 3)
		object.Float64Value = decimal.NewFromInt(int64(i % 3))
		obj.StringValue = fmt.Sprintf("%v", i%3)
		obj.Status = CrudObjectStatusNormal
		_, err = InsertFilter(queryer, context.Background(), obj, "^tid#all", "returning", "tid#all")
		if err != nil {
			return
		}
		// objects = append(objects, obj)
	}
	return
}

func TestInsert(t *testing.T) {
	clearPG()
	testInsert(t, getPG())
}

func testInsert(t *testing.T, queryer Queryer) {
	var err error
	{
		object := newTestObject()
		insertSQL, insertArg := InsertSQL(object, "", "returning tid")
		fmt.Printf("insert\n")
		fmt.Printf("   --->%v\n", insertSQL)
		fmt.Printf("   --->%v\n", insertArg)
		err = queryer.QueryRow(context.Background(), insertSQL, insertArg...).Scan(&object.TID)
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		object := newTestObject()
		insertSQL, insertArg := Default.InsertSQL(object, "", "returning tid")
		fmt.Printf("insert\n")
		fmt.Printf("   --->%v\n", insertSQL)
		fmt.Printf("   --->%v\n", insertArg)
		err = queryer.QueryRow(context.Background(), insertSQL, insertArg...).Scan(&object.TID)
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		object := newTestObject()
		table, fields, param, args := InsertArgs(object, "")
		if table != "crud_object" || len(fields) < 1 || len(param) < 1 || len(args) < 1 {
			t.Error("error")
			return
		}
		fmt.Println("InsertArgs-->", fields, param, args)
		fields, param, args = AppendInsert(fields, param, args, true, "data=$%v", xsql.M{"abc": 1})
		if len(fields) < 1 || len(param) < 1 || len(args) < 1 {
			t.Error("error")
			return
		}
		fmt.Println("AppendInsert-->", fields, param, args)
		insertSQL := fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, strings.Join(fields, ","), strings.Join(param, ","), "returning tid")
		fmt.Println("insertSQL-->", insertSQL)
		err = queryer.QueryRow(context.Background(), insertSQL, args...).Scan(&object.TID)
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		object := newTestObject()
		table, fields, param, args := Default.InsertArgs(object, "")
		if table != "crud_object" || len(fields) < 1 || len(param) < 1 || len(args) < 1 {
			t.Error("err")
			return
		}
		fmt.Println("InsertArgs-->", fields, param, args)
		fields, param, args = AppendInsertf(fields, param, args, "int_value=$%v,int64_value=$%v,string_value=$%v", 1, 2, "")
		if len(fields) < 1 || len(param) < 1 || len(args) < 1 {
			t.Error("err")
			return
		}
		fmt.Println("AppendInsertf-->", fields, param, args)
		fields, param, args = AppendInsertf(fields, param, args, "int_ptr=$%v,int64_ptr=$%v,string_ptr=$%v#all", nil, converter.Int64Ptr(0), nil)
		if len(fields) < 1 || len(param) < 1 || len(args) < 1 {
			t.Error("err")
			return
		}
		fmt.Println("AppendInsertf-->", fields, param, args)
		insertSQL := fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, strings.Join(fields, ","), strings.Join(param, ","), "returning tid")
		fmt.Println("insertSQL-->", insertSQL)
		err = queryer.QueryRow(context.Background(), insertSQL, args...).Scan(&object.TID)
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		object := newTestObject()
		object.TID = 0
		_, err = InsertFilter(queryer, context.Background(), object, "^tid#all", "returning", "tid#all")
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
		object.TID = 0
		_, err = InsertFilter(queryer, context.Background(), object, "^tid", "returning", "tid#all")
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		object := newTestObject()
		_, err = InsertFilter(queryer, context.Background(), object, "^tid#all", "", "")
		if err != nil || object.TID > 0 {
			t.Error(err)
			return
		}
		_, err = InsertFilter(queryer, context.Background(), object, "^tid", "", "")
		if err != nil {
			t.Error(err)
			return
		}
		_, err = Default.InsertFilter(queryer, context.Background(), object, "^tid#all", "", "")
		if err != nil {
			t.Error(err)
			return
		}
	}
	{ //error
		object := newTestObject()
		object.TID = 0
		_, err = InsertFilter(queryer, context.Background(), object, "^tid#all", "xxxx", "tid#all")
		if err == nil {
			t.Error(err)
			return
		}
		_, err = InsertFilter(queryer, context.Background(), object, "", "xx xx", "")
		if err == nil {
			t.Error(err)
			return
		}
	}
}

func TestUpdate(t *testing.T) {
	clearPG()
	testUpdate(t, getPG())
}

func testUpdate(t *testing.T, queryer Queryer) {
	var err error
	object := newTestObject()
	{
		object.TID = 0
		object.Level = 1
		_, err = InsertFilter(queryer, context.Background(), object, "^tid#all", "returning", "tid#all")
		if err != nil || object.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		var updateSQL string
		var where []string
		var args []interface{}
		updateSQL, args = UpdateSQL(object, "title,image,update_time,status#all", nil)
		if strings.Contains(updateSQL, "tid") {
			err = fmt.Errorf("error")
			t.Error(err)
			return
		}
		where, args = AppendWhere(where, args, object.TID > 0, "tid=$%v", object.TID)
		where, args = AppendWhere(where, args, object.Level > 0, "level=$%v", object.Level)
		updateSQL = JoinWhere(updateSQL, where, "and", "")
		fmt.Printf("update\n")
		fmt.Printf("   --->%v\n", updateSQL)
		fmt.Printf("   --->%v\n", args)
		_, err = queryer.ExecRow(context.Background(), updateSQL, args...)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		updateSQL, args := UpdateSQL(object, "", nil)
		fmt.Printf("update\n")
		fmt.Printf("   --->%v\n", updateSQL)
		fmt.Printf("   --->%v\n", args)
		_, err = queryer.ExecRow(context.Background(), updateSQL, args...)
		if err != nil {
			t.Error(err)
			return
		}
		updateSQL, args = Default.UpdateSQL(object, "", nil)
		fmt.Printf("update\n")
		fmt.Printf("   --->%v\n", updateSQL)
		fmt.Printf("   --->%v\n", args)
		_, err = queryer.ExecRow(context.Background(), updateSQL, args...)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		sql, args := UpdateSQL(object, "", nil)
		if len(args) < 1 {
			err = fmt.Errorf("table error")
			t.Error(err)
			return
		}
		where, args := AppendWhere(nil, args, object.TID > 0, "tid=$%v", object.TID)
		where, args = AppendWhere(where, args, object.Level > 0, "level=$%v", object.Level)
		fmt.Println("AppendWhere-->", where, args)
		_, err = Update(queryer, context.Background(), object, sql, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		err = UpdateRow(queryer, context.Background(), object, sql, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		sql, args := Default.UpdateSQL(object, "", nil)
		if len(args) < 1 {
			err = fmt.Errorf("table error")
			t.Error(err)
			return
		}
		where, args := AppendWhere(nil, args, object.TID > 0, "tid=$%v", object.TID)
		where, args = AppendWhere(where, args, object.Level > 0, "level=$%v", object.Level)
		fmt.Println("AppendWhere-->", where, args)
		_, err = Default.Update(queryer, context.Background(), object, sql, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		err = Default.UpdateRow(queryer, context.Background(), object, sql, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		var where []string
		table, sets, args := UpdateArgs(object, "", nil)
		if table != "crud_object" || len(sets) < 1 || len(args) < 1 {
			err = fmt.Errorf("table error")
			t.Error(err)
			return
		}
		fmt.Println("UpdateArgs-->", sets, args)
		sets, args = AppendSet(sets, args, true, "int_value=$%v", 0)
		fmt.Println("AppendSet-->", sets, args)
		sets, args = AppendSetf(sets, args, "int64_value=$%v,string_value=$%v#all", 0, "")
		fmt.Println("AppendSet-->", sets, args)
		where, args = AppendWhere(where, args, object.TID > 0, "tid=$%v", object.TID)
		where, args = AppendWhere(where, args, object.Level > 0, "level=$%v", object.Level)
		fmt.Println("AppendWhere-->", where, args)
		_, err = UpdateSet(queryer, context.Background(), object, sets, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		err = UpdateSetRow(queryer, context.Background(), object, sets, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		var where []string
		table, sets, args := Default.UpdateArgs(object, "", nil)
		if table != "crud_object" || len(sets) < 1 || len(args) < 1 {
			err = fmt.Errorf("table error")
			t.Error(err)
			return
		}
		where, args = AppendWhere(where, args, object.TID > 0, "tid=$%v", object.TID)
		where, args = AppendWhere(where, args, object.Level > 0, "level=$%v", object.Level)
		_, err = Default.UpdateSet(queryer, context.Background(), object, sets, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		err = Default.UpdateSetRow(queryer, context.Background(), object, sets, where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		object.UpdateTime = xsql.TimeNow()
		where, args := AppendWhere(nil, nil, object.TID > 0, "tid=$%v", object.TID)
		_, err = UpdateFilter(queryer, context.Background(), object, "title,image,update_time,status", where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		object.UpdateTime = xsql.TimeNow()
		_, err = Default.UpdateFilter(queryer, context.Background(), object, "title,image,update_time,status", where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		object.UpdateTime = xsql.TimeNow()
		err = UpdateFilterRow(queryer, context.Background(), object, "title,image,update_time,status", where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
		object.UpdateTime = xsql.TimeNow()
		err = Default.UpdateFilterRow(queryer, context.Background(), object, "title,image,update_time,status", where, "and", args)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{
		_, err = UpdateWheref(queryer, context.Background(), object, "title,image,update_time,status", "tid=$%v", object.TID)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = Default.UpdateWheref(queryer, context.Background(), object, "title,image,update_time,status", "tid=$%v", object.TID)
		if err != nil {
			t.Error(err)
			return
		}
		err = UpdateRowWheref(queryer, context.Background(), object, "title,image,update_time,status", "tid=$%v", object.TID)
		if err != nil {
			t.Error(err)
			return
		}
		err = Default.UpdateRowWheref(queryer, context.Background(), object, "title,image,update_time,status", "tid=$%v", object.TID)
		if err != nil {
			t.Error(err)
			return
		}
	}
	{ //error
		object := newTestObject()
		object.TID = 0
		_, err = Update(queryer, context.Background(), object, "xx", nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		err = UpdateRow(queryer, context.Background(), object, "xx", nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		_, err = UpdateSet(queryer, context.Background(), object, nil, nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		err = UpdateSetRow(queryer, context.Background(), object, nil, nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		_, err = UpdateFilter(queryer, context.Background(), object, "xxstatus", nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		_, err = UpdateWheref(queryer, context.Background(), object, "status", "xxx=$1", -1)
		if err == nil {
			t.Error(err)
			return
		}
		err = UpdateFilterRow(queryer, context.Background(), object, "xxstatus", nil, "", nil)
		if err == nil {
			t.Error(err)
			return
		}
		_, sets, args := UpdateArgs(object, "", nil)
		where, args := AppendWhere(nil, args, true, "tid=$%v", -100)
		err = UpdateRow(queryer, context.Background(), object, `update crud_object set `+strings.Join(sets, ","), where, "and", args)
		if err != ErrNoRows {
			t.Error(err)
			return
		}
		err = UpdateSetRow(queryer, context.Background(), object, sets, where, "and", args)
		if err != ErrNoRows {
			t.Error(err)
			return
		}
		err = UpdateRowWheref(queryer, context.Background(), object, "title,image,update_time,status", "tid=$%v", -100)
		if err != ErrNoRows {
			t.Error(err)
			return
		}
		err = UpdateFilterRow(queryer, context.Background(), object, "title,image,update_time,status", []string{"tid=$1"}, "and", []interface{}{-100})
		if err != ErrNoRows {
			t.Error(err)
			return
		}
	}
}

func TestJoinWhere(t *testing.T) {
	clearPG()
	testJoinWhere(t, getPG())
}

func testJoinWhere(t *testing.T, queryer Queryer) {
	var err error
	object, _, _ := addTestMultiObject(queryer)
	sql := QuerySQL(object, "#all")
	{
		where, args := AppendWheref(nil, nil, "tid=$%v", object.TID)
		querySQL := JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", querySQL, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		where, args := AppendWheref(nil, nil, "tid=$%v", object.TID)
		querySQL := Default.JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", querySQL, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		querySQL, args := JoinWheref(sql, nil, "tid=$%v", object.TID)
		fmt.Println("JoinWhere-->", querySQL, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		querySQL, args := JoinWheref(sql, nil, "or:tid=$%v", object.TID)
		fmt.Println("JoinWhere-->", querySQL, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		querySQL, args := Default.JoinWheref(sql, nil, "or:tid=$%v", object.TID)
		fmt.Println("JoinWhere-->", querySQL, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
}

func TestJoinPage(t *testing.T) {
	clearPG()
	testJoinPage(t, getPG())
}

func testJoinPage(t *testing.T, queryer Queryer) {
	var err error
	object, _, _ := addTestMultiObject(queryer)
	sql := QuerySQL(object, "#all")
	{
		querySQL := JoinPage(sql, "order by tid asc", 0, 1)
		fmt.Println("JoinWhere-->", querySQL)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, nil, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		querySQL := Default.JoinPage(sql, "order by tid asc", 0, 1)
		fmt.Println("JoinWhere-->", querySQL)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", querySQL, nil, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
}

type RowsError struct {
	Rows
}

func (r *RowsError) Next() bool {
	return true
}

type ObjectInfoMap map[int64]string

func (u ObjectInfoMap) Scan(v interface{}) {
	s := v.(*CrudObject)
	u[s.TID] = fmt.Sprintf("%v-%v", s.Title, s.Level)
}

func TestQuery(t *testing.T) {
	clearPG()
	testQuery(t, getPG())
}

func testQuery(t *testing.T, queryer Queryer) {
	var err error
	object, _, _ := addTestMultiObject(queryer)
	{
		table, fields := QueryField(object, "#all")
		sql := fmt.Sprintf("select %v from %v", strings.Join(fields, ","), table)
		where, args := AppendWheref(nil, nil, "tid=$%v", object.TID)
		sql = JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", sql, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", sql, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "#all", "where tid=$1")
		args := []interface{}{object.TID}
		fmt.Println("JoinWhere-->", sql, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "#all", sql, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "o.#all", "where tid=$1")
		args := []interface{}{object.TID}
		fmt.Println("JoinWhere-->", sql, args)
		var result *CrudObject
		err = QueryRow(queryer, context.Background(), object, "o.#all", sql, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "#all")
		where, args := AppendWheref(nil, nil, "tid=$%v", object.TID)
		sql = JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", sql, args)
		var result *CrudObject
		//
		result = nil
		err = QueryRow(queryer, context.Background(), object, "#all", sql, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
		result = nil
		err = QueryFilterRow(queryer, context.Background(), object, "#all", where, "and", args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
		result = nil
		err = QueryRowWheref(queryer, context.Background(), object, "#all", "tid=$%v", []interface{}{object.TID}, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
		result = nil
		err = Default.QueryRow(queryer, context.Background(), object, "#all", sql, args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
		result = nil
		err = Default.QueryFilterRow(queryer, context.Background(), object, "#all", where, "and", args, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
		result = nil
		err = Default.QueryRowWheref(queryer, context.Background(), object, "#all", "tid=$%v", []interface{}{object.TID}, &result)
		if err != nil || result.TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "#all")
		where, args := AppendWheref(nil, nil, "tid=$%v", object.TID)
		sql = JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", sql, args)
		var results []*CrudObject
		err = Query(queryer, context.Background(), object, "#all", sql, args, &results)
		if err != nil || len(results) != 1 || results[0].TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "o.#all")
		where, args := AppendWheref(nil, nil, "o.tid=$%v", object.TID)
		sql = JoinWhere(sql, where, "and", "limit 1")
		fmt.Println("JoinWhere-->", sql, args)
		var results []*CrudObject
		err = Query(queryer, context.Background(), object, "o.#all", sql, args, &results)
		if err != nil || len(results) != 1 || results[0].TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := Default.QuerySQL(object, "#all")
		sql, args := JoinWheref(sql, nil, "tid=$%v", object.TID)
		fmt.Println("JoinWheref-->", sql, args)
		var results []*CrudObject
		err = Query(queryer, context.Background(), object, "#all", sql, args, &results)
		if err != nil || len(results) != 1 || results[0].TID < 1 {
			t.Error(err)
			return
		}
	}
	{
		sql := QuerySQL(object, "#all")
		sql = JoinPage(sql, "order by tid desc", 0, 1)
		where, args := AppendWheref(nil, nil, "level>$%v#all", 0)
		fmt.Println("JoinPage-->", sql)
		var results []*CrudObject
		//
		results = nil
		err = Query(queryer, context.Background(), object, "#all", sql, nil, &results)
		if err != nil || len(results) != 1 || results[0].TID < 1 {
			t.Error(err)
			return
		}
		results = nil
		err = QueryFilter(queryer, context.Background(), object, "#all", where, "and", args, "", 0, 0, &results)
		if err != nil || len(results) != 11 {
			t.Error(err)
			return
		}
		results = nil
		err = QueryWheref(queryer, context.Background(), object, "#all", "level>$%v#all", []interface{}{0}, 0, 0, &results)
		if err != nil || len(results) != 11 {
			t.Error(err)
			return
		}
		results = nil
		err = Default.Query(queryer, context.Background(), object, "#all", sql, nil, &results)
		if err != nil || len(results) != 1 || results[0].TID < 1 {
			t.Error(err)
			return
		}
		results = nil
		err = Default.QueryFilter(queryer, context.Background(), object, "#all", where, "and", args, "", 0, 0, &results)
		if err != nil || len(results) != 11 {
			t.Error(err)
			return
		}
		results = nil
		err = Default.QueryWheref(queryer, context.Background(), object, "#all", "level>$%v#all", []interface{}{0}, 0, 0, &results)
		if err != nil || len(results) != 11 {
			t.Error(err)
			return
		}
	}
	{
		var resultList []*CrudObject
		var result *CrudObject
		var int64IDs0, int64IDs1 []int64
		var images0, images1 []*string
		var resultMap0 map[int64]*CrudObject
		var resultMap1 = map[int64][]*CrudObject{}
		var int64Map0 map[int64]int64
		var int64Map1 map[int64][]int64
		var infoMap = ObjectInfoMap{}
		err = QueryWheref(
			queryer, context.Background(), object, "#all", "", nil, 0, 0,
			&resultList, &result,
			&int64IDs0, "int64_value",
			&int64IDs1, "int64_value#all",
			&images0, "image",
			&images1, "image#all",
			&resultMap0, "tid",
			&resultMap1, "int64_value",
			&int64Map0, "tid:int64_value",
			&int64Map1, "int64_value:tid",
			&infoMap,
			func(v *CrudObject) {
				// objects1[v.TID] = v
			},
		)
		if err != nil {
			t.Error(err)
			return
		}
		if len(resultList) < 1 || result == nil || len(int64IDs0) < 1 || len(int64IDs1) < 1 ||
			len(images0) < 1 || len(images1) < 1 || len(resultMap0) < 1 || len(resultMap1) < 1 ||
			len(int64Map0) < 1 || len(int64Map1) < 1 || len(infoMap) < 1 {
			fmt.Printf(`
				len(resultList)=%v, result=%v, len(int64IDs0)=%v, len(int64IDs1)=%v,
				len(images0)=%v, len(images1)=%v, len(resultMap0)=%v, len(resultMap1)=%v,
				len(int64Map0)=%v, len(int64Map1)=%v, len(infoMap)=%v
			`,
				len(resultList), result == nil, len(int64IDs0), len(int64IDs1),
				len(images0), len(images1), len(resultMap0), len(resultMap1),
				len(int64Map0), len(int64Map1), len(infoMap),
			)
			t.Error("data error")
			return
		}
	}
	{
		var idList []int64
		err = Query(queryer, context.Background(), int64(0), "tid#all", "select tid from crud_object", nil, &idList)
		if err != nil || len(idList) != 11 {
			t.Errorf("%v,%v", err, idList)
			return
		}
		var intList []int
		err = Query(queryer, context.Background(), int(0), "#all", "select int_value from crud_object", nil, &intList)
		if err != nil || len(intList) != 11 {
			t.Errorf("%v,%v", err, intList)
			return
		}
		var intPtrList []*int
		err = Query(queryer, context.Background(), converter.IntPtr(0), "#all", "select int_ptr from crud_object", nil, &intPtrList)
		if err != nil || len(intPtrList) != 11 {
			t.Errorf("%v,%v", err, intPtrList)
			return
		}
		var int64List []int64
		err = Query(queryer, context.Background(), int64(0), "#all", "select int64_value from crud_object", nil, &int64List)
		if err != nil || len(int64List) != 11 {
			t.Errorf("%v,%v", err, int64List)
			return
		}
		int64List = nil
		err = Query(queryer, context.Background(), object, "int64_value#all", "select int64_value from crud_object", nil, &int64List, "int64_value")
		if err != nil || len(int64List) != 7 {
			t.Errorf("%v,%v", err, int64List)
			return
		}
		int64List = nil
		err = Query(queryer, context.Background(), object, "int64_value#all", "select int64_value from crud_object", nil, &int64List, "int64_value#all")
		if err != nil || len(int64List) != 11 {
			t.Errorf("%v,%v", err, int64List)
			return
		}
		var int64PtrList []*int64
		err = Query(queryer, context.Background(), converter.Int64Ptr(0), "#all", "select int64_ptr from crud_object", nil, &int64PtrList)
		if err != nil || len(int64PtrList) != 11 {
			t.Errorf("%v,%v", err, int64PtrList)
			return
		}
		int64PtrList = nil
		err = Query(queryer, context.Background(), object, "int64_ptr#all", "select int64_ptr from crud_object", nil, &int64PtrList, "int64_ptr")
		if err != nil || len(int64PtrList) != 1 {
			t.Errorf("%v,%v", err, int64PtrList)
			return
		}
		int64PtrList = nil
		err = Query(queryer, context.Background(), object, "int64_ptr#all", "select int64_ptr from crud_object", nil, &int64PtrList, "int64_ptr#all")
		if err != nil || len(int64PtrList) != 11 {
			t.Errorf("%v,%v", err, int64PtrList)
			return
		}
		var float64List []float64
		err = Query(queryer, context.Background(), float64(0), "#all", "select float64_value from crud_object", nil, &float64List)
		if err != nil || len(float64List) != 11 {
			t.Errorf("%v,%v", err, float64List)
			return
		}
		var float64PtrList []*float64
		err = Query(queryer, context.Background(), converter.Float64Ptr(0), "#all", "select float64_ptr from crud_object", nil, &float64PtrList)
		if err != nil || len(float64List) != 11 {
			t.Errorf("%v,%v", err, float64List)
			return
		}
		var stringList []string
		err = Query(queryer, context.Background(), string(""), "#all", "select string_value from crud_object", nil, &stringList)
		if err != nil || len(stringList) != 11 {
			t.Errorf("%v,%v", err, stringList)
			return
		}
		var stringPtrList []*string
		err = Query(queryer, context.Background(), converter.StringPtr(""), "#all", "select string_ptr from crud_object", nil, &stringPtrList)
		if err != nil || len(stringPtrList) != 11 {
			t.Errorf("%v,%v", err, stringPtrList)
			return
		}
	}
	{
		var idResult int64
		err = QueryRow(queryer, context.Background(), int64(0), "#all", "select tid from crud_object where tid=$1", []interface{}{object.TID}, &idResult)
		if err != nil || idResult < 1 {
			t.Errorf("%v,%v", err, idResult)
			return
		}
		var intResult int
		err = QueryRow(queryer, context.Background(), int(0), "#all", "select int_value from crud_object where tid=$1", []interface{}{object.TID}, &intResult)
		if err != nil || intResult < 1 {
			t.Errorf("%v,%v", err, intResult)
			return
		}
		var intPtrResult *int
		err = QueryRow(queryer, context.Background(), converter.IntPtr(0), "#all", "select int_ptr from crud_object where tid=$1", []interface{}{object.TID}, &intPtrResult)
		if err != nil || intPtrResult == nil {
			t.Errorf("%v,%v", err, intPtrResult)
			return
		}
		var int64Result int64
		err = QueryRow(queryer, context.Background(), int64(0), "#all", "select int64_value from crud_object where tid=$1", []interface{}{object.TID}, &int64Result)
		if err != nil || int64Result < 1 {
			t.Errorf("%v,%v", err, int64Result)
			return
		}
		var int64PtrResult *int64
		err = QueryRow(queryer, context.Background(), converter.Int64Ptr(0), "#all", "select int64_ptr from crud_object where tid=$1", []interface{}{object.TID}, &int64PtrResult)
		if err != nil || int64PtrResult == nil {
			t.Errorf("%v,%v", err, int64PtrResult)
			return
		}
		var float64Result float64
		err = QueryRow(queryer, context.Background(), float64(0), "#all", "select float64_value from crud_object where tid=$1", []interface{}{object.TID}, &float64Result)
		if err != nil || float64Result < 1 {
			t.Errorf("%v,%v", err, float64Result)
			return
		}
		var float64PtrResult *float64
		err = QueryRow(queryer, context.Background(), converter.Float64Ptr(0), "#all", "select float64_ptr from crud_object where tid=$1", []interface{}{object.TID}, &float64PtrResult)
		if err != nil || float64Result < 1 {
			t.Errorf("%v,%v", err, float64Result)
			return
		}
		var stringResult string
		err = QueryRow(queryer, context.Background(), string(""), "#all", "select string_value from crud_object where tid=$1", []interface{}{object.TID}, &stringResult)
		if err != nil || len(stringResult) < 1 {
			t.Errorf("%v,%v", err, stringResult)
			return
		}
		var stringPtrResult *string
		err = QueryRow(queryer, context.Background(), converter.StringPtr(""), "#all", "select string_ptr from crud_object where tid=$1", []interface{}{object.TID}, &stringPtrResult)
		if err != nil || stringPtrResult == nil {
			t.Errorf("%v,%v", err, stringPtrResult)
			return
		}
	}
	{
		var idResult int64
		err = QueryRowWheref(queryer, context.Background(), MetaWith(object, int64(0)), "tid#all", "tid=$%v", []interface{}{object.TID}, &idResult, "tid")
		if err != nil || idResult < 1 {
			t.Errorf("%v,%v", err, idResult)
			return
		}
		var stringResult string
		err = QueryRowWheref(queryer, context.Background(), MetaWith(object, string("")), "string_value#all", "tid=$%v", []interface{}{object.TID}, &stringResult, "string_value")
		if err != nil || len(stringResult) < 1 {
			t.Errorf("%v,%v", err, stringResult)
			return
		}
		var stringPtrResult *string
		err = QueryRowWheref(queryer, context.Background(), MetaWith(object, converter.StringPtr("")), "string_ptr#all", "tid=$%v", []interface{}{object.TID}, &stringPtrResult, "string_ptr")
		if err != nil || stringPtrResult == nil {
			t.Errorf("%v,%v", err, stringPtrResult)
			return
		}
	}
	{ //alias
		var idList []int64
		err = Query(queryer, context.Background(), MetaWith(object, int64(0)), "o.tid#all", "select o.tid from crud_object o", nil, &idList, "tid")
		if err != nil || len(idList) != 11 {
			t.Errorf("%v,%v", err, idList)
			return
		}
	}
	{
		querySQL := QuerySQL(object, "#all")
		rows, err := queryer.Query(context.Background(), querySQL)
		if err != nil {
			t.Error(err)
			return
		}
		for rows.Next() {
			obj := &CrudObject{}
			args := ScanArgs(obj, "#all")
			err = rows.Scan(args...)
			if err != nil || obj.TID < 1 {
				t.Error(err)
				return
			}
		}
		rows.Close()
	}
	{
		querySQL := QuerySQL(object, "#all")
		rows, err := queryer.Query(context.Background(), querySQL)
		if err != nil {
			t.Error(err)
			return
		}
		for rows.Next() {
			var obj *CrudObject
			err = Scan(rows, &CrudObject{}, "#all", &obj)
			if err != nil || obj == nil {
				t.Error(err)
				return
			}
		}
		rows.Close()
		var obj *CrudObject
		err = Scan(&RowsError{Rows: rows}, &CrudObject{}, "#all", &obj)
		if err == nil {
			t.Error(err)
			return
		}
	}
	{ //error
		var result *CrudObject
		var results []*CrudObject

		sql := QuerySQL(object, "#all", "xxx xxx")
		err = Query(queryer, context.Background(), object, "#all", sql, nil, &results)
		if err == nil {
			t.Error(err)
			return
		}

		where, args := AppendWheref(nil, nil, "lexxvel>$%v#all", 0)
		err = QueryFilter(queryer, context.Background(), object, "#all", where, "and", args, "", 0, 0, &results)
		if err == nil {
			t.Error(err)
			return
		}

		err = QueryWheref(queryer, context.Background(), object, "#all", "levxxel>$%v#all", []interface{}{0}, 0, 0, &results)
		if err == nil {
			t.Error(err)
			return
		}

		err = QueryFilterRow(queryer, context.Background(), object, "#all", where, "and", args, &result)
		if err == nil {
			t.Error(err)
			return
		}
	}
	{ //error
		var idList []int
		var idResult int
		err = Query(queryer, context.Background(), int64(0), "o.tid#all", "select o.tid from crud_object o", nil, &idList)
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = Query(queryer, context.Background(), int64(0), "o.tid#all", "select o.tid from crud_object o", nil, &idList, "tid") //not struct error
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = Query(queryer, context.Background(), int64(0), "o.tid#all", "select o.tid from crud_object o", nil, &idList, &idList)
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = Query(queryer, context.Background(), int64(0), "o.tid#all", "select o.tid from crud_object o", nil, &idList, "")
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = Query(queryer, context.Background(), MetaWith(object, int64(0)), "o.tid#all", "select o.tid from crud_object o", nil, &idList, "xxx") //not field
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = Query(queryer, context.Background(), object, "o.tid#all", "select o.tid from crud_object o", nil, &idList, "tid") //type not supported
		if err == nil {
			t.Errorf("%v,%v", err, idList)
			return
		}
		err = QueryRow(queryer, context.Background(), int64(0), "#all", "select tid from crud_object where tid=$1", []interface{}{object.TID}, &idResult)
		if err == nil {
			t.Errorf("%v,%v", err, idResult)
			return
		}
	}
	{ //error
		var idMap map[int64]int
		err = Query(queryer, context.Background(), object, "tid,int_value#all", "select tid,int_value from crud_object", nil, &idMap)
		if err == nil {
			t.Errorf("%v,%v", err, idMap)
			return
		}
		err = Query(queryer, context.Background(), object, "tid,int_value#all", "select tid,int_value from crud_object", nil, &idMap, &idMap)
		if err == nil {
			t.Errorf("%v,%v", err, idMap)
			return
		}
		err = Query(queryer, context.Background(), object, "tid,int_value#all", "select tid,int_value from crud_object", nil, &idMap, "")
		if err == nil {
			t.Errorf("%v,%v", err, idMap)
			return
		}
		err = Query(queryer, context.Background(), object, "tid,int_value#all", "select tid,int_value from crud_object", nil, &idMap, "xxx")
		if err == nil {
			t.Errorf("%v,%v", err, idMap)
			return
		}
		err = Query(queryer, context.Background(), object, "tid,int_value#all", "select tid,int_value from crud_object", nil, &idMap, "tid:xxx")
		if err == nil {
			t.Errorf("%v,%v", err, idMap)
			return
		}
	}
}

func TestCount(t *testing.T) {
	clearPG()
	testCount(t, getPG())
}

func testCount(t *testing.T, queryer Queryer) {
	var err error
	object, _, _ := addTestMultiObject(queryer)
	{
		var countValue int64
		var countSQL string

		countSQL = CountSQL(object, "")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
		countSQL = CountSQL(object, "*")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
		countSQL = CountSQL(object, "count(*)")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
		countSQL = CountSQL(object, "count(*)#all")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
		countSQL = CountSQL(object, "count(tid)")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
		countSQL = CountSQL(object, "count(tid)#all")
		err = Count(queryer, context.Background(), int64(0), "#all", countSQL, nil, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}

		countSQL = Default.CountSQL(object, "count(tid)#all", "where tid>$1")
		err = Default.Count(queryer, context.Background(), int64(0), "#all", countSQL, []interface{}{0}, &countValue)
		if err != nil || countValue != 11 {
			t.Error(err)
			return
		}
	}
	{
		var countVal int64
		err = CountWheref(queryer, context.Background(), object, "count(tid)#all", "int_value>$1#all", []interface{}{0}, "", &countVal, "tid")
		if err != nil || countVal != 7 {
			t.Errorf("%v,%v", err, countVal)
			return
		}
		err = CountWheref(queryer, context.Background(), MetaWith(object, int64(0)), "count(tid)#all", "int_value>$1#all", []interface{}{0}, "", &countVal, "tid")
		if err != nil || countVal != 7 {
			t.Errorf("%v,%v", err, countVal)
			return
		}

		err = Default.CountWheref(queryer, context.Background(), object, "count(tid)#all", "int_value>$1#all", []interface{}{0}, " ", &countVal, "tid")
		if err != nil || countVal != 7 {
			t.Errorf("%v,%v", err, countVal)
			return
		}
	}
	{
		var countVal int64
		err = CountFilter(queryer, context.Background(), object, "count(tid)#all", nil, "", nil, "", &countVal, "tid")
		if err != nil || countVal != 11 {
			t.Error(err)
			return
		}
		err = Default.CountFilter(queryer, context.Background(), object, "count(tid)#all", nil, "", nil, "", &countVal, "tid")
		if err != nil || countVal != 11 {
			t.Error(err)
			return
		}
	}
	{ //error
		var countVal int64
		err = CountFilter(queryer, context.Background(), object, "abc(tid)#all", nil, "", nil, "", &countVal, "tid")
		if err == nil {
			t.Error(err)
			return
		}
	}
}

type SearchCrudObjectUnify struct {
	Model CrudObject `json:"model"`
	Where struct {
		UserID int64          `json:"user_id"`
		Type   CrudObjectType `json:"type" filter:"#all"`
		Key    struct {
			Title string `json:"title" cmp:"title like $%v"`
			Data  string `json:"data" cmp:"data::text like $%v"`
		} `json:"key" join:"or"`
		Status CrudObjectStatusArray `json:"status" cmp:"status=any($%v)"`
		Ignore int                   `json:"ignore" cmp:"-" filter:"#all"`
	} `json:"where" join:"and"`
	Page struct {
		Order  string `json:"order" default:"order by tid desc"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	} `json:"page"`
	Query struct {
		Enabled bool          `json:"enabled" scan:"-"`
		Objects []*CrudObject `json:"objects"`
		UserIDs []int64       `json:"user_ids" scan:"user_id#all"`
	} `json:"query" filter:"#all"`
	Count struct {
		Enabled bool  `json:"enabled" scan:"-"`
		All     int64 `json:"all" scan:"tid"`
		UserID  int64 `json:"user_id" scan:"user_id"`
	} `json:"count" filter:"count(tid),max(user_id)#all"`
}

type FindCrudObjectUnify struct {
	Model CrudObject `json:"model"`
	Where struct {
		UserID int64                 `json:"user_id" cmp:"user_id="`
		Type   CrudObjectType        `json:"type" filter:"#all"`
		Key    string                `json:"key" cmp:"title like $%v or data::text like $%v"`
		Status CrudObjectStatusArray `json:"status" cmp:"status=any($%v)"`
		Ignore int                   `json:"ignore" cmp:"-" filter:"#all"`
		Empty  int                   `json:"empty"`
	} `json:"where" join:"and"`
	QueryRow struct {
		Enabled bool        `json:"enabled" scan:"-"`
		Object  *CrudObject `json:"object"`
		UserID  int64       `json:"user_id" scan:"user_id#all"`
	} `json:"query_row" filter:"#all"`
}

func TestUnify(t *testing.T) {
	clearPG()
	testUnify(t, getPG())
}

func testUnify(t *testing.T, queryer Queryer) {
	var err error
	addTestMultiObject(queryer)
	newSearch := func() *SearchCrudObjectUnify {
		search := &SearchCrudObjectUnify{}
		search.Where.UserID = 100
		search.Where.Type = CrudObjectTypeA
		search.Where.Key.Title = "%a%"
		search.Where.Key.Data = "%a%"
		search.Where.Status = CrudObjectStatusShow
		search.Query.Enabled = true
		search.Count.Enabled = true
		search.Page.Offset = 0
		search.Page.Limit = 100
		return search
	}
	newFind := func() *FindCrudObjectUnify {
		find := &FindCrudObjectUnify{}
		find.Where.UserID = 100
		find.Where.Type = CrudObjectTypeA
		find.Where.Key = "%a%"
		find.Where.Status = CrudObjectStatusShow
		find.QueryRow.Enabled = true
		return find
	}
	{
		search := newSearch()
		err = ApplyUnify(queryer, context.Background(), search)
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
		find := newFind()
		err = Default.ApplyUnify(queryer, context.Background(), find)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		err = QueryUnify(queryer, context.Background(), search)
		if err != nil {
			t.Error(err)
			return
		}
		err = CountUnify(queryer, context.Background(), search)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		err = Default.QueryUnify(queryer, context.Background(), search)
		if err != nil {
			t.Error(err)
			return
		}
		err = Default.CountUnify(queryer, context.Background(), search)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		err = QueryUnifyRow(queryer, context.Background(), find)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		err = Default.QueryUnifyRow(queryer, context.Background(), find)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		querySQL, queryArgs := QueryUnifySQL(search, "Query")
		rows, err := queryer.Query(context.Background(), querySQL, queryArgs...)
		if err != nil {
			t.Error(err)
			return
		}
		err = ScanUnify(rows, search)
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		countSQL, countArgs := CountUnifySQL(search)
		row := queryer.QueryRow(context.Background(), countSQL, countArgs...)
		modelValue, queryFilter, dests := CountUnifyDest(search)
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		querySQL, queryArgs := Default.QueryUnifySQL(search, "Query")
		rows, err := queryer.Query(context.Background(), querySQL, queryArgs...)
		if err != nil {
			t.Error(err)
			return
		}
		err = Default.ScanUnify(rows, search)
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		countSQL, countArgs := Default.CountUnifySQL(search)
		row := queryer.QueryRow(context.Background(), countSQL, countArgs...)
		modelValue, queryFilter, dests := Default.CountUnifyDest(search)
		err = Default.ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		querySQL := QuerySQL(&search.Model, "#all")
		querySQL, queryArgs := JoinWhereUnify(querySQL, nil, search)
		querySQL = JoinPageUnify(querySQL, search)
		rows, err := queryer.Query(context.Background(), querySQL, queryArgs...)
		if err != nil {
			t.Error(err)
			return
		}
		err = ScanUnify(rows, search)
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		countSQL, countArgs := CountUnifySQL(search)
		row := queryer.QueryRow(context.Background(), countSQL, countArgs...)
		modelValue, queryFilter, dests := CountUnifyDest(search)
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		search := newSearch()
		querySQL := Default.QuerySQL(&search.Model, "#all")
		querySQL, queryArgs := Default.JoinWhereUnify(querySQL, nil, search)
		querySQL = Default.JoinPageUnify(querySQL, search)
		rows, err := queryer.Query(context.Background(), querySQL, queryArgs...)
		if err != nil {
			t.Error(err)
			return
		}
		err = ScanUnify(rows, search)
		if err != nil {
			t.Error(err)
			return
		}
		rows.Close()
		countSQL, countArgs := Default.CountUnifySQL(search)
		row := queryer.QueryRow(context.Background(), countSQL, countArgs...)
		modelValue, queryFilter, dests := Default.CountUnifyDest(search)
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil || len(search.Query.Objects) < 1 || len(search.Query.UserIDs) < 1 || search.Count.All < 1 || search.Count.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		querySQL, queryArgs := Default.QueryUnifySQL(find, "QueryRow")
		row := queryer.QueryRow(context.Background(), querySQL, queryArgs...)
		err = ScanUnifyRow(row, find)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		querySQL, queryArgs := Default.QueryUnifySQL(find, "QueryRow")
		row := queryer.QueryRow(context.Background(), querySQL, queryArgs...)
		modelValue, queryFilter, dests := ScanUnifyDest(find, "QueryRow")
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		querySQL := QuerySQL(&find.Model, "#all")
		querySQL, queryArgs := JoinWhereUnify(querySQL, nil, find)
		row := queryer.QueryRow(context.Background(), querySQL, queryArgs...)
		modelValue, queryFilter, dests := ScanUnifyDest(find, "QueryRow")
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		querySQL := QuerySQL(&find.Model, "#all")
		querySQL, queryArgs := Default.JoinWhereUnify(querySQL, nil, find)
		row := queryer.QueryRow(context.Background(), querySQL, queryArgs...)
		modelValue, queryFilter, dests := ScanUnifyDest(find, "QueryRow")
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
	{
		find := newFind()
		querySQL := QuerySQL(&find.Model, "#all")
		where, queryArgs := AppendWhereUnify(nil, nil, find)
		querySQL = JoinWhere(querySQL, where, "and")
		row := queryer.QueryRow(context.Background(), querySQL, queryArgs...)
		modelValue, queryFilter, dests := ScanUnifyDest(find, "QueryRow")
		err = ScanRow(row, modelValue, queryFilter, dests...)
		if err != nil || find.QueryRow.Object == nil || find.QueryRow.UserID < 1 {
			t.Error(err)
			return
		}
	}
}

type ErrorCrudObjectUnify struct {
	Model CrudObject `json:"model"`
	Where struct {
		UserID int64                 `json:"user_id"`
		Type   CrudObjectType        `json:"type" filter:"#all"`
		Status CrudObjectStatusArray `json:"status" cmp:"statusxx=any($%v)"`
	} `json:"where" join:"and"`
	Page struct {
		Order  string `json:"order" default:"order by tid desc"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	} `json:"page"`
	Query struct {
		Enabled bool          `json:"enabled" scan:"-"`
		Objects []*CrudObject `json:"objects"`
		UserIDs []int64       `json:"user_ids" scan:"user_id#all"`
	} `json:"query" filter:"#all"`
	Count struct {
		Enabled bool  `json:"enabled" scan:"-"`
		All     int64 `json:"all" scan:"tid"`
		UserID  int64 `json:"user_id" scan:"user_id"`
	} `json:"count" filter:"count(tid),max(user_id)#all"`
	QueryRow struct {
		Enabled bool        `json:"enabled" scan:"-"`
		Object  *CrudObject `json:"object"`
		UserID  int64       `json:"user_id" scan:"user_id#all"`
	} `json:"query_row" filter:"#all"`
}

func TestUnifyError(t *testing.T) {
	clearPG()
	testUnifyError(t, getPG())
}

func testUnifyError(t *testing.T, queryer Queryer) {
	var err error
	search := &ErrorCrudObjectUnify{}
	search.Where.UserID = 100
	search.Where.Type = CrudObjectTypeA
	search.Where.Status = CrudObjectStatusShow
	search.Query.Enabled = true
	search.Count.Enabled = true
	search.Page.Offset = 0
	search.Page.Limit = 100
	err = QueryUnify(queryer, context.Background(), search)
	if err == nil {
		t.Error(err)
		return
	}
	err = QueryUnifyRow(queryer, context.Background(), search)
	if err == nil {
		t.Error(err)
		return
	}
	err = CountUnify(queryer, context.Background(), search)
	if err == nil {
		t.Error(err)
		return
	}
	func() {
		defer func() {
			recover()
		}()
		ScanUnifyDest(search, "Abc")
		t.Error("eror")
	}()
}
