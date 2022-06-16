//Package crud is simple curd tools to process database
package crud

import (
	"fmt"
	"reflect"
	"strings"
)

var Default = &CRUD{
	Tag:       "json",
	ArgFormat: "$%v",
	ErrNoRows: fmt.Errorf("no rows"),
	Log:       func(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) },
}

type NameConv func(on, name string, field reflect.StructField) string
type LogF func(format string, args ...interface{})

type CRUD struct {
	Tag       string
	ArgFormat string
	ErrNoRows error
	Verbose   bool
	Log       LogF
	NameConv  NameConv
}

func Table(v interface{}) (table string) {
	table = Default.Table(v)
	return
}

func (c *CRUD) Table(v interface{}) (table string) {
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

func FilterFieldCall(on, filter string, v interface{}, call func(name string, field reflect.StructField, value interface{})) (table string) {
	table = Default.FilterFieldCall(on, filter, v, call)
	return
}

func (c *CRUD) FilterFieldCall(on, filter string, v interface{}, call func(name string, field reflect.StructField, value interface{})) (table string) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	if reflectType.Kind() != reflect.Struct {
		call(filter, reflect.StructField{}, v)
		return
	}
	var inc, exc string
	var incNil, incZero bool
	if len(filter) > 0 {
		filter = strings.TrimSpace(filter)
		parts := strings.SplitN(filter, "#", 2)
		if len(parts[0]) > 0 {
			if strings.HasPrefix(parts[0], "^") {
				exc = "," + strings.TrimPrefix(parts[0], "^") + ","
			} else {
				inc = "," + parts[0] + ","
			}
		}
		if len(parts) > 1 && len(parts[1]) > 0 {
			incNil = strings.Contains(","+parts[1]+",", ",nil,") || strings.Contains(","+parts[1]+",", ",all,")
			incZero = strings.Contains(","+parts[1]+",", ",zero,") || strings.Contains(","+parts[1]+",", ",all,")
		}
	}
	numField := reflectType.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := reflectValue.Field(i)
		fieldType := reflectType.Field(i)
		if fieldType.Name == "_" {
			if t := fieldType.Tag.Get("table"); len(t) > 0 {
				table = t
			}
		}
		fieldName := strings.SplitN(fieldType.Tag.Get(c.Tag), ",", 2)[0]
		fieldKind := fieldValue.Kind()
		checkValue := fieldValue
		if len(fieldName) < 1 || fieldName == "-" {
			continue
		}
		if len(inc) > 0 && !strings.Contains(inc, ","+fieldName+",") {
			continue
		}
		if len(exc) > 0 && strings.Contains(exc, ","+fieldName+",") {
			continue
		}
		if fieldKind == reflect.Ptr && checkValue.IsNil() && !incNil {
			continue
		}
		if fieldKind == reflect.Ptr && !checkValue.IsNil() {
			checkValue = reflect.Indirect(checkValue)
		}
		if checkValue.IsZero() && !incZero {
			continue
		}
		if c.NameConv != nil {
			fieldName = c.NameConv(on, fieldName, fieldType)
		}
		call(fieldName, fieldType, fieldValue.Addr().Interface())
	}
	return
}

func AppendInsert(fields, param []string, args []interface{}, ok bool, formats ...interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = Default.AppendInsert(fields, param, args, ok, formats...)
	return
}

func (c *CRUD) AppendInsert(fields, param []string, args []interface{}, ok bool, formats ...interface{}) (fields_, param_ []string, args_ []interface{}) {
	fields_, param_, args_ = fields, param, args
	if ok {
		n := len(formats) / 2
		for i := 0; i < n; i++ {
			args_ = append(args_, formats[i*2+1])
			parts := strings.SplitN(formats[i*2].(string), "=", 2)
			param_ = append(param_, fmt.Sprintf(parts[1], len(args_)))
			fields_ = append(fields_, parts[0])
		}
	}
	return
}

