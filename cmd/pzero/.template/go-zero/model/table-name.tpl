func (m *default{{.upperStartCamelObject}}Model)withTableFields(fields ...string) []string {
    var withTableFields []string
    for _, col := range fields {
        if strings.Contains(col, ".") {
            withTableFields = append(withTableFields, condition.Quote(col))
        } else {
            withTableFields = append(withTableFields, m.table + "." + condition.Quote(col))
        }
    }
    return withTableFields
}
