//Package crud is simple curd tools to process database
package crud

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/codingeasygo/util/attrscan"
)

type NilChecker interface {
	IsNil() bool
}

type ZeroChecker interface {
	IsZero() bool
}

func jsonString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func NewValue(v interface{}) (value reflect.Value) {
	if v, ok := v.([]interface{}); ok {
		result := []interface{}{}
		for i, f := range v {
			if _, ok := f.(TableName); ok {
				continue
			}
			item := reflect.New(reflect.ValueOf(v[i]).Type())
			result = append(result, item.Interface())
		}
		value = reflect.ValueOf(result)
		return
	}
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	value = reflect.New(reflectType)
	return
}

func MetaWith(o interface{}, fields ...interface{}) (v []interface{}) {
	strName, ok := o.(string)
	if !ok {
		strName = Table(o)
	}
	v = append(v, TableName(strName))
	v = append(v, fields...)
	return
}

var Default = &CRUD{
	Scanner: attrscan.Scanner{
		Tag: "json",
		NameConv: func(on, name string, field reflect.StructField) string {
			return name
		},
	},
	ArgFormat: "$%v",
	ErrNoRows: fmt.Errorf("no rows"),
	Log: func(caller int, format string, args ...interface{}) {
		log.Output(caller+3, fmt.Sprintf(format, args...))
	},
	ParmConv: func(on, fieldName, fieldFunc string, field reflect.StructField, value interface{}) interface{} {
		return value
	},
}

type NameConv func(on, name string, field reflect.StructField) string
type ParmConv func(on, fieldName, fieldFunc string, field reflect.StructField, value interface{}) interface{}
type LogF func(caller int, format string, args ...interface{})
type TableName string

type CRUD struct {
	attrscan.Scanner
	ArgFormat string
	ErrNoRows error
	Verbose   bool
	Log       LogF
	ParmConv  ParmConv
}

func Table(v interface{}) (table string) {
	table = Default.Table(v)
	return
}

func (c *CRUD) Table(v interface{}) (table string) {
	if v, ok := v.([]interface{}); ok {
		for _, f := range v {
			if tableName, ok := f.(TableName); ok {
				table = string(tableName)
				break
			}
		}
		return
	}
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	numField := reflectType.NumField()
	for i := 0; i < numField; i++ {
		fieldType := reflectType.Field(i)
		if fieldType.Name == "_" {
			if t := fieldType.Tag.Get("table"); len(t) > 0 {
				table = t
				break
			}
		}
	}
	return
}

func (c *CRUD) Sprintf(format string, v int) string {
	args := []interface{}{}
	arg := fmt.Sprintf("%d", v)
	n := strings.Count(format, c.ArgFormat)
	for i := 0; i < n; i++ {
		args = append(args, arg)
	}
	return fmt.Sprintf(format, args...)
}

func FilterFieldCall(on string, v interface{}, filter string, call func(fieldName, fieldFunc string, field reflect.StructField, value interface{})) (table string) {
	table = Default.FilterFieldCall(on, v, filter, call)
	return
}

func (c *CRUD) FilterFieldCall(on string, v interface{}, filter string, call func(fieldName, fieldFunc string, field reflect.StructField, value interface{})) (table string) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	if v, ok := v.([]interface{}); ok {
		var tableAlias, fieldAlias string
		parts := strings.SplitN(filter, ".", 2)
		if len(parts) > 1 {
			tableAlias = parts[0]
			fieldAlias = parts[0] + "."
			filter = parts[1]
		}
		filterFields := strings.Split(strings.TrimSpace(strings.SplitN(filter, "#", 2)[0]), ",")
		offset := 0
		for _, f := range v {
			if tableName, ok := f.(TableName); ok {
				table = string(tableName)
				if len(tableAlias) > 0 {
					table = table + " " + tableAlias
				}
				continue
			}
			if offset >= len(filterFields) {
				panic(fmt.Sprintf("meta v[%v] is not found on filter", offset))
			}
			fieldParts := strings.SplitN(strings.Trim(filterFields[offset], ")"), "(", 2)
			fieldName := fieldParts[0]
			fieldFunc := ""
			if len(fieldParts) > 1 {
				fieldName = fieldParts[1]
				fieldFunc = fieldParts[0]
			}
			call(fieldAlias+fieldName, fieldFunc, reflect.StructField{}, f)
			offset++
		}
		return
	}
	if reflectType.Kind() != reflect.Struct {
		var fieldAlias string
		parts := strings.SplitN(filter, ".", 2)
		if len(parts) > 1 {
			fieldAlias = parts[0] + "."
			filter = parts[1]
		}
		fieldParts := strings.SplitN(strings.Trim(strings.TrimSpace(strings.SplitN(filter, "#", 2)[0]), ")"), "(", 2)
		fieldName := fieldParts[0]
		fieldFunc := ""
		if len(fieldParts) > 1 {
			fieldName = fieldParts[1]
			fieldFunc = fieldParts[0]
		}
		call(fieldAlias+fieldName, fieldFunc, reflect.StructField{}, v)
		return
	}
	table = c.Table(v)
	parts := strings.SplitN(filter, ".", 2)
	if len(parts) > 1 {
		table = table + " " + parts[0]
	}
	c.Scanner.FilterFieldCall(on, v, filter, call)
	return
}

func FilterFormatCall(formats string, args []interface{}, call func(format string, arg interface{})) {
	Default.FilterFormatCall(formats, args, call)
}

