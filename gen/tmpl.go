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
	T {{.TableNameType}}  %vtable:"{{.Struct.Table.Name}}"%v /* the table name tag */
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

{{- range $i,$field := .Struct.Fields }}
{{- if $field.Options}}
//EnumValid will valid value by {{$.Struct.Name}}{{$field.Name}}
func (o *{{$.Struct.Name}}{{$field.Name}})EnumValid(v interface{}) (err error) {
	var target {{$.Struct.Name}}{{$field.Name}}
	targetType := reflect.TypeOf({{$.Struct.Name}}{{$field.Name}}({{FieldZero $.Struct $field}}))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().({{$.Struct.Name}}{{$field.Name}})
	}
	for _, value := range {{$.Struct.Name}}{{$field.Name}}All {
		if target == value {
			return nil
		}
	}
	return fmt.Errorf("must be in %v", {{$.Struct.Name}}{{$field.Name}}All)
}

//EnumValid will valid value by {{$.Struct.Name}}{{$field.Name}}Array
func (o *{{$.Struct.Name}}{{$field.Name}}Array)EnumValid(v interface{}) (err error) {
	var target {{$.Struct.Name}}{{$field.Name}}
	targetType := reflect.TypeOf({{$.Struct.Name}}{{$field.Name}}({{FieldZero $.Struct $field}}))
	targetValue := reflect.ValueOf(v)
	if targetValue.CanConvert(targetType) {
		target = targetValue.Convert(targetType).Interface().({{$.Struct.Name}}{{$field.Name}})
	}
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

//InArray will join value to database array
func (o {{$.Struct.Name}}{{$field.Name}}Array) InArray() (res string) {
	{{- if eq $field.Type "string"}}
	res = "'" + converter.JoinSafe(o, "','", converter.JoinPolicyDefault) + "'"
	{{- else}}
	res = "" + converter.JoinSafe(o, ",", converter.JoinPolicyDefault) + ""
	{{- end}}
	return
}
{{- end }}
{{- end }}

//MetaWith{{.Struct.Name}} will return {{.Struct.Table.Name}} meta data
func MetaWith{{.Struct.Name}}(fields ...interface{}) (v []interface{}) {
	v = crud.MetaWith("{{.Struct.Table.Name}}", fields...)
	return
}

//MetaWith will return {{.Struct.Table.Name}} meta data
func ({{.Arg.Name}} *{{.Struct.Name}}) MetaWith(fields ...interface{}) (v []interface{}) {
	v = crud.MetaWith("{{.Struct.Table.Name}}", fields...)
	return
}

//Meta will return {{.Struct.Table.Name}} meta data
func ({{.Arg.Name}} *{{.Struct.Name}}) Meta() (table string, fileds []string) {
	table, fileds = crud.QueryField({{.Arg.Name}}, "#all")
	return
}

//Valid will valid by filter
func ({{.Arg.Name}} *{{.Struct.Name}}) Valid() (err error) {
	if reflect.ValueOf({{.Arg.Name}}.{{PrimaryField .Struct "Name"}}).IsZero() {
		err = attrvalid.Valid({{.Arg.Name}}, {{.Struct.Name}}FilterInsert + "#all", {{.Struct.Name}}FilterOptional)
	} else {
		err = attrvalid.Valid({{.Arg.Name}}, {{.Struct.Name}}FilterUpdate, "")
	}
	return
}

//Insert will add {{.Struct.Table.Name}} to database
func ({{.Arg.Name}} *{{.Struct.Name}}) Insert(caller interface{}, ctx context.Context) (err error) {
	{{.Add.Defaults}}
	{{- if .Add.Return}}
	_, err = crud.InsertFilter(caller, ctx, {{.Arg.Name}}, "{{.Add.Filter}}", "returning", "{{.Add.Return}}")
	{{- else}}
	_, err = crud.InsertFilter(caller, ctx, {{.Arg.Name}}, "{{.Add.Filter}}", "", "")
	{{- end}}
	return
}

//UpdateFilter will update {{.Struct.Table.Name}} to database
func ({{.Arg.Name}} *{{.Struct.Name}}) UpdateFilter(caller interface{}, ctx context.Context, filter string) (err error) {
	err = {{.Arg.Name}}.UpdateFilterWheref(caller, ctx, filter, "")
	return
}

//UpdateWheref will update {{.Struct.Table.Name}} to database
func ({{.Arg.Name}} *{{.Struct.Name}}) UpdateWheref(caller interface{}, ctx context.Context, formats string, formatArgs ...interface{}) (err error) {
	err = {{.Arg.Name}}.UpdateFilterWheref(caller, ctx, {{.Struct.Name}}FilterUpdate, formats, formatArgs...)
	return
}

//UpdateFilterWheref will update {{.Struct.Table.Name}} to database
func ({{.Arg.Name}} *{{.Struct.Name}}) UpdateFilterWheref(caller interface{}, ctx context.Context, filter string, formats string, formatArgs ...interface{}) (err error) {
	{{- if .Update.UpdateTime}}
	{{.Arg.Name}}.UpdateTime = xsql.TimeNow()
	{{- end}}
	whereAll := []string{"{{PrimaryField .Struct "Column"}}=$%v"}
	whereArg := []interface{}{{"{"}}{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}}
	if len(formats) > 0 {
		whereAll = append(whereAll, formats)
		whereArg = append(whereArg, formatArgs...)
	}
	err = crud.UpdateRowWheref(caller, ctx, {{.Arg.Name}}, filter, strings.Join(whereAll, ","), whereArg...)
	return
}

