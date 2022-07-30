package crud

import (
	"fmt"
	"reflect"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xsql"
	"github.com/shopspring/decimal"
)

/***** metadata:CrudObject *****/
type CrudObjectType string
type CrudObjectTypeArray []CrudObjectType

const (
	CrudObjectTypeA CrudObjectType = "1" //test a
	CrudObjectTypeB CrudObjectType = "2" //test b
	CrudObjectTypeC CrudObjectType = "3" //test c
)

//CrudObjectTypeAll is simple type in
var CrudObjectTypeAll = CrudObjectTypeArray{CrudObjectTypeA, CrudObjectTypeB, CrudObjectTypeC}

//CrudObjectTypeShow is simple type in
var CrudObjectTypeShow = CrudObjectTypeArray{CrudObjectTypeA, CrudObjectTypeB, CrudObjectTypeC}

type CrudObjectStatus int
type CrudObjectStatusArray []CrudObjectStatus

const (
	CrudObjectStatusNormal   CrudObjectStatus = 100 //
	CrudObjectStatusDisabled CrudObjectStatus = 200 //
	CrudObjectStatusRemoved  CrudObjectStatus = -1  //
)

//CrudObjectStatusAll is simple status in
var CrudObjectStatusAll = CrudObjectStatusArray{CrudObjectStatusNormal, CrudObjectStatusDisabled, CrudObjectStatusRemoved}

//CrudObjectStatusShow is simple status in
var CrudObjectStatusShow = CrudObjectStatusArray{CrudObjectStatusNormal, CrudObjectStatusDisabled}

//CrudObjectOrderbyAll is crud filter
const CrudObjectOrderbyAll = "type,update_time,create_time"

/*CrudObject  represents crud_object */
type CrudObject struct {
	_            string            `table:"crud_object"`                                           /* the table name tag */
	TID          int64             `json:"tid,omitempty" valid:"tid,r|i,r:0;"`                     /*  */
	UserID       int64             `json:"user_id,omitempty" valid:"user_id,r|i,r:0;"`             /*  */
	Type         CrudObjectType    `json:"type,omitempty" valid:"type,r|s,e:0;"`                   /* simple type in, A=1:test a, B=2:test b, C=3:test c */
	Level        int               `json:"level,omitempty" valid:"level,r|i,r:0;"`                 /*  */
	Title        string            `json:"title,omitempty" valid:"title,r|s,l:0;"`                 /*  */
	Image        *string           `json:"image,omitempty" valid:"image,r|s,l:0;"`                 /*  */
	Description  *string           `json:"description,omitempty" valid:"description,r|s,l:0;"`     /*  */
	Data         xsql.M            `json:"data,omitempty" valid:"data,r|s,l:0;"`                   /*  */
	IntValue     int               `json:"int_value,omitempty" valid:"int_value,r|i,r:0;"`         /*  */
	IntPtr       *int              `json:"int_ptr,omitempty" valid:"int_ptr,r|i,r:0;"`             /*  */
	IntArray     xsql.IntArray     `json:"int_array,omitempty" valid:"int_array,r|s,l:0;"`         /*  */
	Int64Value   int64             `json:"int64_value,omitempty" valid:"int64_value,r|i,r:0;"`     /*  */
	Int64Ptr     *int64            `json:"int64_ptr,omitempty" valid:"int64_ptr,r|i,r:0;"`         /*  */
	Int64Array   xsql.Int64Array   `json:"int64_array,omitempty" valid:"int64_array,r|s,l:0;"`     /*  */
	Float64Value decimal.Decimal   `json:"float64_value,omitempty" valid:"float64_value,r|f,r:0;"` /*  */
	Float64Ptr   decimal.Decimal   `json:"float64_ptr,omitempty" valid:"float64_ptr,r|f,r:0;"`     /*  */
	Float64Array xsql.Float64Array `json:"float64_array,omitempty" valid:"float64_array,r|s,l:0;"` /*  */
	StringValue  string            `json:"string_value,omitempty" valid:"string_value,r|s,l:0;"`   /*  */
	StringPtr    *string           `json:"string_ptr,omitempty" valid:"string_ptr,r|s,l:0;"`       /*  */
	StringArray  xsql.StringArray  `json:"string_array,omitempty" valid:"string_array,r|s,l:0;"`   /*  */
	MapValue     xsql.M            `json:"map_value,omitempty" valid:"map_value,r|s,l:0;"`         /*  */
	MapArray     xsql.MArray       `json:"map_array,omitempty" valid:"map_array,r|s,l:0;"`         /*  */
	TimeValue    xsql.Time         `json:"time_value,omitempty" valid:"time_value,r|i,r:1;"`       /*  */
	UpdateTime   xsql.Time         `json:"update_time,omitempty" valid:"update_time,r|i,r:1;"`     /*  */
	CreateTime   xsql.Time         `json:"create_time,omitempty" valid:"create_time,r|i,r:1;"`     /*  */
	Status       CrudObjectStatus  `json:"status,omitempty" valid:"status,r|i,e:0;"`               /* simple status in, Normal=100, Disabled=200, Removed=-1 */
}

//EnumValid will valid value by CrudObjectType
func (o *CrudObjectType) EnumValid(v interface{}) (err error) {
	var target CrudObjectType
	targetType := reflect.TypeOf(CrudObjectType(""))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().(CrudObjectType)
	}
	for _, value := range CrudObjectTypeAll {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", CrudObjectTypeAll)
}

//EnumValid will valid value by CrudObjectTypeArray
func (o *CrudObjectTypeArray) EnumValid(v interface{}) (err error) {
	var target CrudObjectType
	targetType := reflect.TypeOf(CrudObjectType(""))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().(CrudObjectType)
	}
	for _, value := range CrudObjectTypeAll {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", CrudObjectTypeAll)
}

//DbArray will join value to database array
func (o CrudObjectTypeArray) DbArray() (res string) {
	res = "{" + converter.JoinSafe(o, ",", converter.JoinPolicyDefault) + "}"
	return
}

//InArray will join value to database array
func (o CrudObjectTypeArray) InArray() (res string) {
	res = "'" + converter.JoinSafe(o, "','", converter.JoinPolicyDefault) + "'"
	return
}

//EnumValid will valid value by CrudObjectStatus
func (o *CrudObjectStatus) EnumValid(v interface{}) (err error) {
	var target CrudObjectStatus
	targetType := reflect.TypeOf(CrudObjectStatus(0))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().(CrudObjectStatus)
	}
	for _, value := range CrudObjectStatusAll {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", CrudObjectStatusAll)
}

//EnumValid will valid value by CrudObjectStatusArray
func (o *CrudObjectStatusArray) EnumValid(v interface{}) (err error) {
	var target CrudObjectStatus
	targetType := reflect.TypeOf(CrudObjectStatus(0))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().(CrudObjectStatus)
	}
	for _, value := range CrudObjectStatusAll {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", CrudObjectStatusAll)
}

//DbArray will join value to database array
func (o CrudObjectStatusArray) DbArray() (res string) {
	res = "{" + converter.JoinSafe(o, ",", converter.JoinPolicyDefault) + "}"
	return
}

//InArray will join value to database array
func (o CrudObjectStatusArray) InArray() (res string) {
	res = "" + converter.JoinSafe(o, ",", converter.JoinPolicyDefault) + ""
	return
}