func (c *CRUD) FilterFormatCall(formats string, args []interface{}, call func(format string, arg interface{})) {
	formatParts := strings.SplitN(formats, "#", 2)
	var incNil, incZero bool
	if len(formatParts) > 1 && len(formatParts[1]) > 0 {
		incNil = strings.Contains(","+formatParts[1]+",", ",nil,") || strings.Contains(","+formatParts[1]+",", ",all,")
		incZero = strings.Contains(","+formatParts[1]+",", ",zero,") || strings.Contains(","+formatParts[1]+",", ",all,")
	}
	formatList := strings.Split(formatParts[0], ",")
	if len(formatList) != len(args) {
		panic(fmt.Sprintf("count formats=%v  is not equal to args=%v", len(formatList), len(args)))
	}
	for i, format := range formatList {
		arg := args[i]
		if !c.Scanner.CheckValue(reflect.ValueOf(arg), incNil, incZero) {
			continue
		}
		call(format, arg)
	}
}

func (c *CRUD) FilterWhere(args []interface{}, v interface{}) (where_ []string, args_ []interface{}) {
	args_ = args
	c.FilterFieldCall("where", v, "", func(fieldName, fieldFunc string, field reflect.StructField, fieldValue interface{}) {
		if field.Type.Kind() == reflect.Struct {
			var cmpInner []string
			cmpInner, args_ = c.FilterWhere(args_, fieldValue)
			cmpSep := field.Tag.Get("join")
			where_ = append(where_, strings.Join(cmpInner, " "+cmpSep+" "))
			return
		}
		cmp := field.Tag.Get("cmp")
		if len(cmp) < 1 {
			cmp = "= " + c.ArgFormat
		}
		if !strings.Contains(cmp, c.ArgFormat) {
			cmp += " " + c.ArgFormat
		}
		args_ = append(args_, c.ParmConv("where", fieldName, fieldFunc, field, fieldValue))
		where_ = append(where_, c.Sprintf(fieldName+" "+cmp, len(args_)))
	})
	return
}

func AppendInsert(fields, param []string, args []interface{}, ok bool, format string, v interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = Default.AppendInsert(fields, param, args, ok, format, v)
	return
}

func (c *CRUD) AppendInsert(fields, param []string, args []interface{}, ok bool, format string, v interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = fields, param, args
	if ok {
		args_ = append(args_, c.ParmConv("insert", format, "", reflect.StructField{}, v))
		parts := strings.SplitN(format, "=", 2)
		param_ = append(param_, c.Sprintf(parts[1], len(args_)))
		fields_ = append(fields_, parts[0])
	}
	return
}

func AppendInsertf(fields, param []string, args []interface{}, formats string, v ...interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = Default.AppendInsertf(fields, param, args, formats, v...)
	return
}

func (c *CRUD) AppendInsertf(fields, param []string, args []interface{}, formats string, v ...interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = fields, param, args
	c.FilterFormatCall(formats, v, func(format string, arg interface{}) {
		args_ = append(args_, c.ParmConv("insert", format, "", reflect.StructField{}, arg))
		parts := strings.SplitN(format, "=", 2)
		param_ = append(param_, c.Sprintf(parts[1], len(args_)))
		fields_ = append(fields_, parts[0])
	})
	return
}

func AppendSet(sets []string, args []interface{}, ok bool, format string, v interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = Default.AppendSet(sets, args, ok, format, v)
	return
}

func (c *CRUD) AppendSet(sets []string, args []interface{}, ok bool, format string, v interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = sets, args
	if ok {
		args_ = append(args_, c.ParmConv("update", format, "", reflect.StructField{}, v))
		sets_ = append(sets_, c.Sprintf(format, len(args_)))
	}
	return
}

func AppendSetf(sets []string, args []interface{}, formats string, v ...interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = Default.AppendSetf(sets, args, formats, v...)
	return
}

func (c *CRUD) AppendSetf(sets []string, args []interface{}, formats string, v ...interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = sets, args
	c.FilterFormatCall(formats, v, func(format string, arg interface{}) {
		args_ = append(args_, c.ParmConv("update", format, "", reflect.StructField{}, arg))
		sets_ = append(sets_, c.Sprintf(format, len(args_)))
	})
	return
}

func AppendWhere(where []string, args []interface{}, ok bool, format string, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = Default.AppendWhere(where, args, ok, format, v)
	return
}

func (c *CRUD) AppendWhere(where []string, args []interface{}, ok bool, format string, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = where, args
	if ok {
		args_ = append(args_, c.ParmConv("where", format, "", reflect.StructField{}, v))
		where_ = append(where_, c.Sprintf(format, len(args_)))
	}
	return
}

func AppendWheref(where []string, args []interface{}, format string, v ...interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = Default.AppendWheref(where, args, format, v...)
	return
}

func (c *CRUD) AppendWheref(where []string, args []interface{}, formats string, v ...interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = where, args
	c.FilterFormatCall(formats, v, func(format string, arg interface{}) {
		args_ = append(args_, c.ParmConv("where", format, "", reflect.StructField{}, arg))
		where_ = append(where_, c.Sprintf(format, len(args_)))
	})
	return
}

func AppendWhereUnify(where []string, args []interface{}, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = Default.AppendWhereUnify(where, args, v)
	return
}

func (c *CRUD) AppendWhereUnify(where []string, args []interface{}, v interface{}) (where_ []string, args_ []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	modelValue := reflectValue.FieldByName("Where")
	where_, args_ = where, args
	filterWhere, filterArgs := c.FilterWhere(args, modelValue.Addr().Interface())
	where_ = append(where_, filterWhere...)
	args_ = append(args_, filterArgs...)
	return
}

func JoinWhere(sql string, where []string, sep string, suffix ...string) (sql_ string) {
	sql_ = Default.joinWhere(1, sql, where, sep, suffix...)
	return
}

func (c *CRUD) JoinWhere(sql string, where []string, sep string, suffix ...string) (sql_ string) {
	sql_ = c.joinWhere(1, sql, where, sep, suffix...)
	return
}