{{if .Add.Normal}}
//Add{{.Struct.Name}} will add {{.Struct.Table.Name}} to database
func Add{{.Struct.Name}}({{.Arg.Name}} *{{.Struct.Name}}) (err error) {
	err = Add{{.Struct.Name}}Call(GetQueryer, context.Background(), {{.Arg.Name}})
	return
}

//Add{{.Struct.Name}} will add {{.Struct.Table.Name}} to database
func Add{{.Struct.Name}}Call(caller interface{}, ctx context.Context, {{.Arg.Name}} *{{.Struct.Name}}) (err error) {
	err = {{.Arg.Name}}.Insert(caller, ctx)
	return
}
{{end}}

//Update{{.Struct.Name}}Filter will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}Filter({{.Arg.Name}} *{{.Struct.Name}}, filter string) (err error) {
	err = Update{{.Struct.Name}}FilterCall(GetQueryer, context.Background(), {{.Arg.Name}}, filter)
	return
}

//Update{{.Struct.Name}}FilterCall will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}FilterCall(caller interface{}, ctx context.Context, {{.Arg.Name}} *{{.Struct.Name}}, filter string) (err error) {
	err = {{.Arg.Name}}.UpdateFilter(caller, ctx, filter)
	return
}

//Update{{.Struct.Name}}Wheref will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}Wheref({{.Arg.Name}} *{{.Struct.Name}}, formats string, formatArgs ...interface{}) (err error) {
	err = Update{{.Struct.Name}}WherefCall(GetQueryer, context.Background(), {{.Arg.Name}}, formats, formatArgs...)
	return
}

//Update{{.Struct.Name}}WherefCall will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}WherefCall(caller interface{}, ctx context.Context, {{.Arg.Name}} *{{.Struct.Name}}, formats string, formatArgs ...interface{}) (err error) {
	err = {{.Arg.Name}}.UpdateWheref(caller, ctx, formats, formatArgs...)
	return
}

//Update{{.Struct.Name}}FilterWheref will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}FilterWheref({{.Arg.Name}} *{{.Struct.Name}}, filter string, formats string, formatArgs ...interface{}) (err error) {
	err = Update{{.Struct.Name}}FilterWherefCall(GetQueryer, context.Background(), {{.Arg.Name}}, filter, formats, formatArgs...)
	return
}

//Update{{.Struct.Name}}FilterWherefCall will update {{.Struct.Table.Name}} to database
func Update{{.Struct.Name}}FilterWherefCall(caller interface{}, ctx context.Context, {{.Arg.Name}} *{{.Struct.Name}}, filter string, formats string, formatArgs ...interface{}) (err error) {
	err = {{.Arg.Name}}.UpdateFilterWheref(caller, ctx, filter, formats, formatArgs...)
	return
}

