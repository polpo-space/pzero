var (
    {{.lowerStartCamelObject}}FieldNames = condition.RawFieldNamesWithFlavor(sqlbuilder.PostgreSQL, &{{.upperStartCamelObject}}{})
    {{.lowerStartCamelObject}}Rows = strings.Join({{.lowerStartCamelObject}}FieldNames, ",")
    {{.lowerStartCamelObject}}RowsExpectAutoFieldNames = condition.RemoveIgnoreColumnsWithFlavor(sqlbuilder.PostgreSQL, {{.lowerStartCamelObject}}FieldNames, {{if .autoIncrement}}"{{.originalPrimaryKey}}", {{end}} {{.ignoreColumns}})
    {{.lowerStartCamelObject}}RowsExpectAutoSet = strings.Join({{.lowerStartCamelObject}}RowsExpectAutoFieldNames, ",")

    {{if .withCache}}{{.cacheKeys}}{{end}}
)

const (
    {{range $index, $v := .data.Fields}}{{$v.Name.ToCamel}} condition.Field = "{{$v.NameOriginal}}"
    {{end}}
)