func (c *CRUD) joinWhere(caller int, sql string, where []string, sep string, suffix ...string) (sql_ string) {
	sql_ = sql
	if len(where) > 0 {
		sql_ += " where " + strings.Join(where, " "+sep+" ") + " " + strings.Join(suffix, " ")
	}
	if c.Verbose {
		c.Log(caller, "CRUD join where done with sql:%v", sql_)
	}
	return
}

func JoinWhereUnify(sql string, args []interface{}, v interface{}) (sql_ string, args_ []interface{}) {
	sql_, args_ = Default.joinWhereUnify(1, sql, args, v)
	return
}

func (c *CRUD) JoinWhereUnify(sql string, args []interface{}, v interface{}) (sql_ string, args_ []interface{}) {
	sql_, args_ = c.joinWhereUnify(1, sql, args, v)
	return
}

func (c *CRUD) joinWhereUnify(caller int, sql string, args []interface{}, v interface{}) (sql_ string, args_ []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	whereType, _ := reflectType.FieldByName("Where")
	whereJoin := whereType.Tag.Get("join")
	args_ = args
	where, args_ := c.AppendWhereUnify(nil, args_, v)
	sql_ = c.joinWhere(caller+1, sql, where, whereJoin)
	return
}

func JoinPage(sql, orderby string, offset, limit int) (sql_ string) {
	sql_ = Default.joinPage(1, sql, orderby, offset, limit)
	return
}

func (c *CRUD) JoinPage(sql, orderby string, offset, limit int) (sql_ string) {
	sql_ = c.joinPage(1, sql, orderby, offset, limit)
	return
}

func (c *CRUD) joinPage(caller int, sql, orderby string, offset, limit int) (sql_ string) {
	sql_ = sql
	if offset >= 0 || limit > 0 {
		sql_ += " " + orderby
	}
	if offset >= 0 {
		sql_ += fmt.Sprintf(" offset %v", offset)
	}
	if limit > 0 {
		sql_ += fmt.Sprintf(" limit %v", limit)
	}
	if c.Verbose {
		c.Log(caller, "CRUD join page done with sql:%v", sql_)
	}
	return
}

func JoinPageUnify(sql string, v interface{}) (sql_ string) {
	sql_ = Default.joinPageUnify(1, sql, v)
	return
}

func (c *CRUD) JoinPageUnify(sql string, v interface{}) (sql_ string) {
	sql_ = c.joinPageUnify(1, sql, v)
	return
}

func (c *CRUD) joinPageUnify(caller int, sql string, v interface{}) (sql_ string) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	pageValue := reflectValue.FieldByName("Page")
	pageType, _ := reflectType.FieldByName("Page")
	if !pageValue.IsValid() {
		return
	}
	order := ""
	orderType, _ := pageType.Type.FieldByName("Order")
	orderValue := pageValue.FieldByName("Order")
	if orderValue.IsValid() {
		order = orderValue.String()
		if len(order) < 1 {
			order = orderType.Tag.Get("default")
		}
	}
	offset := 0
	offsetValue := pageValue.FieldByName("Offset")
	if offsetValue.IsValid() {
		offset = int(offsetValue.Int())
	}
	limit := 0
	limitValue := pageValue.FieldByName("Limit")
	if limitValue.IsValid() {
		limit = int(limitValue.Int())
	}
	sql_ = c.joinPage(caller+1, sql, order, offset, limit)
	return
}

func (c *CRUD) queryerExec(queryer interface{}, sql string, args []interface{}) (affected int64, err error) {
	if q, ok := queryer.(Queryer); ok {
		affected, err = q.Exec(sql, args...)
	} else if q, ok := queryer.(CrudQueryer); ok {
		affected, err = q.CrudExec(sql, args...)
	} else {
		panic("queryer is not supported")
	}
	return
}

func (c *CRUD) queryerQuery(queryer interface{}, sql string, args []interface{}) (rows Rows, err error) {
	if q, ok := queryer.(Queryer); ok {
		rows, err = q.Query(sql, args...)
	} else if q, ok := queryer.(CrudQueryer); ok {
		rows, err = q.CrudQuery(sql, args...)
	} else {
		panic(fmt.Sprintf("queryer %v is not supported", reflect.TypeOf(queryer)))
	}
	return
}

func (c *CRUD) queryerQueryRow(queryer interface{}, sql string, args []interface{}) (row Row) {
	if q, ok := queryer.(Queryer); ok {
		row = q.QueryRow(sql, args...)
	} else if q, ok := queryer.(CrudQueryer); ok {
		row = q.CrudQueryRow(sql, args...)
	} else {
		panic(fmt.Sprintf("queryer %v is not supported", reflect.TypeOf(queryer)))
	}
	return
}

func InsertArgs(v interface{}, filter string) (table string, fields, param []string, args []interface{}) {
	table, fields, param, args = Default.insertArgs(1, v, filter)
	return
}

func (c *CRUD) InsertArgs(v interface{}, filter string) (table string, fields, param []string, args []interface{}) {
	table, fields, param, args = c.insertArgs(1, v, filter)
	return
}

func (c *CRUD) insertArgs(caller int, v interface{}, filter string) (table string, fields, param []string, args []interface{}) {
	table = c.FilterFieldCall("insert", v, filter, func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		args = append(args, c.ParmConv("insert", fieldName, fieldFunc, field, value))
		fields = append(fields, fieldName)
		param = append(param, fmt.Sprintf(c.ArgFormat, len(args)))
	})
	if c.Verbose {
		c.Log(caller, "CRUD generate insert args by struct:%v,filter:%v, result is fields:%v,param:%v,args:%v", reflect.TypeOf(v), filter, fields, param, jsonString(args))
	}
	return
}

