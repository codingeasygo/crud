package gen

import (
	"reflect"
	"strings"

	"github.com/codingeasygo/util/xsql"
)

const TableSQLPG = `
SELECT
    c.relname AS name,
	c.relkind::text AS type,
    coalesce(obj_description(c.oid),'') as comment
FROM pg_class c
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1
AND c.relkind = 'r'
ORDER BY c.relname
`

const ColumnSQLPG = `
SELECT
    a.attname AS name,
    format_type(a.atttypid, a.atttypmod) AS type,
	COALESCE(ct.contype = 'p', false) AS  is_pk,
    a.attnotnull AS not_null,
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') AS default_value,
    a.attnum AS ordinal,
    CASE
        WHEN a.atttypid = ANY ('{int,int8,int2}'::regtype[])
          AND EXISTS (
             SELECT 1 FROM pg_attrdef ad
             WHERE  ad.adrelid = a.attrelid
             AND    ad.adnum   = a.attnum
             AND    pg_get_expr(ad.adbin, ad.adrelid) = 'nextval('''
                || (pg_get_serial_sequence (a.attrelid::regclass::text
                                          , a.attname))::regclass
                || '''::regclass)'
             )
            THEN CASE a.atttypid
                    WHEN 'int'::regtype  THEN 'serial'
                    WHEN 'int8'::regtype THEN 'bigserial'
                    WHEN 'int2'::regtype THEN 'smallserial'
                 END
        WHEN a.atttypid = ANY ('{uuid}'::regtype[]) AND COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') != ''
            THEN 'autogenuuid'
        ELSE format_type(a.atttypid, a.atttypmod)
    END AS ddl_type,
    coalesce(col_description(a.attrelid, a.attnum),'') AS comment
FROM pg_attribute a
JOIN ONLY pg_class c ON c.oid = a.attrelid
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
AND a.attnum = ANY(ct.conkey) AND ct.contype = 'p'
LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
WHERE a.attisdropped = false
    AND n.nspname = $1
    AND c.relname = $2
    AND a.attnum > 0
ORDER BY a.attnum
`

var TypeMapPG = map[string][]string{
	//int
	"smallint":    {"int", "*int"},
	"integer":     {"int", "*int"},
	"bigint":      {"int64", "*int64"},
	"smallserial": {"int", "*int"},
	"serial":      {"int", "*int"},
	//float
	"real":             {"decimal.Decimal", "decimal.Decimal"},
	"numeric":          {"decimal.Decimal", "decimal.Decimal"},
	"double precision": {"decimal.Decimal", "decimal.Decimal"},
	//string
	"character":         {"string", "*string"},
	"character varying": {"string", "*string"},
	"text":              {"string", "*string"},
	//time
	"time with time zone":         {"xsql.Time", "xsql.Time"},
	"time without time zone":      {"xsql.Time", "xsql.Time"},
	"timestamp with time zone":    {"xsql.Time", "xsql.Time"},
	"timestamp with without zone": {"xsql.Time", "xsql.Time"},
	"date":                        {"xsql.Time", "xsql.Time"},
	//bool
	"boolean": {"bool", "*bool"},
	//json
	"json":  {"xsql.M", "xsql.M"},
	"jsonb": {"xsql.M", "xsql.M"},
}

func NameConvPG(on, name string, field reflect.StructField) string {
	if on == "query" && strings.HasPrefix(field.Type.String(), "xsql.") && field.Type.String() != "xsql.Time" {
		return name + "::text"
	} else {
		return name
	}
}

func ParmConvPG(on, fieldName, fieldFunc string, field reflect.StructField, value interface{}) interface{} {
	if c, ok := value.(xsql.DbArrayConverter); on == "where" && ok {
		return c.DbArray()
	}
	return value
}