func AppendSet(sets []string, args []interface{}, ok bool, formats ...interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = Default.AppendSet(sets, args, ok, formats...)
	return
}

func (c *CRUD) AppendSet(sets []string, args []interface{}, ok bool, formats ...interface{}) (sets_ []string, args_ []interface{}) {
	sets_, args_ = sets, args
	if ok {
		n := len(formats) / 2
		for i := 0; i < n; i++ {
			args_ = append(args_, formats[i*2+1])
			sets_ = append(sets_, fmt.Sprintf(formats[i*2].(string), len(args_)))
		}
	}
	return
}

func AppendWhere(where []string, args []interface{}, ok bool, formats ...interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = Default.AppendWhere(where, args, ok, formats...)
	return
}

func (c *CRUD) AppendWhere(where []string, args []interface{}, ok bool, formats ...interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = where, args
	if ok {
		n := len(formats) / 2
		for i := 0; i < n; i++ {
			args_ = append(args_, formats[i*2+1])
			where_ = append(where_, fmt.Sprintf(formats[i*2].(string), len(args_)))
		}
	}
	return
}

func AppendWhereN(where []string, args []interface{}, ok bool, format string, n int, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = Default.AppendWhereN(where, args, ok, format, n, v)
	return
}

func (c *CRUD) AppendWhereN(where []string, args []interface{}, ok bool, format string, n int, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = where, args
	if ok {
		args_ = append(args_, v)
		fargs := []interface{}{}
		for i := 0; i < n; i++ {
			fargs = append(fargs, fmt.Sprintf("%v", len(args_)))
		}
		where_ = append(where_, fmt.Sprintf(format, fargs...))
	}
	return
}

func JoinWhere(sql string, where []string, sep string, suffix ...string) (sql_ string) {
	sql_ = Default.JoinWhere(sql, where, sep, suffix...)
	return
}

func (c *CRUD) JoinWhere(sql string, where []string, sep string, suffix ...string) (sql_ string) {
	sql_ = sql
	if len(where) > 0 {
		sql_ += " where " + strings.Join(where, " "+sep+" ") + " " + strings.Join(suffix, " ")
	}
	if c.Verbose {
		c.Log("CRUD join where done with sql:%v", sql_)
	}
	return
}

func JoinPage(sql, orderby string, offset, limit int) (sql_ string) {
	sql_ = Default.JoinPage(sql, orderby, offset, limit)
	return
}

func (c *CRUD) JoinPage(sql, orderby string, offset, limit int) (sql_ string) {
	sql_ = sql
	if offset >= 0 || limit > 0 {
		sql_ += " " + orderby + " "
	}
	if offset >= 0 {
		sql_ += fmt.Sprintf(" offset %v ", offset)
	}
	if limit > 0 {
		sql_ += fmt.Sprintf(" limit %v ", limit)
	}
	if c.Verbose {
		c.Log("CRUD join page doen with sql:%v", sql_)
	}
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
	table, fields, param, args = Default.InsertArgs(v, filter)
	return
}

func (c *CRUD) InsertArgs(v interface{}, filter string) (table string, fields, param []string, args []interface{}) {
	table = c.FilterFieldCall("insert", filter, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
		fields = append(fields, name)
		param = append(param, fmt.Sprintf(c.ArgFormat, len(args)))
	})
	if c.Verbose {
		c.Log("CRUD generate insert args by struct:%v,filter:%v, result is fields:%v,param:%v", reflect.TypeOf(v), filter, fields, param)
	}
	return
}

func InsertSQL(v interface{}, filter string, suffix ...string) (sql string, args []interface{}) {
	sql, args = Default.InsertSQL(v, filter, suffix...)
	return
}

