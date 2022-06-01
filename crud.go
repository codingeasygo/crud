package crud

import (
	"fmt"
	"reflect"
	"strings"
)

func FilterField(tag string, skipNil, skipZero bool, v interface{}, found func(name string, value interface{})) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	reflectType := reflectValue.Type()
	numField := reflectType.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := reflectValue.Field(i)
		fieldName := reflectType.Field(i).Tag.Get(tag)
		fieldKind := fieldValue.Kind()
		if len(fieldName) < 1 || fieldName == "-" {
			continue
		}
		if fieldKind == reflect.Ptr && fieldValue.IsNil() && skipNil {
			continue
		}
		if fieldKind == reflect.Ptr && !fieldValue.IsNil() {
			fieldValue = reflect.Indirect(fieldValue)
		}
		if fieldValue.IsZero() && skipZero {
			continue
		}
		found(fieldName, fieldValue.Addr().Interface())
	}
}

func BuildInsert(tag string, skipNil, skipZero bool, v interface{}) (fields, param string, args []interface{}) {
	fieldsList := []string{}
	paramList := []string{}
	FilterField(tag, skipNil, skipZero, v, func(name string, value interface{}) {
		fieldsList = append(fieldsList, name)
		args = append(args, value)
		paramList = append(paramList, fmt.Sprintf("$%v", len(args)))
	})
	fields = strings.Join(fieldsList, ",")
	param = strings.Join(paramList, ",")
	return
}

func BuildUpdate(tag string, skipNil, skipZero bool, v interface{}) (sets string, args []interface{}) {
	fieldsList := []string{}
	FilterField(tag, skipNil, skipZero, v, func(name string, value interface{}) {
		args = append(args, value)
		fieldsList = append(fieldsList, fmt.Sprintf("%v=$%v", name, len(args)))
	})
	sets = strings.Join(fieldsList, ",")
	return
}
