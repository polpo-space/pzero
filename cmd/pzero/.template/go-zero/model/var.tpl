var (
    // PostgreSQL metadata retained for handwritten model extensions. Constructors
    // never mutate it; generated operations use the model's instance fields.
    {{.lowerStartCamelObject}}FieldNames = condition.RawFieldNames(&{{.upperStartCamelObject}}{})
    {{.lowerStartCamelObject}}RowsExpectAutoFieldNames = condition.RemoveIgnoreColumns({{.lowerStartCamelObject}}FieldNames, {{if .autoIncrement}}"{{.originalPrimaryKey}}", {{end}} {{.ignoreColumns}})
    {{.lowerStartCamelObject}}Rows = strings.Join({{.lowerStartCamelObject}}FieldNames, ",")
    {{.lowerStartCamelObject}}RowsExpectAutoSet = strings.Join({{.lowerStartCamelObject}}RowsExpectAutoFieldNames, ",")

    {{if .withCache}}{{.cacheKeys}}{{end}}
)

const (
    {{range $index, $v := .data.Fields}}{{$v.Name.ToCamel}} condition.Field = "{{$v.NameOriginal}}"
    {{end}}
)

func model{{.upperStartCamelObject}}Columns() ([]string, []string) {
    return slices.Clone({{.lowerStartCamelObject}}FieldNames), slices.Clone({{.lowerStartCamelObject}}RowsExpectAutoFieldNames)
}
