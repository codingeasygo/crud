package gen

import (
	"fmt"
)

var StructTmpl = fmt.Sprintf(`
/***** metadata:{{.Struct.Name}} *****/
{{- range $i,$field := .Struct.Fields }}
{{- if $field.Options}}
type {{$.Struct.Name}}{{$field.Name}} {{$field.Type}}
type {{$.Struct.Name}}{{$field.Name}}Array []{{$.Struct.Name}}{{$field.Name}}
const({{"\n"}}{{range $field.Options }}	{{.Name}} {{$.Struct.Name}}{{$field.Name}} ={{.Value}} //{{.Comment}}{{"\n"}}{{end }})
//{{$.Struct.Name}}{{$field.Name}}All is {{$field.Comment}}
var {{$.Struct.Name}}{{$field.Name}}All={{$.Struct.Name}}{{$field.Name}}Array{{"{"}}{{JoinOption $field.Options "Name" ","}}{{"}"}}
//{{$.Struct.Name}}{{$field.Name}}Show is {{$field.Comment}}
var {{$.Struct.Name}}{{$field.Name}}Show={{$.Struct.Name}}{{$field.Name}}Array{{"{"}}{{JoinShowOption $field.Options "Name" ","}}{{"}"}}
{{- end }}
{{- end }}

{{- if .Filter.Order}}
//{{.Struct.Name}}OrderbyAll is crud filter
const {{.Struct.Name}}OrderbyAll = "{{.Filter.Order}}"
{{- end }}

/*{{.Struct.Name}} {{ .Struct.Comment}} represents {{ .Struct.Table.Name }} */
type {{ .Struct.Name }} struct {
	_ string  %vtable:"{{.Struct.Table.Name}}"%v /* the table name tag */
{{- range .Struct.Fields }}
	{{ .Name }} {{FieldType $.Struct . }}  %vjson:"{{.Column.Name}},omitempty"{{FieldTags $.Struct . }}%v /* {{ .Column.Comment }} */
{{- end }}
}
`, "`", "`", "`", "`")

var DefineTmpl = `
/**
 * @apiDefine {{.Struct.Name}}Update
{{- range .Update.Fields }}
{{- if not .External.Optional}}
 * @apiParam ({{$.Struct.Name}}) {{"{"}}{{FieldDefineType $.Struct . }}{{"}"}} {{$.Struct.Name}}.{{.Column.Name}} only required when add, {{ .Comment }}{{if .Options}}, all suported is <a href="#metadata-{{$.Struct.Name}}">{{$.Struct.Name}}{{.Name}}All</a>{{end}}
 {{- end }}
 {{- end }}
{{- range .Update.Fields }}
{{- if .External.Optional}}
 * @apiParam ({{$.Struct.Name}}) {{"{"}}{{FieldDefineType $.Struct . }}{{"}"}} [{{$.Struct.Name}}.{{.Column.Name}}] {{ .Comment }}{{if .Options}}, all suported is <a href="#metadata-{{$.Struct.Name}}">{{$.Struct.Name}}{{.Name}}All</a>{{end}}
 {{- end }}
 {{- end }}
 */
/**
 * @apiDefine {{.Struct.Name}}Object
{{- range .Struct.Fields }}
 * @apiSuccess ({{$.Struct.Name}}) {{"{"}}{{FieldDefineType $.Struct . }}{{"}"}} {{$.Struct.Name}}.{{.Column.Name}} {{ .Comment }}{{if .Options}}, all suported is <a href="#metadata-{{$.Struct.Name}}">{{$.Struct.Name}}{{.Name}}All</a>{{end}}
{{- end }}
 */
`