//Find{{.Struct.Name}}Call will find {{.Struct.Table.Name}} by id from database
func Find{{.Struct.Name}}({{.Arg.Name}}ID {{PrimaryField .Struct "Type"}}) ({{.Arg.Name}} *{{.Struct.Name}}, err error) {
	{{.Arg.Name}}, err = Find{{.Struct.Name}}Call(GetQueryer, context.Background(), {{.Arg.Name}}ID, false)
	return
}

//Find{{.Struct.Name}}Call will find {{.Struct.Table.Name}} by id from database
func Find{{.Struct.Name}}Call(caller interface{}, ctx context.Context, {{.Arg.Name}}ID {{PrimaryField .Struct "Type"}}, lock bool) ({{.Arg.Name}} *{{.Struct.Name}}, err error) {
	where, args := crud.AppendWhere(nil, nil, true, "{{PrimaryField .Struct "Column"}}=$%v", {{.Arg.Name}}ID)
	{{.Arg.Name}}, err = Find{{.Struct.Name}}WhereCall(caller, ctx, lock, "and", where, args)
	return
}

//Find{{.Struct.Name}}WhereCall will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}WhereCall(caller interface{}, ctx context.Context, lock bool, join string, where []string, args []interface{}) ({{.Arg.Name}} *{{.Struct.Name}}, err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	querySQL = crud.JoinWhere(querySQL, where, join)
	if lock {
		querySQL += " {{.Code.RowLock}} "
	}
	err = crud.QueryRow(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, args, &{{.Arg.Name}})
	return
}

//Find{{.Struct.Name}}Wheref will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}Wheref(format string, args ...interface{}) ({{.Arg.Name}} *{{.Struct.Name}}, err error) {
	{{.Arg.Name}}, err = Find{{.Struct.Name}}WherefCall(GetQueryer, context.Background(), false, format, args...)
	return
}

//Find{{.Struct.Name}}WherefCall will find {{.Struct.Table.Name}} by where from database
func Find{{.Struct.Name}}WherefCall(caller interface{}, ctx context.Context, lock bool, format string, args ...interface{}) ({{.Arg.Name}} *{{.Struct.Name}}, err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	where, queryArgs := crud.AppendWheref(nil, nil, format, args...)
	querySQL = crud.JoinWhere(querySQL, where, "and")
	if lock {
		querySQL += " {{.Code.RowLock}} "
	}
	err = crud.QueryRow(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, queryArgs, &{{.Arg.Name}})
	return
}

//List{{.Struct.Name}}ByID will list {{.Struct.Table.Name}} by id from database
func List{{.Struct.Name}}ByID({{.Arg.Name}}IDs ...{{PrimaryField .Struct "Type"}}) ({{.Arg.Name}}List []*{{.Struct.Name}}, {{.Arg.Name}}Map map[{{PrimaryField .Struct "Type"}}]*{{.Struct.Name}}, err error) {
	{{.Arg.Name}}List, {{.Arg.Name}}Map, err = List{{.Struct.Name}}ByIDCall(GetQueryer, context.Background(), {{.Arg.Name}}IDs...)
	return
}

//List{{.Struct.Name}}ByIDCall will list {{.Struct.Table.Name}} by id from database
func List{{.Struct.Name}}ByIDCall(caller interface{}, ctx context.Context, {{.Arg.Name}}IDs ...{{PrimaryField .Struct "Type"}}) ({{.Arg.Name}}List []*{{.Struct.Name}}, {{.Arg.Name}}Map map[{{PrimaryField .Struct "Type"}}]*{{.Struct.Name}}, err error) {
	if len({{.Arg.Name}}IDs) < 1 {
		{{.Arg.Name}}Map = map[{{PrimaryField .Struct "Type"}}]*{{.Struct.Name}}{}
		return
	}
	err = Scan{{.Struct.Name}}ByIDCall(caller, ctx, {{.Arg.Name}}IDs, &{{.Arg.Name}}List, &{{.Arg.Name}}Map, "{{PrimaryField .Struct "Column"}}")
	return
}