func (c *CRUD) InsertSQL(v interface{}, filter string, suffix ...string) (sql string, args []interface{}) {
	table, fields, param, args := c.InsertArgs(v, filter)
	sql = fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, strings.Join(fields, ","), strings.Join(param, ","), strings.Join(suffix, " "))
	if c.Verbose {
		c.Log("CRUD generate insert sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func InsertFilter(queryer, v interface{}, filter, join, scan string) (affected int64, err error) {
	affected, err = Default.InsertFilter(queryer, v, filter, join, scan)
	return
}

func (c *CRUD) InsertFilter(queryer, v interface{}, filter, join, scan string) (affected int64, err error) {
	table, fields, param, args := c.InsertArgs(v, filter)
	sql := fmt.Sprintf(`insert into %v(%v) values(%v)`, table, strings.Join(fields, ","), strings.Join(param, ","))
	if len(scan) < 1 {
		affected, err = c.queryerExec(queryer, sql, args)
		return
	}
	_, scanFields := c.QueryField(v, scan)
	scanArgs := c.ScanArgs(v, scan)
	if len(join) > 0 {
		sql += " " + join
	}
	sql += " " + strings.Join(scanFields, ",")
	err = c.queryerQueryRow(queryer, sql, args).Scan(scanArgs...)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD insert filter by struct:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log("CRUD insert filter by struct:%v,sql:%v, result is success", reflect.TypeOf(v), sql)
	}
	return
}

func UpdateArgs(v interface{}, filter string, args []interface{}) (table string, sets []string, args_ []interface{}) {
	table, sets, args_ = Default.UpdateArgs(v, filter, args)
	return
}

func (c *CRUD) UpdateArgs(v interface{}, filter string, args []interface{}) (table string, sets []string, args_ []interface{}) {
	args_ = args
	table = c.FilterFieldCall("update", filter, v, func(name string, field reflect.StructField, value interface{}) {
		args_ = append(args_, value)
		sets = append(sets, fmt.Sprintf("%v="+c.ArgFormat, name, len(args_)))
	})
	if c.Verbose {
		c.Log("CRUD generate update args by struct:%v,filter:%v, result is sets:%v", reflect.TypeOf(v), filter, sets)
	}
	return
}

func UpdateSQL(v interface{}, filter string, args []interface{}, suffix ...string) (sql string, args_ []interface{}) {
	sql, args_ = Default.UpdateSQL(v, filter, args, suffix...)
	return
}