var StructFuncTmpl = `
//{{.Struct.Name}}FilterOptional is crud filter
const {{.Struct.Name}}FilterOptional = "{{.Filter.Optional}}"

//{{.Struct.Name}}FilterRequired is crud filter
const {{.Struct.Name}}FilterRequired = "{{.Filter.Required}}"

//{{.Struct.Name}}FilterInsert is crud filter
const {{.Struct.Name}}FilterInsert = "{{.Filter.Insert}}"

//{{.Struct.Name}}FilterUpdate is crud filter
const {{.Struct.Name}}FilterUpdate = "{{.Filter.Update}}"

{{range $i,$field := .Struct.Fields }}
{{if $field.Options}}
//EnumValid will valid value by {{$.Struct.Name}}{{$field.Name}}
func (o *{{$.Struct.Name}}{{$field.Name}})EnumValid(v interface{}) (err error) {
	var target = {{$.Struct.Name}}{{$field.Name}}(v.(int64))
	for _, value := range {{$.Struct.Name}}{{$field.Name}}All {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", {{$.Struct.Name}}{{$field.Name}}All)
}

//EnumValid will valid value by {{$.Struct.Name}}{{$field.Name}}Array
func (o *{{$.Struct.Name}}{{$field.Name}}Array)EnumValid(v interface{}) (err error) {
	var target = {{$.Struct.Name}}{{$field.Name}}(v.(int64))
	for _, value := range {{$.Struct.Name}}{{$field.Name}}All {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", {{$.Struct.Name}}{{$field.Name}}All)
}

//DbArray will join value to database array
func (o {{$.Struct.Name}}{{$field.Name}}Array) DbArray() (res string) {
	res = "{" + converter.JoinSafe(o, ",", converter.JoinPolicyDefault) + "}"
	return
}
{{end }}{{end }}

//MetaWith{{.Struct.Name}} will return {{.Struct.Table.Name}} meta data
func MetaWith{{.Struct.Name}}(fields ...interface{}) (v []interface{}) {
	v = crud.MetaWith("{{.Struct.Table.Name}}", fields...)
	return
}

//MetaWith will return {{.Struct.Table.Name}} meta data
func ({{.Add.Arg}} *{{.Struct.Name}}) MetaWith(fields ...interface{}) (v []interface{}) {
	v = crud.MetaWith("{{.Struct.Table.Name}}", fields...)
	return
}

//Meta will return {{.Struct.Table.Name}} meta data
func ({{.Add.Arg}} *{{.Struct.Name}}) Meta() (table string, fileds []string) {
	table, fileds = crud.QueryField({{.Add.Arg}}, "#all")
	return
}

//Valid will valid by filter
func ({{.Add.Arg}} *{{.Struct.Name}}) Valid() (err error) {
	if {{.Add.Arg}}.TID > 0 {
		err = attrvalid.Valid({{.Add.Arg}}, {{.Struct.Name}}FilterUpdate, "")
	} else {
		err = attrvalid.Valid({{.Add.Arg}}, {{.Struct.Name}}FilterInsert + "#all", {{.Struct.Name}}FilterOptional)
	}
	return
}

//Insert will add {{.Struct.Table.Name}} to database
func ({{.Add.Arg}} *{{.Struct.Name}}) Insert(caller crud.Queryer, ctx context.Context) (err error) {
	{{.Add.Defaults}}
	{{- if .Add.Return}}
	_, err = crud.InsertFilter(caller, ctx, {{.Add.Arg}}, "{{.Add.Filter}}", "returning", "{{.Add.Return}}")
	{{- else}}
	_, err = crud.InsertFilter(caller, ctx, {{.Add.Arg}}, "{{.Add.Filter}}", "", "")
	{{- end}}
	return
}

//Update will update {{.Struct.Table.Name}} to database
func ({{.Add.Arg}} *{{.Struct.Name}}) Update(caller crud.Queryer, ctx context.Context, filter string) (err error) {
	err = {{.Add.Arg}}.UpdateFilterWheref(caller, ctx, filter, "")
	return
}

//UpdateWheref will update {{.Struct.Table.Name}} to database
func ({{.Add.Arg}} *{{.Struct.Name}}) UpdateWheref(caller crud.Queryer, ctx context.Context, formats string, formatArgs ...interface{}) (err error) {
	err = {{.Add.Arg}}.UpdateFilterWheref(caller, ctx, {{.Struct.Name}}FilterUpdate, formats, formatArgs...)
	return
}

//UpdateFilterWheref will update {{.Struct.Table.Name}} to database
func ({{.Add.Arg}} *{{.Struct.Name}}) UpdateFilterWheref(caller crud.Queryer, ctx context.Context, filter string, formats string, formatArgs ...interface{}) (err error) {
	{{- if .Update.UpdateTime}}
	{{.Add.Arg}}.UpdateTime = xsql.TimeNow()
	{{- end}}
	where := []string{}
	updateSQL, args := crud.UpdateSQL({{.Add.Arg}}, filter, nil)
	where, args = crud.AppendWhere(where, args, true, "tid=$%v", {{.Add.Arg}}.TID)
	if len(formats) > 0 {
		where, args = crud.AppendWheref(where, args, formats, formatArgs...)
	}
	updateSQL = crud.JoinWhere(updateSQL, where, " and ")
	err = caller.ExecRow(updateSQL, args...)
	return
}

{{if .Add.Normal}}
//Add{{.Struct.Name}} will add {{.Struct.Table.Name}} to database
func Add{{.Struct.Name}}({{.Add.Arg}} *{{.Struct.Name}}) (err error) {
	err = {{.Add.Arg}}.Insert(Pool())
	return
}
{{end}}

//Find{{.Struct.Name}}Call will find {{.Struct.Table.Name}} by id from database
func Find{{.Struct.Name}}({{.Add.Arg}}ID int64) ({{.Add.Arg}} *{{.Struct.Name}}, err error) {
	{{.Add.Arg}}, err = Find{{.Struct.Name}}Call(Pool(), {{.Add.Arg}}ID, false)
	return
}

//Find{{.Struct.Name}}Call will find {{.Struct.Table.Name}} by id from database
func Find{{.Struct.Name}}Call(caller crud.Queryer, ctx context.Context, {{.Add.Arg}}ID int64, lock bool) ({{.Add.Arg}} *{{.Struct.Name}}, err error) {
	where, args := crud.AppendWhere(nil, nil, true, "tid=$%v", {{.Add.Arg}}ID)
	{{.Add.Arg}}, err = Find{{.Struct.Name}}WhereCall(caller, ctx, lock, "and", where, args)
	return
}

//Find{{.Struct.Name}}WhereCall will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}WhereCall(caller crud.Queryer, ctx context.Context, lock bool, join string, where []string, args []interface{}) ({{.Add.Arg}} *{{.Struct.Name}}, err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	querySQL = crud.JoinWhere(querySQL, where, join)
	if lock {
		querySQL += " for update "
	}
	err = crud.QueryRow(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, args, &{{.Add.Arg}})
	return
}

//Find{{.Struct.Name}}Wheref will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}Wheref(format string, args ...interface{}) ({{.Add.Arg}} *{{.Struct.Name}}, err error) {
	{{.Add.Arg}}, err = Find{{.Struct.Name}}WherefCall(Pool(), false, format, args...)
	return
}

//Find{{.Struct.Name}}WherefCall will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}WherefCall(caller crud.Queryer, ctx context.Context, lock bool, format string, args ...interface{}) ({{.Add.Arg}} *{{.Struct.Name}}, err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	where, queryArgs := crud.AppendWheref(nil, nil, format, args...)
	querySQL = crud.JoinWhere(querySQL, where, "and")
	if lock {
		querySQL += " for update "
	}
	err = crud.QueryRow(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, queryArgs, &{{.Add.Arg}})
	return
}

//List{{.Struct.Name}}ByID will list {{.Struct.Table.Name}} by id from database
func List{{.Struct.Name}}ByID({{.Add.Arg}}IDs ...int64) ({{.Add.Arg}}List []*{{.Struct.Name}}, {{.Add.Arg}}Map map[int64]*{{.Struct.Name}}, err error) {
	{{.Add.Arg}}List, {{.Add.Arg}}Map, err = List{{.Struct.Name}}ByIDCall(Pool(), {{.Add.Arg}}IDs...)
	return
}

//List{{.Struct.Name}}ByIDCall will list {{.Struct.Table.Name}} by id from database
func List{{.Struct.Name}}ByIDCall(caller crud.Queryer, ctx context.Context, {{.Add.Arg}}IDs ...int64) ({{.Add.Arg}}List []*{{.Struct.Name}}, {{.Add.Arg}}Map map[int64]*{{.Struct.Name}}, err error) {
	if len({{.Add.Arg}}IDs) < 1 {
		{{.Add.Arg}}Map = map[int64]*{{.Struct.Name}}{}
		return
	}
	err = Scan{{.Struct.Name}}ByIDCall(caller, ctx, {{.Add.Arg}}IDs, &{{.Add.Arg}}List, &{{.Add.Arg}}Map, "tid")
	return
}

//Scan{{.Struct.Name}}ByID will list {{.Struct.Table.Name}} by id from database
func Scan{{.Struct.Name}}ByID({{.Add.Arg}}IDs []int64, dest ...interface{}) (err error) {
	err = Scan{{.Struct.Name}}ByIDCall(Pool(), {{.Add.Arg}}IDs, dest...)
	return
}

//Scan{{.Struct.Name}}ByIDCall will list {{.Struct.Table.Name}} by id from database
func Scan{{.Struct.Name}}ByIDCall(caller crud.Queryer, ctx context.Context, {{.Add.Arg}}IDs []int64, dest ...interface{}) (err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	where, args := crud.AppendWhere(nil, nil, true, "tid=any($%v)", xsql.Int64Array({{.Add.Arg}}IDs).DbArray())
	querySQL = crud.JoinWhere(querySQL, where, " and ")
	err = crud.Query(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, args, dest...)
	return
}

//Scan{{.Struct.Name}} will list {{.Struct.Table.Name}} by format from database
func Scan{{.Struct.Name}}Wheref(format string, args []interface{}, dest ...interface{}) (err error) {
	err = Scan{{.Struct.Name}}WherefCall(Pool(), format, args, dest...)
	return
}

//Scan{{.Struct.Name}}Call will list {{.Struct.Table.Name}} by format from database
func Scan{{.Struct.Name}}WherefCall(caller crud.Queryer, ctx context.Context, format string, args []interface{}, dest ...interface{}) (err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	var where []string
	if len(format) > 0 {
		where, args = crud.AppendWheref(nil, nil, format, args...)
	}
	querySQL = crud.JoinWhere(querySQL, where, " and ")
	err = crud.Query(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, args, dest...)
	return
}

`

