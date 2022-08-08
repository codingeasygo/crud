package gen

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/codingeasygo/crud"
	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/util/xsql"
)

func stringTitle(v string) string {
	if len(v) < 1 {
		return v
	}
	return strings.ToUpper(v[:1]) + v[1:]
}

func ConvCamelCase(isTable bool, name string) (result string) {
	parts := strings.Split(name, "_")
	for _, part := range parts {
		result += stringTitle(part)
	}
	return
}

func ConvSizeTrim(typeMap map[string][]string, s *Struct, column *Column) (result string) {
	typ := regexp.MustCompile(`\([^\)]*\)`).ReplaceAllString(column.Type, "")
	types := typeMap[strings.ToLower(typ)]
	if len(types) < 1 {
		types = typeMap["*"]
	}
	if len(types) < 1 {
		result = "interface{}"
	} else if column.NotNull {
		result = types[0]
	} else {
		result = types[1]
	}
	return
}

func ConvKeyValueOption(s *Struct, field *Field) (remain string, result []*Option) {
	remainAll := []string{}
	for _, comment := range strings.Split(field.Comment, ",") {
		comment = strings.TrimSpace(comment)
		parts := strings.SplitN(comment, ":", 2)
		kv := strings.SplitN(parts[0], "=", 2)
		if len(kv) < 2 {
			remainAll = append(remainAll, comment)
			continue
		}
		key := strings.Trim(strings.TrimSpace(kv[0]), `"`)
		val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		comment := ""
		if field.Type == "string" {
			val = fmt.Sprintf(`"%v"`, val)
		}
		if len(parts) > 1 {
			comment = strings.TrimSpace(parts[1])
		}
		result = append(result, &Option{
			Name:    fmt.Sprintf("%v%v%v", s.Name, field.Name, key),
			Value:   val,
			Comment: comment,
		})
	}
	remain = strings.Join(remainAll, ",")
	return
}

type Column struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	IsPK         bool    `json:"is_pk"`
	NotNull      bool    `json:"not_null"`
	DefaultValue *string `json:"default_value"`
	Ordinal      int     `json:"ordinal"`
	DDLType      string  `json:"ddl_type"`
	Comment      string  `json:"comment"`
}

