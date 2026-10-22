func (m *default{{.upperStartCamelObject}}Model)withTableFields(fields ...string) []string {
    var withTableFields []string
    for _, col := range fields {
        if strings.Contains(col, ".") {
            withTableFields = append(withTableFields, condition.QuoteWithFlavor(sqlbuilder.PostgreSQL, col))
        } else {
            withTableFields = append(withTableFields, m.table + "." + condition.QuoteWithFlavor(sqlbuilder.PostgreSQL, col))
        }
    }
    return withTableFields
}