//Scan{{.Struct.Name}}ByID will list {{.Struct.Table.Name}} by id from database
func Scan{{.Struct.Name}}ByID({{.Arg.Name}}IDs []{{PrimaryField .Struct "Type"}}, dest ...interface{}) (err error) {
	err = Scan{{.Struct.Name}}ByIDCall(GetQueryer, context.Background(), {{.Arg.Name}}IDs, dest...)
	return
}

//Scan{{.Struct.Name}}ByIDCall will list {{.Struct.Table.Name}} by id from database
func Scan{{.Struct.Name}}ByIDCall(caller interface{}, ctx context.Context, {{.Arg.Name}}IDs []{{PrimaryField .Struct "Type"}}, dest ...interface{}) (err error) {
	querySQL := crud.QuerySQL(&{{.Struct.Name}}{}, "#all")
	where := append([]string{}, fmt.Sprintf("{{PrimaryField .Struct "Column"}} in (%v)", {{PrimaryField .Struct "TypeArray"}}({{.Arg.Name}}IDs).InArray()))
	querySQL = crud.JoinWhere(querySQL, where, " and ")
	err = crud.Query(caller, ctx, &{{.Struct.Name}}{}, "#all", querySQL, nil, dest...)
	return
}

//Scan{{.Struct.Name}} will list {{.Struct.Table.Name}} by format from database
func Scan{{.Struct.Name}}Wheref(format string, args []interface{}, dest ...interface{}) (err error) {
	err = Scan{{.Struct.Name}}WherefCall(GetQueryer, context.Background(), format, args, dest...)
	return
}

