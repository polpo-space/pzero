func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, session sqlx.Session, data *{{.upperStartCamelObject}}) error {
    {{if .withCache}}{{.keys}}{{end}}
    sb := sqlbuilder.PostgreSQL.NewInsertBuilder().
        InsertInto(m.table).
        Cols(m.rowsExpectAutoSet).
        Values({{.expressionValues}})
    {{if .data.Table.PrimaryKey.AutoIncrement}}sb.Returning(condition.Quote("{{.data.Table.PrimaryKey.Name.Source}}")){{end}}
    statement, args := sb.Build()
    {{if .data.Table.PrimaryKey.AutoIncrement}}
    var primaryKey {{.data.Table.PrimaryKey.Field.DataType}}
    var err error
    if session != nil {
        err = session.QueryRowCtx(ctx, &primaryKey, statement, args...)
    } else {
        {{if .withCache}}err = m.cachedConn.QueryRowNoCacheCtx(ctx, &primaryKey, statement, args...){{else}}err = m.conn.QueryRowCtx(ctx, &primaryKey, statement, args...){{end}}
    }
    if err != nil {
        return err
    }
    data.{{.data.Table.PrimaryKey.Name.ToCamel}} = primaryKey
    {{if .withCache}}return m.cachedConn.DelCacheCtx(ctx, {{.keyValues}}){{else}}return nil{{end}}
    {{else}}
    {{if .withCache}}_, err := m.cachedConn.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
        if session != nil {
            return session.ExecCtx(ctx, statement, args...)
        }
        return conn.ExecCtx(ctx, statement, args...)
    }, {{.keyValues}})
    {{else}}var err error
    if session != nil {
        _, err = session.ExecCtx(ctx, statement, args...)
    } else {
        _, err = m.conn.ExecCtx(ctx, statement, args...)
    }{{end}}
    return err
    {{end}}
}
