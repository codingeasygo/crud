package crud

import (
	"fmt"
	"reflect"
	"strings"
)

func Join(v interface{}, sep, prefix, suffix string) string {
	vtype := reflect.TypeOf(v)
	if vtype.Kind() != reflect.Slice {
		panic("not slice")
	}
	vval := reflect.ValueOf(v)
	if vval.Len() < 1 {
		return ""
	}
	val := fmt.Sprintf("%v", reflect.Indirect(vval.Index(0)).Interface())
	for i := 1; i < vval.Len(); i++ {
		val += fmt.Sprintf("%v%v", sep, reflect.Indirect(vval.Index(i)).Interface())
	}
	return prefix + val + suffix
}

func FilterFieldCall(tag string, skipNil, skipZero bool, inc, exc string, v interface{}, call func(name string, field reflect.StructField, value interface{})) {
	if len(inc) > 0 {
		inc = "," + inc + ","
	}
	if len(exc) > 0 {
		exc = "," + exc + ","
	}
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
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
		if fieldKind == reflect.Ptr && checkValue.IsNil() && skipNil {
			continue
		}
		if fieldKind == reflect.Ptr && !checkValue.IsNil() {
			checkValue = reflect.Indirect(checkValue)
		}
		if checkValue.IsZero() && skipZero {
			continue
		}
		call(fieldName, fieldType, fieldValue.Addr().Interface())
	}
}

var Tag = "json"
var ArgFormat = "$%v"

func InsertArgs(v interface{}, skipNil, skipZero bool, inc, exc string) (fields, param string, args []interface{}) {
	fieldsList := []string{}
	paramList := []string{}
	FilterFieldCall(Tag, skipNil, skipZero, inc, exc, v, func(name string, field reflect.StructField, value interface{}) {
		fieldsList = append(fieldsList, name)
		args = append(args, value)
		paramList = append(paramList, fmt.Sprintf(ArgFormat, len(args)))
	})
	fields = strings.Join(fieldsList, ",")
	param = strings.Join(paramList, ",")
	return
}

func InsertSQL(v interface{}, skipNil, skipZero bool, inc, exc string, table string, suffix ...string) (sql string, args []interface{}) {
	fields, param, args := InsertArgs(v, skipNil, skipZero, inc, exc)
	sql = fmt.Sprintf(`insert into %v(%v) values(%v) %v`, table, fields, param, strings.Join(suffix, " "))
	return
}

func InsertAllArgs(v interface{}) (fields, param string, args []interface{}) {
	fields, param, args = InsertArgs(v, true, true, "", "")
	return
}

func InsertAllSQL(v interface{}, table string, suffix ...string) (sql string, args []interface{}) {
	sql, args = InsertSQL(v, true, true, "", "", table, suffix...)
	return
}

func UpdateArgs(v interface{}, skipNil, skipZero bool, inc, exc string) (sets string, args []interface{}) {
	fieldsList := []string{}
	FilterFieldCall(Tag, skipNil, skipZero, inc, exc, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
		fieldsList = append(fieldsList, fmt.Sprintf("%v="+ArgFormat, name, len(args)))
	})
	sets = strings.Join(fieldsList, ",")
	return
}

func UpdateAllArgs(v interface{}) (sets string, args []interface{}) {
	sets, args = UpdateArgs(v, true, true, "", "")
	return
}

func UpdateSQL(v interface{}, skipNil, skipZero bool, inc, exc string, table string, suffix ...string) (sql string, args []interface{}) {
	sets, args := UpdateArgs(v, skipNil, skipZero, inc, exc)
	sql = fmt.Sprintf(`update %v set %v %v`, table, sets, strings.Join(suffix, " "))
	return
}

func UpdateAllSQL(v interface{}, table string, suffix ...string) (sql string, args []interface{}) {
	sql, args = UpdateSQL(v, true, true, "", "", table, suffix...)
	return
}

func QueryField(v interface{}, skipNil, skipZero bool, inc, exc string) (fields string) {
	fieldsList := []string{}
	FilterFieldCall(Tag, skipNil, skipZero, inc, exc, v, func(name string, field reflect.StructField, value interface{}) {
		conv := field.Tag.Get("conv")
		fieldsList = append(fieldsList, fmt.Sprintf("%v%v", name, conv))
	})
	fields = strings.Join(fieldsList, ",")
	return
}

func QuerySQL(v interface{}, skipNil, skipZero bool, inc, exc string, table string, suffix ...string) (sql string) {
	fields := QueryField(v, skipNil, skipNil, inc, exc)
	sql = fmt.Sprintf(`select %v from %v %v`, fields, table, strings.Join(suffix, " "))
	return
}

func QueryAllField(v interface{}) (fields string) {
	fields = QueryField(v, false, false, "", "")
	return
}

func QueryAllSQL(v interface{}, table string, suffix ...string) (sql string) {
	sql = QuerySQL(v, false, false, "", "", table, suffix...)
	return
}

func AppendWhere(where []string, args []interface{}, ok bool, format string, v interface{}) (where_ []string, args_ []interface{}) {
	where_, args_ = where, args
	if ok {
		args_ = append(args_, v)
		where_ = append(where_, fmt.Sprintf(format, len(args_)))
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

func ScanArgs(v interface{}, skipNil, skipZero bool, inc, exc string) (args []interface{}) {
	FilterFieldCall(Tag, skipNil, skipZero, inc, exc, v, func(name string, field reflect.StructField, value interface{}) {
		args = append(args, value)
	})
	return
}

func ScanAllArgs(v interface{}) (args []interface{}) {
	args = ScanArgs(v, false, false, "", "")
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
				destValue.Set(reflect.Append(reflect.Indirect(destValue), value))
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
		case reflect.Ptr:
			destValue.Set(value)
		default:
			err = fmt.Errorf("type %v is not supported", destKind)
		}
		if err != nil {
			break
		}
	}
	return
}

func Scan(rows Rows, v interface{}, skipNil, skipZero bool, inc, exc string, dest ...interface{}) (err error) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	for rows.Next() {
		value := reflect.New(reflectType)
		err = rows.Scan(ScanArgs(value.Interface(), skipNil, skipZero, inc, exc)...)
		if err != nil {
			break
		}
		err = destSet(value, dest...)
		if err != nil {
			break
		}
	}
	return
}

func ScanAll(rows Rows, v interface{}, dest ...interface{}) (err error) {
	err = Scan(rows, v, false, false, "", "", dest...)
	return
}

func Query(queryer Queryer, v interface{}, skipNil, skipZero bool, inc, exc string, sql string, args []interface{}, dest ...interface{}) (err error) {
	rows, err := queryer.Query(sql, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	err = Scan(rows, v, skipNil, skipZero, inc, exc, dest...)
	return
}

func QueryAll(queryer Queryer, v interface{}, sql string, args []interface{}, dest ...interface{}) (err error) {
	err = Query(queryer, v, false, false, "", "", sql, args, dest...)
	return
}