func (c *CRUD) UpdateSQL(v interface{}, filter string, args []interface{}, suffix ...string) (sql string, args_ []interface{}) {
	table, sets, args_ := c.UpdateArgs(v, filter, args)
	sql = fmt.Sprintf(`update %v set %v %v`, table, strings.Join(sets, ","), strings.Join(suffix, " "))
	if c.Verbose {
		c.Log("CRUD generate update sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func Update(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = Default.Update(queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) Update(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (affected int64, err error) {
	table := c.Table(v)
	sql := fmt.Sprintf(`update %v set %v`, table, strings.Join(sets, ","))
	sql = c.JoinWhere(sql, where, sep)
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD update by struct:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log("CRUD update by struct:%v,sql:%v, result is success affected:%v", reflect.TypeOf(v), sql, affected)
	}
	return
}

func UpdateRow(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (err error) {
	err = Default.UpdateRow(queryer, v, sets, where, sep, args)
	return
}

func (c *CRUD) UpdateRow(queryer, v interface{}, sets, where []string, sep string, args []interface{}) (err error) {
	affected, err := c.Update(queryer, v, sets, where, sep, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func UpdateFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (affected int64, err error) {
	affected, err = Default.UpdateFilter(queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) UpdateFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (affected int64, err error) {
	sql, args := c.UpdateSQL(v, filter, args)
	sql = c.JoinWhere(sql, where, sep)
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD update filter by struct:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log("CRUD update filter by struct:%v,sql:%v, result is success affected:%v", reflect.TypeOf(v), sql, affected)
	}
	return
}

func UpdateFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (err error) {
	err = Default.UpdateFilterRow(queryer, v, filter, where, sep, args)
	return
}

func (c *CRUD) UpdateFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}) (err error) {
	affected, err := c.UpdateFilter(queryer, v, filter, where, sep, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func UpdateSimple(queryer, v interface{}, filter, suffix string, args []interface{}) (affected int64, err error) {
	affected, err = Default.UpdateSimple(queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) UpdateSimple(queryer, v interface{}, filter, suffix string, args []interface{}) (affected int64, err error) {
	sql, args := c.UpdateSQL(v, filter, args)
	if len(suffix) > 0 {
		sql += " " + suffix
	}
	affected, err = c.queryerExec(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD update simple by struct:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log("CRUD update simple by struct:%v,sql:%v, result is success affected:%v", reflect.TypeOf(v), sql, affected)
	}
	return
}

func UpdateSimpleRow(queryer, v interface{}, filter, suffix string, args []interface{}) (err error) {
	err = Default.UpdateSimpleRow(queryer, v, filter, suffix, args)
	return
}

func (c *CRUD) UpdateSimpleRow(queryer, v interface{}, filter, suffix string, args []interface{}) (err error) {
	affected, err := c.UpdateSimple(queryer, v, filter, suffix, args)
	if err == nil && affected < 1 {
		err = c.ErrNoRows
	}
	return
}

func QueryField(v interface{}, filter string) (table string, fields []string) {
	table, fields = Default.QueryField(v, filter)
	return
}

func (c *CRUD) QueryField(v interface{}, filter string) (table string, fields []string) {
	table = c.FilterFieldCall("query", filter, v, func(name string, field reflect.StructField, value interface{}) {
		conv := field.Tag.Get("conv")
		fields = append(fields, fmt.Sprintf("%v%v", name, conv))
	})
	if c.Verbose {
		c.Log("CRUD generate query field by struct:%v,filter:%v, result is fields:%v", reflect.TypeOf(v), filter, fields)
	}
	return
}

func QuerySQL(v interface{}, filter string, suffix ...string) (sql string) {
	sql = Default.QuerySQL(v, filter, suffix...)
	return
}

func (c *CRUD) QuerySQL(v interface{}, filter string, suffix ...string) (sql string) {
	table, fields := c.QueryField(v, filter)
	sql = fmt.Sprintf(`select %v from %v %v`, strings.Join(fields, ","), table, strings.Join(suffix, " "))
	if c.Verbose {
		c.Log("CRUD generate query sql by struct:%v,filter:%v, result is sql:%v", reflect.TypeOf(v), filter, sql)
	}
	return
}

func ScanArgs(v interface{}, filter string) (args []interface{}) {
	args = Default.ScanArgs(v, filter)
	return
}

func (c *CRUD) ScanArgs(v interface{}, filter string) (args []interface{}) {
	c.FilterFieldCall("scan", filter, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
	})
	return
}

func (c *CRUD) destSet(value reflect.Value, dests ...interface{}) (err error) {
	valueField := func(key string) (v reflect.Value, e error) {
		targetValue := reflect.Indirect(value)
		targetType := targetValue.Type()
		k := targetValue.NumField()
		for i := 0; i < k; i++ {
			field := targetType.Field(i)
			fieldName := field.Tag.Get(c.Tag)
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
			parts := strings.Split(v, ",")
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
		case reflect.Slice:
			if destType.Elem() == value.Type() {
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
			parts := strings.Split(v, ",")
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
			destValue.Set(reflect.Append(reflect.Indirect(destValue), targetValue))
		default:
			err = fmt.Errorf("type %v is not supported", destKind)
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
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	for rows.Next() {
		value := reflect.New(reflectType)
		err = rows.Scan(c.ScanArgs(value.Interface(), filter)...)
		if err != nil {
			break
		}
		if !isPtr {
			value = reflect.Indirect(value)
		}
		err = c.destSet(value, dest...)
		if err != nil {
			break
		}
	}
	return
}

func Query(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.Query(queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) Query(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	rows, err := c.queryerQuery(queryer, sql, args)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD query by struct:%v,filter:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), filter, sql, err)
		}
		return
	}
	defer rows.Close()
	if c.Verbose {
		c.Log("CRUD query by struct:%v,filter:%v,sql:%v, result is success", reflect.TypeOf(v), filter, sql)
	}
	err = c.Scan(rows, v, filter, dest...)
	return
}

func QueryFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, orderby string, offset, limit int, dest ...interface{}) (err error) {
	err = Default.QueryFilter(queryer, v, filter, where, sep, args, orderby, offset, limit, dest...)
	return
}

func (c *CRUD) QueryFilter(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, orderby string, offset, limit int, dest ...interface{}) (err error) {
	sql := c.QuerySQL(v, filter)
	sql = c.JoinWhere(sql, where, sep)
	sql = c.JoinPage(sql, orderby, offset, limit)
	err = c.Query(queryer, v, filter, sql, args, dest...)
	return
}

func QuerySimple(queryer, v interface{}, filter string, suffix string, args []interface{}, offset, limit int, dest ...interface{}) (err error) {
	err = Default.QuerySimple(queryer, v, filter, suffix, args, offset, limit, dest...)
	return
}

func (c *CRUD) QuerySimple(queryer, v interface{}, filter string, suffix string, args []interface{}, offset, limit int, dest ...interface{}) (err error) {
	sql := c.QuerySQL(v, filter) + " " + suffix
	sql = c.JoinPage(sql, "", offset, limit)
	err = c.Query(queryer, v, filter, sql, args, dest...)
	return
}

func ScanRow(row Row, v interface{}, filter string, dest ...interface{}) (err error) {
	err = Default.ScanRow(row, v, filter, dest...)
	return
}

func (c *CRUD) ScanRow(row Row, v interface{}, filter string, dest ...interface{}) (err error) {
	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	value := reflect.New(reflectType)
	err = row.Scan(c.ScanArgs(value.Interface(), filter)...)
	if err != nil {
		return
	}
	if !isPtr {
		value = reflect.Indirect(value)
	}
	err = c.destSet(value, dest...)
	if err != nil {
		return
	}
	return
}

func QueryRow(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.QueryRow(queryer, v, filter, sql, args, dest...)
	return
}

func (c *CRUD) QueryRow(queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = c.ScanRow(c.queryerQueryRow(queryer, sql, args), v, filter, dest...)
	if err != nil {
		if c.Verbose {
			c.Log("CRUD query by struct:%v,filter:%v,sql:%v, result is fail:%v", reflect.TypeOf(v), filter, sql, err)
		}
		return
	}
	if c.Verbose {
		c.Log("CRUD query by struct:%v,filter:%v,sql:%v, result is success", reflect.TypeOf(v), filter, sql)
	}
	return
}

func QueryFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.QueryFilterRow(queryer, v, filter, where, sep, args, dest...)
	return
}

func (c *CRUD) QueryFilterRow(queryer, v interface{}, filter string, where []string, sep string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.QuerySQL(v, filter)
	sql = c.JoinWhere(sql, where, sep)
	err = c.QueryRow(queryer, v, filter, sql, args, dest...)
	return
}

func QuerySimpleRow(queryer, v interface{}, filter string, suffix string, args []interface{}, dest ...interface{}) (err error) {
	err = Default.QuerySimpleRow(queryer, v, filter, suffix, args, dest...)
	return
}

func (c *CRUD) QuerySimpleRow(queryer, v interface{}, filter string, suffix string, args []interface{}, dest ...interface{}) (err error) {
	sql := c.QuerySQL(v, filter) + " " + suffix
	err = c.QueryRow(queryer, v, filter, sql, args, dest...)
	return
}