type Table struct {
	Schema  string    `json:"schema"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Comment string    `json:"comment"`
	Columns []*Column `json:"columns"`
}

func Query(queryer interface{}, tableSQL, columnSQL, schema string) (tables []*Table, err error) {
	tableArg := []interface{}{}
	if len(schema) > 0 {
		tableArg = append(tableArg, schema)
	}
	err = crud.Query(queryer, context.Background(), &Table{}, "name,type,comment#all", tableSQL, tableArg, &tables)
	if err != nil {
		return
	}
	for _, table := range tables {
		columnArg := []interface{}{}
		if len(schema) > 0 {
			columnArg = append(columnArg, schema)
		}
		columnArg = append(columnArg, table.Name)
		err = crud.Query(queryer, context.Background(), &Column{}, "#all", columnSQL, columnArg, &table.Columns)
		if err != nil {
			break
		}
	}
	return
}

type NameConv func(isTable bool, name string) string
type TypeConv func(typeMap map[string][]string, s *Struct, column *Column) string
type OptionConv func(s *Struct, field *Field) (comment string, options []*Option)

type Option struct {
	Name    string
	Value   string
	Comment string
}

type Field struct {
	Name     string
	Type     string
	Tag      string
	Comment  string
	Column   *Column
	Options  []*Option
	External interface{}
}

type Struct struct {
	Name     string
	Comment  string
	Table    *Table
	Fields   []*Field
	External interface{}
}

type Gen struct {
	Tables     []*Table
	TypeMap    map[string][]string
	FuncMap    template.FuncMap
	NameConv   NameConv
	TypeConv   TypeConv
	OptionConv OptionConv
	OnPre      func(*Gen, *Table) interface{}
}

func NewGen(typeMap map[string][]string, tables []*Table) (gen *Gen) {
	gen = &Gen{
		Tables:     tables,
		TypeMap:    typeMap,
		FuncMap:    template.FuncMap{},
		NameConv:   ConvCamelCase,
		TypeConv:   ConvSizeTrim,
		OptionConv: ConvKeyValueOption,
		OnPre:      nil,
	}
	gen.FuncMap["JoinOption"] = gen.JoinOption
	return
}

func (g *Gen) Funcs(funcs template.FuncMap) {
	for k, v := range funcs {
		g.FuncMap[k] = v
	}
}

func (g *Gen) AsStruct(t *Table) (s *Struct) {
	s = &Struct{
		Name:    g.NameConv(true, t.Name),
		Comment: t.Comment,
		Table:   t,
	}
	for _, col := range t.Columns {
		field := &Field{
			Tag:     col.Name,
			Comment: col.Comment,
			Column:  col,
		}
		field.Name = g.NameConv(false, col.Name)
		field.Type = g.TypeConv(g.TypeMap, s, col)
		field.Comment, field.Options = g.OptionConv(s, field)
		s.Fields = append(s.Fields, field)
	}
	return
}

func (g *Gen) convStruct(t *Table) (data interface{}) {
	if g.OnPre != nil {
		data = g.OnPre(g, t)
	} else {
		data = map[string]interface{}{
			"Struct": g.AsStruct(t),
		}
	}
	return
}

func (g *Gen) JoinOption(options []*Option, key, seq string) string {
	values := []string{}
	for _, option := range options {
		switch key {
		case "Name":
			values = append(values, option.Name)
		case "Value":
			values = append(values, option.Value)
		case "Comment":
			values = append(values, option.Comment)
		}
	}
	return strings.Join(values, seq)
}

func (g *Gen) Generate(writer io.Writer, call func(buffer io.Writer, data interface{}) error) (err error) {
	var source []byte
	for _, table := range g.Tables {
		buffer := bytes.NewBuffer(nil)
		data := g.convStruct(table)
		err = call(buffer, data)
		if err != nil {
			break
		}
		source, err = format.Source(buffer.Bytes())
		if err != nil {
			err = fmt.Errorf("format source fail with %v by \n%v", err, buffer.String())
			break
		}
		source = []byte("\n" + strings.TrimSpace(string(source)) + "\n")
		_, err = writer.Write(source)
		if err != nil {
			break
		}
	}
	return
}

func (g *Gen) GenerateByTemplate(name, tmpl string, writer io.Writer) (err error) {
	structTmpl := template.New(name).Funcs(g.FuncMap)
	_, err = structTmpl.Parse(tmpl)
	if err == nil {
		err = g.Generate(writer, structTmpl.Execute)
	}
	return
}

const (
	FieldsOptional = "optional"
	FieldsRequired = "required"
	FieldsUpdate   = "update"
	FieldsOrder    = "order"
	FieldsFind     = "find"
	FieldsScan     = "scan"
)

type AutoGen struct {
	TypeField     map[string]map[string]string
	FieldFilter   map[string]map[string]string
	CodeAddInit   map[string]string
	CodeTestInit  map[string]string
	CodeSlice     map[string]string
	Comments      map[string]map[string]string
	TableGenAdd   xsql.StringArray
	TableRetAdd   map[string]string
	TableInclude  xsql.StringArray
	TableExclude  xsql.StringArray
	TableNameType string
	Queryer       interface{}
	TableQueryer  func(queryer interface{}, tableSQL, columnSQL, schema string) (tables []*Table, err error)
	TableSQL      string
	ColumnSQL     string
	Schema        string
	TypeMap       map[string][]string
	NameConv      NameConv
	GetQueryer    string
	Out           string
	OutPackage    string
	OutStructPre  string
	OutStructFile string
	OutDefinePre  string
	OutDefineFile string
	OutFuncPre    string
	OutFuncCommon string
	OutFuncFile   string
	OutTestPre    string
	OutTestCommon string
	OutTestFile   string
}

func (g *AutoGen) FuncMap() (funcs template.FuncMap) {
	return template.FuncMap{
		"JoinShowOption":  g.JoinShowOption,
		"PrimaryField":    g.PrimaryField,
		"FieldZero":       g.FieldZero,
		"FieldType":       g.FieldType,
		"FieldTags":       g.FieldTags,
		"FieldDefineType": g.FieldDefineType,
	}
}

func (g *AutoGen) JoinShowOption(options []*Option, key, seq string) string {
	values := []string{}
	for _, option := range options {
		if strings.HasSuffix(option.Name, "Removed") {
			continue
		}
		values = append(values, option.Name)
	}
	return strings.Join(values, seq)
}

func (g *AutoGen) PrimaryField(s *Struct, key string) string {
	for _, f := range s.Fields {
		if !f.Column.IsPK {
			continue
		}
		switch key {
		case "Name":
			return f.Name
		case "Type":
			return f.Type
		case "TypeArray":
			return fmt.Sprintf("xsql.%vArray", stringTitle(f.Type))
		case "Column":
			return f.Column.Name
		default:
			return ""
		}
	}
	return ""
}

func (g *AutoGen) FieldZero(s *Struct, field *Field) (typ string) {
	switch field.Type {
	case "string":
		typ = `""`
	default:
		typ = `0`
	}
	return
}

func (g *AutoGen) FieldType(s *Struct, field *Field) (typ string) {
	if g.TypeField == nil {
		g.TypeField = map[string]map[string]string{}
	}
	if len(field.Options) > 0 {
		typ = fmt.Sprintf("%v%v", s.Name, field.Name)
		return
	}
	if typeFields, ok := g.TypeField[s.Table.Name]; ok {
		typ = typeFields[field.Column.Name]
	}
	if len(typ) < 1 {
		typ = field.Type
	}
	return
}

func (g *AutoGen) FieldTags(s *Struct, field *Field) (allTag string) {
	var tags []string
	addTag := func(format string, args ...interface{}) {
		tags = append(tags, fmt.Sprintf(format, args...))
	}
	func() { //valid
		if len(field.Options) > 0 {
			if field.Type == "string" {
				addTag(`valid:"%v,r|s,e:0;"`, field.Column.Name)
			} else {
				addTag(`valid:"%v,r|i,e:0;"`, field.Column.Name)
			}
			return
		}
		switch field.Type {
		case "int", "int64", "*int", "*int64":
			addTag(`valid:"%v,r|i,r:0;"`, field.Column.Name)
		case "string", "*string", "xsql.M":
			if field.Column.Name == "phone" {
				addTag(`valid:"%v,r|s,p:^\\d{11}$;"`, field.Column.Name)
			} else {
				addTag(`valid:"%v,r|s,l:0;"`, field.Column.Name)
			}
		case "decimal.Decimal":
			addTag(`valid:"%v,r|f,r:0;"`, field.Column.Name)
		case "xsql.Time":
			addTag(`valid:"%v,r|i,r:1;"`, field.Column.Name)
		}
	}()
	if len(tags) > 0 {
		allTag = " " + strings.Join(tags, " ")
	}
	return
}

func (g *AutoGen) FieldDefineType(s *Struct, field *Field) (result string) {
	typ := g.FieldType(s, field)
	if strings.HasPrefix(typ, "*") {
		result = stringTitle(strings.TrimPrefix(typ, "*")) + "Ptr"
	} else if strings.HasPrefix(typ, "xsql.") {
		result = strings.TrimPrefix(typ, "xsql.")
		if result == "M" {
			result = "Object"
		} else if strings.HasSuffix(result, "Array") {
			result = "Array"
		}
	} else if strings.HasPrefix(typ, "decimal.") {
		result = strings.TrimPrefix(typ, "decimal.")
	} else {
		result = stringTitle(typ)
	}
	return
}

func (g *AutoGen) OnPre(gen *Gen, table *Table) (data interface{}) {
	if g.FieldFilter == nil {
		g.FieldFilter = map[string]map[string]string{}
	}
	if g.CodeAddInit == nil {
		g.CodeAddInit = map[string]string{}
	}
	if g.CodeTestInit == nil {
		g.CodeTestInit = map[string]string{}
	}
	if g.Comments == nil {
		g.Comments = map[string]map[string]string{}
	}
	if g.TableRetAdd == nil {
		g.TableRetAdd = map[string]string{}
	}
	if g.TableGenAdd == nil {
		g.TableGenAdd = xsql.StringArray{}
	}
	if g.CodeSlice == nil {
		g.CodeSlice = map[string]string{
			"RowLock": "",
		}
	}
	if len(g.TableNameType) < 1 {
		g.TableNameType = "string"
	}
	for _, column := range table.Columns {
		comments, ok := g.Comments[table.Name]
		if !ok {
			continue
		}
		comment, ok := comments[column.Name]
		if !ok {
			continue
		}
		column.Comment = comment
	}
	s := gen.AsStruct(table)
	result := map[string]interface{}{
		"TableNameType": g.TableNameType,
		"Struct":        s,
		"Code":          g.CodeSlice,
		"GetQueryer":    g.GetQueryer,
	}
	fieldOptional := ""
	fieldRequired := ""
	fieldInsert := ""
	fieldUpdate := ""
	fieldOrder := ""
	fieldFind := ""
	fieldScan := ""
	var fieldOptionalValue xsql.StringArray
	var fieldRequiredValue xsql.StringArray
	var fieldUpdateValue xsql.StringArray
	if fieldConfig := g.FieldFilter[table.Name]; len(fieldConfig) > 0 {
		fieldOptional = fieldConfig[FieldsOptional]
		fieldRequired = fieldConfig[FieldsRequired]
		fieldUpdate = fieldConfig[FieldsUpdate]
		fieldOrder = fieldConfig[FieldsOrder]
		fieldFind = fieldConfig[FieldsFind]
		fieldScan = fieldConfig[FieldsScan]
		if len(fieldOptional) > 0 {
			fieldOptionalValue = xsql.AsStringArray(strings.SplitN(fieldOptional, "#", 2)[0])
		}
		if len(fieldRequired) > 0 {
			fieldRequiredValue = xsql.AsStringArray(strings.SplitN(fieldRequired, "#", 2)[0])
		}
		if len(fieldUpdate) > 0 {
			fieldUpdateValue = xsql.AsStringArray(strings.SplitN(fieldUpdate, "#", 2)[0])
		}
		parts := []string{}
		if len(fieldOptional) > 0 {
			parts = append(parts, fieldOptional)
		}
		if len(fieldRequired) > 0 {
			parts = append(parts, fieldRequired)
		}
		if len(fieldInsert) < 1 && len(parts) > 0 {
			fieldInsert = strings.Join(parts, ",")
		}
		if len(fieldUpdate) < 1 && len(parts) > 0 {
			fieldUpdate = strings.Join(parts, ",")
			fieldUpdateValue = xsql.AsStringArray(strings.SplitN(fieldUpdate, "#", 2)[0])
		}
	}
	if len(fieldFind) < 1 {
		fieldFind = "#all"
	}
	if len(fieldScan) < 1 {
		fieldScan = "#all"
	}
	fieldUpdateAll := []*Field{}
	for _, field := range s.Fields {
		update := fieldUpdateValue.HavingOne(field.Column.Name)
		onlyAdd := fieldRequiredValue.HavingOne(field.Column.Name) && !update
		optional := fieldOptionalValue.HavingOne(field.Column.Name)
		field.External = xmap.M{
			"Update":   update,
			"OnlyAdd":  onlyAdd,
			"Optional": optional,
		}
		if update || onlyAdd {
			fieldUpdateAll = append(fieldUpdateAll, field)
		}
	}
	result["Filter"] = map[string]interface{}{
		"Optional": fieldOptional,
		"Required": fieldRequired,
		"Insert":   fieldInsert,
		"Update":   strings.TrimPrefix(fieldUpdate+",update_time", ","),
		"Order":    fieldOrder,
		"Find":     fieldFind,
		"Scan":     fieldScan,
	}
	arg := strings.ToLower(s.Name[0:1]) + s.Name[1:]
	result["Arg"] = map[string]interface{}{
		"Name": arg,
	}
	{

		defaults := ""
		for _, field := range s.Fields {
			switch field.Type {
			case "xsql.Time":
				if field.Column.Name == "create_time" || field.Column.Name == "update_time" {
					defaults += fmt.Sprintf(`
						if %v.%v.Timestamp() < 1 {
							%v.%v = xsql.TimeNow()
						}
					`, arg, field.Name, arg, field.Name)
				}
			case "xsql.M":
				typ := g.FieldType(s, field)
				defaults += fmt.Sprintf(`
					if len(%v.%v) < 1 {
						%v.%v = %v{}
					}
				`, arg, field.Name, arg, field.Name, typ)
			}
		}
		if code, ok := g.CodeAddInit[s.Table.Name]; ok {
			defaults += strings.ReplaceAll(code, "ARG.", arg+".")
		}
		addFilter := fmt.Sprintf("^%v#all", g.PrimaryField(s, "Column"))
		addReturn := fmt.Sprintf("%v#all", g.PrimaryField(s, "Column"))
		if column, ok := g.TableRetAdd[s.Table.Name]; ok {
			if len(column) > 0 {
				addFilter = fmt.Sprintf("^%v#all", column)
				addReturn = fmt.Sprintf("%v#all", column)
			} else {
				addFilter = "#all"
				addReturn = ""
			}
		}
		result["Add"] = map[string]interface{}{
			"Defaults": defaults,
			"Filter":   addFilter,
			"Return":   addReturn,
			"Normal":   g.TableGenAdd.HavingOne(table.Name),
		}
	}
	{
		defaults := ""
		if code, ok := g.CodeTestInit[s.Table.Name]; ok {
			defaults += strings.ReplaceAll(code, "ARG.", arg+".")
		}
		result["Test"] = map[string]interface{}{
			"Defaults": defaults,
		}
	}
	{
		havingUpdateTime := false
		for _, field := range s.Fields {
			if field.Name == "UpdateTime" {
				havingUpdateTime = true
				break
			}
		}
		result["Update"] = map[string]interface{}{
			"UpdateTime": havingUpdateTime,
			"Fields":     fieldUpdateAll,
		}
	}
	data = result
	return
}

func (g *AutoGen) Generate() (err error) {
	if g.TypeMap == nil {
		g.TypeMap = map[string][]string{}
	}
	if g.TableQueryer == nil {
		g.TableQueryer = Query
	}
	if len(g.OutPackage) < 1 {
		g.OutPackage = "autogen"
	}
	if len(g.OutStructPre) < 1 {
		g.OutStructPre = `
			//auto gen models by autogen
			package %v
			import (
				"github.com/codingeasygo/util/xsql"
				"github.com/shopspring/decimal"
			)
		`
	}
	if len(g.OutDefinePre) < 1 {
		g.OutDefinePre = `
			//auto gen func by autogen
			package %v
		`
	}
	if len(g.OutFuncPre) < 1 {
		g.OutFuncPre = `
			//auto gen func by autogen
			package %v
			import (
				"reflect"
				"context"
				"fmt"
				"strings"

				"github.com/codingeasygo/crud"
				"github.com/codingeasygo/util/attrvalid"
				"github.com/codingeasygo/util/converter"
				"github.com/codingeasygo/util/xsql"
			)
		`
		if len(g.GetQueryer) > 0 && g.GetQueryer != "GetQueryer" {
			g.OutFuncPre += fmt.Sprintf(`
				var GetQueryer interface{} = func() crud.Queryer { return %v() }
			`, g.GetQueryer)
		} else if len(g.GetQueryer) > 0 && g.GetQueryer == "GetQueryer" {
			g.OutFuncPre += `
				var GetQueryer interface{} = func() crud.Queryer { panic("get crud queryer is not setted") }
			`
		}
	}
	if len(g.OutFuncCommon) < 1 {
		g.OutFuncCommon = `
			//Validable is interface to valid
			type Validable interface {
				Valid() error
			}
		`
	}
	if len(g.OutTestPre) < 1 {
		g.OutTestPre = `
			//auto gen func by autogen
			package %v
			import (
				"context"
				"fmt"
				"reflect"
				"strings"
				"testing"

				"github.com/codingeasygo/crud"
			)
		`
		if len(g.GetQueryer) < 1 {
			g.OutFuncPre += fmt.Sprintf(`
				var %v interface{} = func() crud.Queryer {
					panic("get crud queryer is not setted")
				}
			`, "GetQueryer")
		}
	}
	allTables, err := g.TableQueryer(g.Queryer, g.TableSQL, g.ColumnSQL, g.Schema)
	if err != nil {
		return
	}
	if len(allTables) < 1 {
		err = fmt.Errorf("table is not found")
		return
	}
	tables := []*Table{}
	for _, table := range allTables {
		if g.TableExclude.HavingOne(table.Name) {
			continue
		}
		if len(g.TableInclude) < 1 || g.TableInclude.HavingOne(table.Name) {
			tables = append(tables, table)
		}
	}
	{
		var source []byte
		generator := NewGen(g.TypeMap, tables)
		generator.Funcs(g.FuncMap())
		generator.NameConv = g.NameConv
		generator.OnPre = g.OnPre
		buffer := bytes.NewBuffer(nil)
		fmt.Fprintf(buffer, g.OutStructPre, g.OutPackage)
		err = generator.GenerateByTemplate("mod", StructTmpl, buffer)
		if err != nil {
			return
		}
		source, err = format.Source(buffer.Bytes())
		if err != nil {
			return
		}
		structFile := g.OutStructFile
		if len(structFile) < 1 {
			structFile = "auto_models.go"
		}
		err = ioutil.WriteFile(filepath.Join(g.Out, structFile), source, os.ModePerm)
		if err != nil {
			return
		}
	}
	{
		var source []byte
		generator := NewGen(g.TypeMap, tables)
		generator.Funcs(g.FuncMap())
		generator.NameConv = g.NameConv
		generator.OnPre = g.OnPre
		buffer := bytes.NewBuffer(nil)
		fmt.Fprintf(buffer, g.OutDefinePre, g.OutPackage)
		err = generator.GenerateByTemplate("fields", DefineTmpl, buffer)
		if err != nil {
			return
		}
		source, err = format.Source(buffer.Bytes())
		if err != nil {
			return
		}
		defineFile := g.OutDefineFile
		if len(defineFile) < 1 {
			defineFile = "auto_define.go"
		}
		err = ioutil.WriteFile(filepath.Join(g.Out, defineFile), source, os.ModePerm)
		if err != nil {
			return
		}
	}
	{
		var source []byte
		generator := NewGen(g.TypeMap, tables)
		generator.Funcs(g.FuncMap())
		generator.NameConv = g.NameConv
		generator.OnPre = g.OnPre
		buffer := bytes.NewBuffer(nil)
		fmt.Fprintf(buffer, g.OutFuncPre, g.OutPackage)
		fmt.Fprintf(buffer, "%v", g.OutFuncCommon)
		err = generator.GenerateByTemplate("func", StructFuncTmpl, buffer)
		if err != nil {
			return
		}
		source, err = format.Source(buffer.Bytes())
		if err != nil {
			return
		}
		funcFile := g.OutFuncFile
		if len(funcFile) < 1 {
			funcFile = "auto_func.go"
		}
		err = ioutil.WriteFile(filepath.Join(g.Out, funcFile), source, os.ModePerm)
		if err != nil {
			return
		}
	}
	{
		var source []byte
		generator := NewGen(g.TypeMap, tables)
		generator.Funcs(g.FuncMap())
		generator.NameConv = g.NameConv
		generator.OnPre = g.OnPre
		buffer := bytes.NewBuffer(nil)
		fmt.Fprintf(buffer, g.OutTestPre, g.OutPackage)
		fmt.Fprintf(buffer, "%v", g.OutTestCommon)
		err = generator.GenerateByTemplate("test", StructTestTmpl, buffer)
		if err != nil {
			return
		}
		source, err = format.Source(buffer.Bytes())
		if err != nil {
			return
		}
		testFile := g.OutTestFile
		if len(testFile) < 1 {
			testFile = "auto_func_test.go"
		}
		err = ioutil.WriteFile(filepath.Join(g.Out, testFile), source, os.ModePerm)
		if err != nil {
			return
		}
	}
	return
}