func InsertSQL(v interface{}, filter string, suffix ...string) (sql string, args []interface{}) {
	sql, args = Default.insertSQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) InsertSQL(v interface{}, filter string, suffix ...string) (sql string, args []interface{}) {
	sql, args = c.insertSQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) insertSQL(caller int, v interface{}, filter string, suffix ...string) (sql string, args []interface{}) {
	table, fields, param, args := c.insertArgs(caller+1, v, filter)
	sql = fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, strings.Join(fields, ","), strings.Join(param, ","), strings.Join(suffix, " "))
	if c.Verbose {
		c.Log(caller, "CRUD generate insert sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func InsertFilter(queryer, v interface{}, filter, join, scan string) (affected int64, err error) {
	affected, err = Default.insertFilter(1, queryer, v, filter, join, scan)
	return
}

func (c *CRUD) InsertFilter(queryer, v interface{}, filter, join, scan string) (affected int64, err error) {
	affected, err = c.insertFilter(1, queryer, v, filter, join, scan)
	return
}

func (c *CRUD) insertFilter(caller int, queryer, v interface{}, filter, join, scan string) (affected int64, err error) {
	table, fields, param, args := c.insertArgs(caller+1, v, filter)
	sql := fmt.Sprintf(`insert into %v(%v) values(%v)`, table, strings.Join(fields, ","), strings.Join(param, ","))
	if len(scan) < 1 {
		affected, err = c.queryerExec(queryer, sql, args)
		return
	}
	_, scanFields := c.queryField(caller+1, v, scan)
	scanArgs := c.ScanArgs(v, scan)
	if len(join) > 0 {
		sql += " " + join
	}
	sql += " " + strings.Join(scanFields, ",")
	err = c.queryerQueryRow(queryer, sql, args).Scan(scanArgs...)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD insert filter by struct:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD insert filter by struct:%v,sql:%v, result is success", reflect.TypeOf(v), sql)
	}
	return
}

func UpdateArgs(v interface{}, filter string, args []interface{}) (table string, sets []string, args_ []interface{}) {
	table, sets, args_ = Default.updateArgs(1, v, filter, args)
	return
}

func (c *CRUD) UpdateArgs(v interface{}, filter string, args []interface{}) (table string, sets []string, args_ []interface{}) {
	table, sets, args_ = c.updateArgs(1, v, filter, args)
	return
}

func (c *CRUD) updateArgs(caller int, v interface{}, filter string, args []interface{}) (table string, sets []string, args_ []interface{}) {
	args_ = args
	table = c.FilterFieldCall("update", v, filter, func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		args_ = append(args_, c.ParmConv("update", fieldName, fieldFunc, field, value))
		sets = append(sets, fmt.Sprintf("%v="+c.ArgFormat, fieldName, len(args_)))
	})
	if c.Verbose {
		c.Log(caller, "CRUD generate update args by struct:%v,filter:%v, result is sets:%v,args:%v", reflect.TypeOf(v), filter, sets, jsonString(args_))
	}
	return
}

func UpdateSQL(v interface{}, filter string, args []interface{}, suffix ...string) (sql string, args_ []interface{}) {
	sql, args_ = Default.updateSQL(1, v, filter, args, suffix...)
	return
}

func (c *CRUD) UpdateSQL(v interface{}, filter string, args []interface{}, suffix ...string) (sql string, args_ []interface{}) {
	sql, args_ = c.updateSQL(1, v, filter, args, suffix...)
	return
}

func (c *CRUD) updateSQL(caller int, v interface{}, filter string, args []interface{}, suffix ...string) (sql string, args_ []interface{}) {
	table, sets, args_ := c.updateArgs(caller+1, v, filter, args)
	sql = fmt.Sprintf(`update %v set %v %v`, table, strings.Join(sets, ","), strings.Join(suffix, " "))
	if c.Verbose {
		c.Log(caller, "CRUD generate update sql by struct:%v,filter:%v, result is sql:%v,args:%v", reflect.TypeOf(v), filter, sql, jsonString(args_))
	}
	return
}