var StructTestTmpl = `
func TestAuto{{.Struct.Name}}(t *testing.T) {
	var err error
	{{range $i,$field := .Struct.Fields }}
	{{if $field.Options}}
	for _, value := range {{$.Struct.Name}}{{$field.Name}}All {
		if value.EnumValid(int64(value)) != nil {
			t.Error("not enum valid")
			return
		}
		if value.EnumValid(int64(0)) == nil {
			t.Error("not enum valid")
			return
		}
		if {{$.Struct.Name}}{{$field.Name}}All.EnumValid(int64(value)) != nil {
			t.Error("not enum valid")
			return
		}
		if {{$.Struct.Name}}{{$field.Name}}All.EnumValid(int64(0)) == nil {
			t.Error("not enum valid")
			return
		}
	}
	if len({{$.Struct.Name}}{{$field.Name}}All.DbArray()) < 1 {
		t.Error("not array")
		return
	}
	{{end }}{{end }}
	metav := MetaWith{{.Struct.Name}}()
	if len(metav) < 1 {
		t.Error("not meta")
		return
	}
	{{.Add.Arg}} := &{{.Struct.Name}}{}
	table, fields := {{.Add.Arg}}.Meta()
	if len(table) < 1 || len(fields) < 1 {
		t.Error("not meta")
		return
	}
	fmt.Println(table, "---->", strings.Join(fields, ","))
	if table := crud.Table({{.Add.Arg}}.MetaWith(int64(0))); len(table) < 1 {
		t.Error("not table")
		return
	}
	{{- if .Add.Normal}}
	err = Add{{.Struct.Name}}({{.Add.Arg}})
	{{- else}}
	err = {{.Add.Arg}}.Insert(Pool())
	{{- end}}
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Add.Arg}}.TID < 1 {
		t.Error("not id")
		return
	}
	find{{.Struct.Name}}, err := Find{{.Struct.Name}}({{.Add.Arg}}.TID)
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Add.Arg}}.TID != find{{.Struct.Name}}.TID {
		t.Error("find id error")
		return
	}
	find{{.Struct.Name}}, err = Find{{.Struct.Name}}Wheref("tid=$%v", {{.Add.Arg}}.TID)
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Add.Arg}}.TID != find{{.Struct.Name}}.TID {
		t.Error("find id error")
		return
	}
	{{.Add.Arg}}List, {{.Add.Arg}}Map, err := List{{.Struct.Name}}ByID()
	if err != nil || len({{.Add.Arg}}List) > 0 || {{.Add.Arg}}Map == nil || len({{.Add.Arg}}Map) > 0 {
		t.Error(err)
		return
	}
	{{.Add.Arg}}List, {{.Add.Arg}}Map, err = List{{.Struct.Name}}ByID({{.Add.Arg}}.TID)
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Add.Arg}}List) != 1 || {{.Add.Arg}}List[0].TID != {{.Add.Arg}}.TID || len({{.Add.Arg}}Map) != 1 || {{.Add.Arg}}Map[{{.Add.Arg}}.TID] == nil || {{.Add.Arg}}Map[{{.Add.Arg}}.TID].TID != {{.Add.Arg}}.TID {
		t.Error("list id error")
		return
	}
	{{.Add.Arg}}List = nil
	{{.Add.Arg}}Map = nil
	err = Scan{{.Struct.Name}}ByID([]int64{{"{"}}{{.Add.Arg}}.TID{{"}"}}, &{{.Add.Arg}}List, &{{.Add.Arg}}Map, "tid")
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Add.Arg}}List) != 1 || {{.Add.Arg}}List[0].TID != {{.Add.Arg}}.TID || len({{.Add.Arg}}Map) != 1 || {{.Add.Arg}}Map[{{.Add.Arg}}.TID] == nil || {{.Add.Arg}}Map[{{.Add.Arg}}.TID].TID != {{.Add.Arg}}.TID {
		t.Error("list id error")
		return
	}
	{{.Add.Arg}}List = nil
	{{.Add.Arg}}Map = nil
	err = Scan{{.Struct.Name}}Wheref("tid=$%v", []interface{}{{"{"}}{{.Add.Arg}}.TID{{"}"}}, &{{.Add.Arg}}List, &{{.Add.Arg}}Map, "tid")
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Add.Arg}}List) != 1 || {{.Add.Arg}}List[0].TID != {{.Add.Arg}}.TID || len({{.Add.Arg}}Map) != 1 || {{.Add.Arg}}Map[{{.Add.Arg}}.TID] == nil || {{.Add.Arg}}Map[{{.Add.Arg}}.TID].TID != {{.Add.Arg}}.TID {
		t.Error("list id error")
		return
	}
	tx, err := Pool().Begin()
	if err != nil {
		t.Error(err)
		return
	}

	if err != nil {
		defer tx.Rollback()
	}

	find{{.Struct.Name}}, err = Find{{.Struct.Name}}Call(tx, {{.Add.Arg}}.TID, true)
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Add.Arg}}.TID != find{{.Struct.Name}}.TID {
		t.Error("find id error")
		return
	}
	err = find{{.Struct.Name}}.UpdateWheref(tx, "tid>$%v", 0)
	if err != nil {
		t.Error(err)
		return
	}
	err = find{{.Struct.Name}}.Update(tx, "")
	if err != nil {
		t.Error(err)
		return
	}
	find{{.Struct.Name}}, err = Find{{.Struct.Name}}WherefCall(tx, true, "tid=$%v", {{.Add.Arg}}.TID)
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Add.Arg}}.TID != find{{.Struct.Name}}.TID {
		t.Error("find id error")
		return
	}

	tx.Commit()

	(&{{.Struct.Name}}{}).Valid()
	(&{{.Struct.Name}}{TID: 10}).Valid()
}

`
