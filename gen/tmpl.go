package gen

import "fmt"

var StructTmpl = fmt.Sprintf(`
{{range $i,$field := .Struct.Fields }}
{{if $field.Options}}
const({{"\n"}}{{range $field.Options }}	{{.Name}}={{.Value}} //{{.Comment}}{{"\n"}}{{end }})
//{{$.Struct.Name}}{{$field.Name}}All is {{$field.Comment}}
var {{$.Struct.Name}}{{$field.Name}}All=[]{{$field.Type}}{{"{"}}{{JoinOption $field.Options "Name" ","}}{{"}"}}
{{end }}{{end }}
/*{{.Struct.Name}} {{ .Struct.Comment}} represents {{ .Struct.Table.Name }} */
type {{ .Struct.Name }} struct {
	_ string  %vtable:"{{.Struct.Table.Name}}"%v /* the table name tag */
{{- range .Struct.Fields }}
	{{ .Name }} {{ .Type }}  %vjson:"{{.Column.Name}}"%v /* {{ .Column.Comment }} */
{{- end }}
}
`, "`", "`", "`", "`")
