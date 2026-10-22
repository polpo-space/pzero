{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) FindOne(ctx context.Context, session sqlx.Session, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
	{{.cacheKey}}
	var resp {{.upperStartCamelObject}}
	err := m.cachedConn.QueryRowCtx(ctx, &resp, {{.cacheKeyVariable}}, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
	    sb := sqlbuilder.Select(m.rows).From(m.table)
	    sb.Where(sb.EQ(condition.Quote("{{.originalPrimaryKey}}"), {{.lowerStartCamelPrimaryKey}}))
        sql, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
        if session != nil {
		    return session.QueryRowCtx(ctx, v, sql, args...)
	    }
		return conn.QueryRowCtx(ctx, v, sql, args...)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
{{else}}
func (m *default{{.upperStartCamelObject}}Model) FindOne(ctx context.Context, session sqlx.Session, {{.lowerStartCamelPrimaryKey}} {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
	sb := sqlbuilder.Select(m.rows).From(m.table)
	sb.Where(sb.EQ(condition.Quote("{{.originalPrimaryKey}}"), {{.lowerStartCamelPrimaryKey}}))
	sb.Limit(1)
	sql, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	var resp {{.upperStartCamelObject}}
    var err error
    if session != nil {
    	err = session.QueryRowCtx(ctx, &resp, sql, args...)
	} else {
		err = m.conn.QueryRowCtx(ctx, &resp, sql, args...)
	}
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
{{end}}
// FindOneForUpdate locks a row using the caller's transaction and never reads the cache.
// Pass the session supplied by SqlConn.TransactCtx; this method does not own the transaction.
func (m *default{{.upperStartCamelObject}}Model) FindOneForUpdate(ctx context.Context, session sqlx.Session, id {{.dataType}}) (*{{.upperStartCamelObject}}, error) {
    if session == nil {
        return nil, modelx.ErrTransactionRequired
    }
    sb := sqlbuilder.PostgreSQL.NewSelectBuilder().Select(m.rows).From(m.table)
    sb.Where(sb.EQ(condition.Quote("{{.originalPrimaryKey}}"), id))
    sb.Limit(1).ForUpdate()
    statement, args := sb.Build()
    var row {{.upperStartCamelObject}}
    if err := session.QueryRowCtx(ctx, &row, statement, args...); err != nil {
        if err == sqlx.ErrNotFound {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return &row, nil
}
