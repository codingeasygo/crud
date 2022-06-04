//Package crud is simple curd tools to process database
package crud

import (
	"fmt"
	"reflect"
	"strings"
)

var Tag = "json"
var ArgFormat = "$%v"
var ErrNoRows = fmt.Errorf("no rows")

func FilterFieldCall(tag, filter string, v interface{}, call func(name string, field reflect.StructField, value interface{})) {
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
		fieldName := fieldType.Tag.Get(tag)
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
		call(fieldName, fieldType, fieldValue.Addr().Interface())
	}
}

func AppendInsert(fields, param []string, args []interface{}, ok bool, formats ...interface{}) (fields_, param_ []string, args_ []interface{}) {
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
	sql_ = sql
	if len(where) > 0 {
		sql_ += " where " + strings.Join(where, " "+sep+" ") + " " + strings.Join(suffix, " ")
	}
	return
}

func JoinPage(sql, orderby string, offset, limit int) (sql_ string) {
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
	return
}

func InsertArgs(v interface{}, filter string) (fields, param []string, args []interface{}) {
	FilterFieldCall(Tag, filter, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
		fields = append(fields, name)
		param = append(param, fmt.Sprintf(ArgFormat, len(args)))
	})
	return
}

func InsertSQL(v interface{}, filter, table string, suffix ...string) (sql string, args []interface{}) {
	fields, param, args := InsertArgs(v, filter)
	sql = fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, strings.Join(fields, ","), strings.Join(param, ","), strings.Join(suffix, " "))
	return
}

func UpdateArgs(v interface{}, filter string) (sets []string, args []interface{}) {
	FilterFieldCall(Tag, filter, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%v="+ArgFormat, name, len(args)))
	})
	return
}

func UpdateSQL(v interface{}, filter string, table string, suffix ...string) (sql string, args []interface{}) {
	sets, args := UpdateArgs(v, filter)
	sql = fmt.Sprintf(`update %v set %v %v`, table, strings.Join(sets, ","), strings.Join(suffix, " "))
	return
}

func QueryField(v interface{}, filter string) (fields string) {
	fieldsList := []string{}
	FilterFieldCall(Tag, filter, v, func(name string, field reflect.StructField, value interface{}) {
		conv := field.Tag.Get("conv")
		fieldsList = append(fieldsList, fmt.Sprintf("%v%v", name, conv))
	})
	fields = strings.Join(fieldsList, ",")
	return
}

func QuerySQL(v interface{}, filter, table string, suffix ...string) (sql string) {
	fields := QueryField(v, filter)
	sql = fmt.Sprintf(`select %v from %v %v`, fields, table, strings.Join(suffix, " "))
	return
}

func ScanArgs(v interface{}, filter string) (args []interface{}) {
	FilterFieldCall(Tag, filter, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
	})
	return
}

func destSet(value reflect.Value, dests ...interface{}) (err error) {
	valueField := func(key string) (v reflect.Value, e error) {
		targetValue := reflect.Indirect(value)
		targetType := targetValue.Type()
		k := targetValue.NumField()
		for i := 0; i < k; i++ {
			field := targetType.Field(i)
			fieldName := field.Tag.Get(Tag)
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
	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	for rows.Next() {
		value := reflect.New(reflectType)
		err = rows.Scan(ScanArgs(value.Interface(), filter)...)
		if err != nil {
			break
		}
		if !isPtr {
			value = reflect.Indirect(value)
		}
		err = destSet(value, dest...)
		if err != nil {
			break
		}
	}
	return
}

func Query(queryer Queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	rows, err := queryer.Query(sql, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	err = Scan(rows, v, filter, dest...)
	return
}

func ScanRow(rows Rows, v interface{}, filter string, dest ...interface{}) (err error) {
	isPtr := reflect.ValueOf(v).Kind() == reflect.Ptr
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	if !rows.Next() {
		err = ErrNoRows
		return
	}
	value := reflect.New(reflectType)
	err = rows.Scan(ScanArgs(value.Interface(), filter)...)
	if err != nil {
		return
	}
	if !isPtr {
		value = reflect.Indirect(value)
	}
	err = destSet(value, dest...)
	if err != nil {
		return
	}
	return
}

func QueryRow(queryer Queryer, v interface{}, filter, sql string, args []interface{}, dest ...interface{}) (err error) {
	rows, err := queryer.Query(sql, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	err = ScanRow(rows, v, filter, dest...)
	return
}