func Update(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = Default.update(1, queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) Update(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = c.update(1, queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) update(caller int, queryer, v interface{}, sets, where []string, sep string, args []interface{}) (affected int64, err error) {
	table := c.Table(v)
	sql := fmt.Sprintf(`update %v set %v`, table, strings.Join(sets, ","))
	sql = c.joinWhere(caller+1, sql, where, sep)
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD update by struct:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD update by struct:%v,sql:%v,args:%v, result is success affected:%v", reflect.TypeOf(v), sql, jsonString(args), affected)
	}
	return
}

func UpdateRow(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (err error) {
	err = Default.updateRow(1, queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) UpdateRow(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (err error) {
	err = c.updateRow(1, queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) updateRow(caller int, queryer, v interface{}, sets, where []string, sep string, args []interface{}) (err error) {
	affected, err := c.update(caller+1, queryer, v, sets, where, sep, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func UpdateFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = Default.updateFilter(1, queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) UpdateFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = c.updateFilter(1, queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) updateFilter(caller int, queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (affected int64, err error) {
	sql, args := c.updateSQL(caller+1, v, filter, args)
	sql = c.joinWhere(caller+1, sql, where, sep)
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD update filter by struct:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD update filter by struct:%v,sql:%v,args:%v, result is success affected:%v", reflect.TypeOf(v), sql, jsonString(args), affected)
	}
	return
}

func UpdateFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (err error) {
	err = Default.updateFilterRow(1, queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) UpdateFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (err error) {
	err = c.updateFilterRow(1, queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) updateFilterRow(caller int, queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (err error) {
	affected, err := c.updateFilter(caller+1, queryer, v, filter, where, sep, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func UpdateSimple(queryer, v interface{}, filter, suffix string, args []interface{}) (affected int64, err error) {
	affected, err = Default.updateSimple(1, queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) UpdateSimple(queryer, v interface{}, filter, suffix string, args []interface{}) (affected int64, err error) {
	affected, err = c.updateSimple(1, queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) updateSimple(caller int, queryer, v interface{}, filter, suffix string, args []interface{}) (affected int64, err error) {
	sql, args := c.updateSQL(caller+1, v, filter, args)
	if len(suffix) > 0 {
		sql += " " + suffix
	}
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD update simple by struct:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD update simple by struct:%v,sql:%v,args:%v, result is success affected:%v", reflect.TypeOf(v), sql, jsonString(args), affected)
	}
	return
}

func UpdateSimpleRow(queryer, v interface{}, filter, suffix string, args []interface{}) (err error) {
	err = Default.updateSimpleRow(1, queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) UpdateSimpleRow(queryer, v interface{}, filter, suffix string, args []interface{}) (err error) {
	err = c.updateSimpleRow(1, queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) updateSimpleRow(caller int, queryer, v interface{}, filter, suffix string, args []interface{}) (err error) {
	affected, err := c.updateSimple(caller+1, queryer, v, filter, suffix, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func QueryField(v interface{}, filter string) (table string, fields []string) {
	table, fields = Default.queryField(1, v, filter)
	return
}

func (c *CRUD) QueryField(v interface{}, filter string) (table string, fields []string) {
	table, fields = c.queryField(1, v, filter)
	return
}

func (c *CRUD) queryField(caller int, v interface{}, filter string) (table string, fields []string) {
	table = c.FilterFieldCall("query", v, filter, func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		conv := field.Tag.Get("conv")
		if len(fieldFunc) > 0 {
			fields = append(fields, fmt.Sprintf("%v(%v%v)", fieldFunc, fieldName, conv))
		} else {
			fields = append(fields, fmt.Sprintf("%v%v", fieldName, conv))
		}
	})
	if c.Verbose {
		c.Log(caller, "CRUD generate query field by struct:%v,filter:%v, result is fields:%v", reflect.TypeOf(v), filter, fields)
	}
	return
}

func QuerySQL(v interface{}, filter string, suffix ...string) (sql string) {
	sql = Default.querySQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) QuerySQL(v interface{}, filter string, suffix ...string) (sql string) {
	sql = c.querySQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) querySQL(caller int, v interface{}, filter string, suffix ...string) (sql string) {
	table, fields := c.queryField(caller+1, v, filter)
	sql = fmt.Sprintf(`select %v from %v %v`, strings.Join(fields, ","), table, strings.Join(suffix, " "))
	if c.Verbose {
		c.Log(caller, "CRUD generate query sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func QueryUnifySQL(v interface{}) (sql string, args []interface{}) {
	sql, args = Default.queryUnifySQL(1, v)
	return
}

func (c *CRUD) QueryUnifySQL(v interface{}) (sql string, args []interface{}) {
	sql, args = c.queryUnifySQL(1, v)
	return
}

func (c *CRUD) queryUnifySQL(caller int, v interface{}) (sql string, args []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	modelValue := reflectValue.FieldByName("Model")
	queryType, _ := reflectType.FieldByName("Query")
	queryFilter := queryType.Tag.Get("filter")
	sql = c.querySQL(caller+1, modelValue.Addr().Interface(), queryFilter)
	sql, args = c.joinWhereUnify(caller+1, sql, nil, v)
	sql = c.joinPageUnify(caller+1, sql, v)
	return
}

func ScanArgs(v interface{}, filter string) (args []interface{}) {
	args = Default.ScanArgs(v, filter)
	return
}

func (c *CRUD) ScanArgs(v interface{}, filter string) (args []interface{}) {
	c.FilterFieldCall("scan", v, filter, func(fieldName, fieldFunc string, field reflect.StructField, value interface{}) {
		args = append(args, c.ParmConv("scan", fieldName, fieldFunc, field, value))
	})
	return
}

func ScanUnifyDest(v interface{}, queryName string) (modelValue interface{}, queryFilter string, dests []interface{}) {
	modelValue, queryFilter, dests = Default.ScanUnifyDest(v, queryName)
	return
}

func (c *CRUD) ScanUnifyDest(v interface{}, queryName string) (modelValue interface{}, queryFilter string, dests []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	modelValue = reflectValue.FieldByName("Model").Addr().Interface()
	queryType, _ := reflectType.FieldByName(queryName)
	queryValue := reflectValue.FieldByName(queryName)
	if !queryValue.IsValid() {
		panic(fmt.Sprintf("%v is not exits in %v", queryName, reflectType))
	}
	queryFilter = queryType.Tag.Get("filter")
	queryNum := queryType.Type.NumField()
	for i := 0; i < queryNum; i++ {
		scan := queryType.Type.Field(i).Tag.Get("scan")
		if scan == "-" {
			continue
		}
		dests = append(dests, queryValue.Field(i).Addr().Interface())
		if len(scan) > 0 {
			dests = append(dests, scan)
		}
	}
	return
}

func (c *CRUD) destSet(value reflect.Value, filter string, dests ...interface{}) (err error) {
	valueField := func(key string) (v reflect.Value, e error) {
		if _, ok := value.Interface().([]interface{}); ok {
			filterFields := strings.Split(strings.TrimSpace(strings.SplitN(filter, "#", 2)[0]), ",")
			indexField := -1
			for i, filterField := range filterFields {
				fieldParts := strings.SplitN(strings.Trim(filterField, ")"), "(", 2)
				fieldName := fieldParts[0]
				if len(fieldParts) > 1 {
					fieldName = fieldParts[1]
				}
				if fieldName == key {
					indexField = i
					break
				}
			}
			if indexField < 0 {
				e = fmt.Errorf("field %v is not exists", key)
				return
			}
			v = reflect.Indirect(value.Index(indexField).Elem())
			return
		}
		targetValue := reflect.Indirect(value)
		targetType := targetValue.Type()
		k := targetValue.NumField()
		for i := 0; i < k; i++ {
			field := targetType.Field(i)
			fieldName := strings.SplitN(field.Tag.Get(c.Tag), ",", 2)[0]
			if fieldName == key {
				v = targetValue.Field(i)
				return
			}
		}
		e = fmt.Errorf("field %v is not exists", key)
		return
	}
	n := len(dests)
	for i := 0; i < n; i++ {
		if scanner, ok := dests[i].(Scanner); ok {
			scanner.Scan(value.Interface())
			continue
		}
		destValue := reflect.Indirect(reflect.ValueOf(dests[i]))
		destKind := destValue.Kind()
		destType := destValue.Type()
		if destType == value.Type() {
			destValue.Set(value)
			continue
		}
		if value.Kind() == reflect.Ptr {
			indirectValue := reflect.Indirect(value)
			if destType == indirectValue.Type() {
				destValue.Set(indirectValue)
				continue
			}
		}
		switch destKind {
		case reflect.Func:
			destValue.Call([]reflect.Value{value})
		case reflect.Map:
			if i+1 >= len(dests) {
				err = fmt.Errorf("dest[%v] pattern is not setted", i)
				break
			}
			v, ok := dests[i+1].(string)
			if !ok {
				err = fmt.Errorf("dest[%v] pattern is not string", i)
				break
			}
			if len(v) < 1 {
				err = fmt.Errorf("dest[%v] pattern is empty", i)
				break
			}
			i++
			parts := strings.Split(v, "#")
			kvs := strings.SplitN(parts[0], ":", 2)
			targetKey, xerr := valueField(kvs[0])
			if xerr != nil {
				err = xerr
				break
			}
			targetValue := value
			if len(kvs) > 1 {
				targetValue, xerr = valueField(kvs[1])
				if xerr != nil {
					err = xerr
					break
				}
			}
			if destValue.IsNil() {
				destValue.Set(reflect.MakeMap(destType))
			}
			destElemValue := destValue.MapIndex(targetKey)
			destElemType := destType.Elem()
			if destElemType.Kind() == reflect.Slice {
				if !destElemValue.IsValid() {
					destElemValue = reflect.Indirect(reflect.New(destElemType))
				}
				destValue.SetMapIndex(targetKey, reflect.Append(destElemValue, targetValue))
			} else {
				destValue.SetMapIndex(targetKey, targetValue)
			}
		default:
			if destKind == reflect.Slice && destType.Elem() == value.Type() {
				destValue.Set(reflect.Append(destValue, value))
				continue
			}
			if i+1 >= len(dests) {
				err = fmt.Errorf("dest[%v] pattern is not setted", i)
				break
			}
			v, ok := dests[i+1].(string)
			if !ok {
				err = fmt.Errorf("dest[%v] pattern is not string", i)
				break
			}
			if len(v) < 1 {
				err = fmt.Errorf("dest[%v] pattern is empty", i)
				break
			}
			parts := strings.Split(v, "#")
			skipNil, skipZero := true, true
			if len(parts) > 1 && parts[1] == "all" {
				skipNil, skipZero = false, false
			}
			i++
			targetValue, xerr := valueField(parts[0])
			if xerr != nil {
				err = xerr
				break
			}
			checkValue := targetValue
			targetKind := targetValue.Kind()
			if targetKind == reflect.Ptr && checkValue.IsNil() && skipNil {
				continue
			}
			if targetKind == reflect.Ptr && !checkValue.IsNil() {
				checkValue = reflect.Indirect(checkValue)
			}
			if checkValue.IsZero() && skipZero {
				continue
			}
			if destKind == reflect.Slice && destType.Elem() == targetValue.Type() {
				destValue.Set(reflect.Append(reflect.Indirect(destValue), targetValue))
			} else if destType == targetValue.Type() {
				destValue.Set(targetValue)
			} else {
				err = fmt.Errorf("not supported on dests[%v] to set %v=>%v", i-1, targetValue.Type(), destType)
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func Scan(rows Rows, v interface{}, filter string, dest ...interface{}) (err error) {
	err = Default.Scan(rows, v, filter, dest...)
	return
}

func (c *CRUD) Scan(rows Rows, v interface{}, filter string, dest ...interface{}) (err error) {
	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	for rows.Next() {
		value := NewValue(v)
		err = rows.Scan(c.ScanArgs(value.Interface(), filter)...)
		if err != nil {
			break
		}
		if !isPtr {
			value = reflect.Indirect(value)
		}
		err = c.destSet(value, filter, dest...)
		if err != nil {
			break
		}
	}
	return
}

func ScanUnify(rows Rows, v interface{}) (err error) {
	err = Default.ScanUnify(rows, v)
	return
}

func (c *CRUD) ScanUnify(rows Rows, v interface{}) (err error) {
	modelValue, modelFilter, dests := c.ScanUnifyDest(v, "Query")
	err = c.Scan(rows, modelValue, modelFilter, dests...)
	return
}

func Query(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.query(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) Query(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.query(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) query(caller int, queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	rows, err := c.queryerQuery(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD query by struct:%v,filter:%v,sql:%v,args:%v result is fail:%v", reflect.TypeOf(v), filter, sql, jsonString(args), err)
		}
		return
	}
	defer rows.Close()
	if c.Verbose {
		c.Log(caller, "CRUD query by struct:%v,filter:%v,sql:%v,args:%v result is success", reflect.TypeOf(v), filter, sql, jsonString(args))
	}
	err = c.Scan(rows, v, filter, dest...)
	return
}

func QueryFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, orderby string, offset, limit int, dest ...interface{}) (err error) {
	err = Default.queryFilter(1, queryer, v, filter, where, sep, args, orderby, offset, limit, dest...)
	return
}

func (c *CRUD) QueryFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, orderby string, offset, limit int, dest ...interface{}) (err error) {
	err = c.queryFilter(1, queryer, v, filter, where, sep, args, orderby, offset, limit, dest...)
	return
}

func (c *CRUD) queryFilter(caller int, queryer, v interface{}, filter string, where []string, sep string, args []interface{}, orderby string, offset, limit int, dest ...interface{}) (err error) {
	sql := c.querySQL(caller+1, v, filter)
	sql = c.joinWhere(caller+1, sql, where, sep)
	sql = c.joinPage(caller+1, sql, orderby, offset, limit)
	err = c.query(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func QuerySimple(queryer, v interface{}, filter string, suffix string, args []interface{}, offset, limit int, dest ...interface{}) (err error) {
	err = Default.querySimple(1, queryer, v, filter, suffix, args, offset, limit, dest...)
	return
}

func (c *CRUD) QuerySimple(queryer, v interface{}, filter string, suffix string, args []interface{}, offset, limit int, dest ...interface{}) (err error) {
	err = c.querySimple(1, queryer, v, filter, suffix, args, offset, limit, dest...)
	return
}

func (c *CRUD) querySimple(caller int, queryer, v interface{}, filter string, suffix string, args []interface{}, offset, limit int, dest ...interface{}) (err error) {
	sql := c.querySQL(caller+1, v, filter) + " " + suffix
	sql = c.joinPage(caller+1, sql, "", offset, limit)
	err = c.query(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func QueryUnify(queryer, v interface{}) (err error) {
	err = Default.queryUnify(1, queryer, v)
	return
}

func (c *CRUD) QueryUnify(queryer, v interface{}) (err error) {
	err = c.queryUnify(1, queryer, v)
	return
}

func (c *CRUD) queryUnify(caller int, queryer, v interface{}) (err error) {
	sql, args := c.queryUnifySQL(caller+1, v)
	rows, err := c.queryerQuery(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD query unify by struct:%v,sql:%v,args:%v result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	defer rows.Close()
	if c.Verbose {
		c.Log(caller, "CRUD query unify by struct:%v,sql:%v,args:%v result is success", reflect.TypeOf(v), sql, jsonString(args))
	}
	err = c.ScanUnify(rows, v)
	return
}

func ScanRow(row Row, v interface{}, filter string, dest ...interface{}) (err error) {
	err = Default.ScanRow(row, v, filter, dest...)
	return
}

func (c *CRUD) ScanRow(row Row, v interface{}, filter string, dest ...interface{}) (err error) {
	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	value := NewValue(v)
	err = row.Scan(c.ScanArgs(value.Interface(), filter)...)
	if err != nil {
		return
	}
	if !isPtr {
		value = reflect.Indirect(value)
	}
	err = c.destSet(value, filter, dest...)
	if err != nil {
		return
	}
	return
}

func ScanUnifyRow(row Row, v interface{}) (err error) {
	err = Default.ScanUnifyRow(row, v)
	return
}

func (c *CRUD) ScanUnifyRow(row Row, v interface{}) (err error) {
	modelValue, modelFilter, dests := c.ScanUnifyDest(v, "QueryRow")
	err = c.ScanRow(row, modelValue, modelFilter, dests...)
	return
}

func QueryRow(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.queryRow(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) QueryRow(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.queryRow(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) queryRow(caller int, queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.ScanRow(c.queryerQueryRow(queryer, sql, args), v, filter, dest...)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD query by struct:%v,filter:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), filter, sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD query by struct:%v,filter:%v,sql:%v,args:%v, result is success", reflect.TypeOf(v), filter, sql, jsonString(args))
	}
	return
}

func QueryFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.queryFilterRow(1, queryer, v, filter, where, sep, args, dest...)
	return
}

func (c *CRUD) QueryFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	err = c.queryFilterRow(1, queryer, v, filter, where, sep, args, dest...)
	return
}

func (c *CRUD) queryFilterRow(caller int, queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.querySQL(caller+1, v, filter)
	sql = c.joinWhere(caller+1, sql, where, sep)
	err = c.queryRow(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func QuerySimpleRow(queryer, v interface{}, filter string, suffix string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.querySimpleRow(1, queryer, v, filter, suffix, args, dest...)
	return
}

func (c *CRUD) QuerySimpleRow(queryer, v interface{}, filter string, suffix string, args []interface{}, dest ...interface{}) (err error) {
	err = c.querySimpleRow(1, queryer, v, filter, suffix, args, dest...)
	return
}

func (c *CRUD) querySimpleRow(caller int, queryer, v interface{}, filter string, suffix string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.querySQL(caller+1, v, filter) + " " + suffix
	err = c.queryRow(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func QueryUnifyRow(queryer, v interface{}) (err error) {
	err = Default.queryUnifyRow(1, queryer, v)
	return
}

func (c *CRUD) QueryUnifyRow(queryer, v interface{}) (err error) {
	err = c.queryUnifyRow(1, queryer, v)
	return
}

func (c *CRUD) queryUnifyRow(caller int, queryer, v interface{}) (err error) {
	sql, args := c.queryUnifySQL(caller+1, v)
	err = c.ScanUnifyRow(c.queryerQueryRow(queryer, sql, args), v)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD query unify row by struct:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD query unify row by struct:%v,sql:%v,args:%v, result is success", reflect.TypeOf(v), sql, jsonString(args))
	}
	return
}

func CountSQL(v interface{}, filter string, suffix ...string) (sql string) {
	sql = Default.countSQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) CountSQL(v interface{}, filter string, suffix ...string) (sql string) {
	sql = c.countSQL(1, v, filter, suffix...)
	return
}

func (c *CRUD) countSQL(caller int, v interface{}, filter string, suffix ...string) (sql string) {
	var table string
	var fields []string
	if len(filter) < 1 || filter == "*" || filter == "count(*)" || filter == "count(*)#all" {
		table = c.Table(v)
		fields = []string{"count(*)"}
	} else {
		table, fields = c.queryField(caller+1, v, filter)
	}
	sql = fmt.Sprintf(`select %v from %v %v`, strings.Join(fields, ","), table, strings.Join(suffix, " "))
	if c.Verbose {
		c.Log(caller, "CRUD generate count sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func CountUnifySQL(v interface{}) (sql string, args []interface{}) {
	sql, args = Default.countUnifySQL(1, v)
	return
}

func (c *CRUD) CountUnifySQL(v interface{}) (sql string, args []interface{}) {
	sql, args = c.countUnifySQL(1, v)
	return
}

func (c *CRUD) countUnifySQL(caller int, v interface{}) (sql string, args []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	modelValue := reflectValue.FieldByName("Model").Addr().Interface()
	queryType, _ := reflectType.FieldByName("Count")
	queryFilter := queryType.Tag.Get("filter")
	sql = c.countSQL(caller+1, modelValue, queryFilter)
	sql, args = c.joinWhereUnify(caller+1, sql, nil, v)
	return
}

func CountUnifyDest(v interface{}) (modelValue interface{}, queryFilter string, dests []interface{}) {
	modelValue, queryFilter, dests = Default.CountUnifyDest(v)
	return
}

func (c *CRUD) CountUnifyDest(v interface{}) (modelValue interface{}, queryFilter string, dests []interface{}) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	modelValueList := []interface{}{
		TableName(c.Table(reflectValue.FieldByName("Model").Addr().Interface())),
	}
	queryType, _ := reflectType.FieldByName("Count")
	queryValue := reflectValue.FieldByName("Count")
	queryFilter = queryType.Tag.Get("filter")
	queryNum := queryType.Type.NumField()
	for i := 0; i < queryNum; i++ {
		scan := queryType.Type.Field(i).Tag.Get("scan")
		if scan == "-" {
			continue
		}
		modelValueList = append(modelValueList, queryValue.Field(i).Interface())
		dests = append(dests, queryValue.Field(i).Addr().Interface())
		if len(scan) > 0 {
			dests = append(dests, scan)
		}
	}
	modelValue = modelValueList
	return
}

func Count(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.count(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) Count(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.count(1, queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) count(caller int, queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.ScanRow(c.queryerQueryRow(queryer, sql, args), v, filter, dest...)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD count by struct:%v,filter:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), filter, sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD count by struct:%v,filter:%v,sql:%v,args:%v, result is success", reflect.TypeOf(v), filter, sql, jsonString(args))
	}
	return
}

func CountFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.countFilter(1, queryer, v, filter, where, sep, args, dest...)
	return
}

func (c *CRUD) CountFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	err = c.countFilter(1, queryer, v, filter, where, sep, args, dest...)
	return
}

func (c *CRUD) countFilter(caller int, queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.countSQL(caller+1, v, filter)
	sql = c.joinWhere(caller+1, sql, where, sep)
	err = c.count(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func CountSimple(queryer, v interface{}, filter, suffix string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.countSimple(1, queryer, v, filter, suffix, args, dest...)
	return
}

func (c *CRUD) CountSimple(queryer, v interface{}, filter, suffix string, args []interface{}, dest ...interface{}) (err error) {
	err = c.countSimple(1, queryer, v, filter, suffix, args, dest...)
	return
}

func (c *CRUD) countSimple(caller int, queryer, v interface{}, filter, suffix string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.countSQL(caller+1, v, filter, suffix)
	err = c.count(caller+1, queryer, v, filter, sql, args, dest...)
	return
}

func CountUnify(queryer, v interface{}) (err error) {
	err = Default.countUnify(1, queryer, v)
	return
}

func (c *CRUD) CountUnify(queryer, v interface{}) (err error) {
	err = c.countUnify(1, queryer, v)
	return
}

func (c *CRUD) countUnify(caller int, queryer, v interface{}) (err error) {
	sql, args := c.countUnifySQL(caller+1, v)
	modelValue, queryFilter, dests := c.CountUnifyDest(v)
	err = c.ScanRow(c.queryerQueryRow(queryer, sql, args), modelValue, queryFilter, dests...)
	if err != nil {
		if c.Verbose {
			c.Log(caller, "CRUD count unify by struct:%v,sql:%v,args:%v, result is fail:%v", reflect.TypeOf(v), sql, jsonString(args), err)
		}
		return
	}
	if c.Verbose {
		c.Log(caller, "CRUD count unify by struct:%v,sql:%v,args:%v, result is success", reflect.TypeOf(v), sql, jsonString(args))
	}
	return
}

func ApplyUnify(queryer, v interface{}) (err error) {
	err = Default.applyUnify(1, queryer, v)
	return
}

func (c *CRUD) ApplyUnify(queryer, v interface{}) (err error) {
	err = c.applyUnify(1, queryer, v)
	return
}

func (c *CRUD) applyUnify(caller int, queryer, v interface{}) (err error) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	if value := reflectValue.FieldByName("Query"); value.IsValid() && err == nil {
		enabled := value.FieldByName("Enabled")
		if enabled.IsValid() || (enabled.IsValid() && enabled.Bool()) {
			err = c.queryUnify(caller+1, queryer, v)
		}
	}
	if value := reflectValue.FieldByName("QueryRow"); value.IsValid() && err == nil {
		enabled := value.FieldByName("Enabled")
		if enabled.IsValid() || (enabled.IsValid() && enabled.Bool()) {
			err = c.queryUnifyRow(caller+1, queryer, v)
		}
	}
	if value := reflectValue.FieldByName("Count"); value.IsValid() && err == nil {
		enabled := value.FieldByName("Enabled")
		if enabled.IsValid() || (enabled.IsValid() && enabled.Bool()) {
			err = c.countUnify(caller+1, queryer, v)
		}
	}
	return
}