//Scan{{.Struct.Name}}Call will list {{.Struct.Table.Name}} by format from database
func Scan{{.Struct.Name}}WherefCall(caller interface{}, ctx context.Context, format string, args []interface{}, dest ...interface{}) (err error) {
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
	{{- range $i,$field := .Struct.Fields }}
	{{- if $field.Options}}
	for _, value := range {{$.Struct.Name}}{{$field.Name}}All {
		if value.EnumValid({{$field.Type}}(value)) != nil {
			t.Error("not enum valid")
			return
		}
		if value.EnumValid({{$field.Type}}({{FieldZero $.Struct $field}})) == nil {
			t.Error("not enum valid")
			return
		}
		if {{$.Struct.Name}}{{$field.Name}}All.EnumValid({{$field.Type}}(value)) != nil {
			t.Error("not enum valid")
			return
		}
		if {{$.Struct.Name}}{{$field.Name}}All.EnumValid({{$field.Type}}({{FieldZero $.Struct $field}})) == nil {
			t.Error("not enum valid")
			return
		}
	}
	if len({{$.Struct.Name}}{{$field.Name}}All.DbArray()) < 1 {
		t.Error("not array")
		return
	}
	if len({{$.Struct.Name}}{{$field.Name}}All.InArray()) < 1 {
		t.Error("not array")
		return
	}
	{{- end }}
	{{- end }}
	metav := MetaWith{{.Struct.Name}}()
	if len(metav) < 1 {
		t.Error("not meta")
		return
	}
	{{.Arg.Name}} := &{{.Struct.Name}}{}
	{{.Arg.Name}}.Valid()
	{{.Test.Defaults}}
	table, fields := {{.Arg.Name}}.Meta()
	if len(table) < 1 || len(fields) < 1 {
		t.Error("not meta")
		return
	}
	fmt.Println(table, "---->", strings.Join(fields, ","))
	if table := crud.Table({{.Arg.Name}}.MetaWith(int64(0))); len(table) < 1 {
		t.Error("not table")
		return
	}
	{{- if .Add.Normal}}
	err = Add{{.Struct.Name}}({{.Arg.Name}})
	{{- else}}
	err = {{.Arg.Name}}.Insert(GetQueryer, context.Background())
	{{- end}}
	if err != nil {
		t.Error(err)
		return
	}
	if reflect.ValueOf({{.Arg.Name}}.{{PrimaryField .Struct "Name"}}).IsZero() {
		t.Error("not id")
		return
	}
	{{.Arg.Name}}.Valid()
	err = Update{{.Struct.Name}}Filter({{.Arg.Name}}, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = Update{{.Struct.Name}}Wheref({{.Arg.Name}}, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = Update{{.Struct.Name}}FilterWheref({{.Arg.Name}}, {{.Struct.Name}}FilterUpdate, "{{PrimaryField .Struct "Column"}}=$%v", {{.Arg.Name}}.{{PrimaryField .Struct "Name"}})
	if err != nil {
		t.Error(err)
		return
	}
	find{{.Struct.Name}}, err := Find{{.Struct.Name}}({{.Arg.Name}}.{{PrimaryField .Struct "Name"}})
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} != find{{.Struct.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("find id error")
		return
	}
	find{{.Struct.Name}}, err = Find{{.Struct.Name}}Wheref("{{PrimaryField .Struct "Column"}}=$%v", {{.Arg.Name}}.{{PrimaryField .Struct "Name"}})
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} != find{{.Struct.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("find id error")
		return
	}
	find{{.Struct.Name}}, err = Find{{.Struct.Name}}WhereCall(GetQueryer, context.Background(), true, "and", []string{"{{PrimaryField .Struct "Column"}}=$1"}, []interface{}{{"{"}}{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}{{"}"}})
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} != find{{.Struct.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("find id error")
		return
	}
	find{{.Struct.Name}}, err = Find{{.Struct.Name}}WherefCall(GetQueryer, context.Background(), true, "{{PrimaryField .Struct "Column"}}=$%v", {{.Arg.Name}}.{{PrimaryField .Struct "Name"}})
	if err != nil {
		t.Error(err)
		return
	}
	if {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} != find{{.Struct.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("find id error")
		return
	}
	{{.Arg.Name}}List, {{.Arg.Name}}Map, err := List{{.Struct.Name}}ByID()
	if err != nil || len({{.Arg.Name}}List) > 0 || {{.Arg.Name}}Map == nil || len({{.Arg.Name}}Map) > 0 {
		t.Error(err)
		return
	}
	{{.Arg.Name}}List, {{.Arg.Name}}Map, err = List{{.Struct.Name}}ByID({{.Arg.Name}}.{{PrimaryField .Struct "Name"}})
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Arg.Name}}List) != 1 || {{.Arg.Name}}List[0].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} || len({{.Arg.Name}}Map) != 1 || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}] == nil || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("list id error")
		return
	}
	{{.Arg.Name}}List = nil
	{{.Arg.Name}}Map = nil
	err = Scan{{.Struct.Name}}ByID([]{{PrimaryField .Struct "Type"}}{{"{"}}{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}{{"}"}}, &{{.Arg.Name}}List, &{{.Arg.Name}}Map, "{{PrimaryField .Struct "Column"}}")
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Arg.Name}}List) != 1 || {{.Arg.Name}}List[0].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} || len({{.Arg.Name}}Map) != 1 || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}] == nil || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("list id error")
		return
	}
	{{.Arg.Name}}List = nil
	{{.Arg.Name}}Map = nil
	err = Scan{{.Struct.Name}}Wheref("{{PrimaryField .Struct "Column"}}=$%v", []interface{}{{"{"}}{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}{{"}"}}, &{{.Arg.Name}}List, &{{.Arg.Name}}Map, "{{PrimaryField .Struct "Column"}}")
	if err != nil {
		t.Error(err)
		return
	}
	if len({{.Arg.Name}}List) != 1 || {{.Arg.Name}}List[0].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} || len({{.Arg.Name}}Map) != 1 || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}] == nil || {{.Arg.Name}}Map[{{.Arg.Name}}.{{PrimaryField .Struct "Name"}}].{{PrimaryField .Struct "Name"}} != {{.Arg.Name}}.{{PrimaryField .Struct "Name"}} {
		t.Error("list id error")
		return
	}
}

`
