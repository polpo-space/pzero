func new{{.upperStartCamelObject}}Model(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) *default{{.upperStartCamelObject}}Model {
	o := opts.DefaultApply(op...)
    var cachedConn sqlc.CachedConn
    if len(o.CacheConf) > 0 {
    	cachedConn = sqlc.NewConn(conn, o.CacheConf, o.CacheOpts...)
    }
    if o.CachedConn != nil {
    	cachedConn = *o.CachedConn
    }

    fields, insertFields := model{{.upperStartCamelObject}}Columns()

	return &default{{.upperStartCamelObject}}Model{
		cachedConn: cachedConn,
		conn:       conn,
		table:      condition.Quote("{{.data.Name.Source}}"),
		fieldNames: fields,
		rows: strings.Join(fields, ","),
		rowsExpectAutoFieldNames: insertFields,
		rowsExpectAutoSet: strings.Join(insertFields, ","),
	}
}

// NewOriginal{{.upperStartCamelObject}}Model returns a original model for the database table.
func NewOriginal{{.upperStartCamelObject}}Model(conn sqlx.SqlConn{{if .withCache}}, c cache.CacheConf, opts ...cache.Option{{end}}) {{.upperStartCamelObject}}Model {
        return &custom{{.upperStartCamelObject}}Model{
                default{{.upperStartCamelObject}}Model: new{{.upperStartCamelObject}}Model(conn{{if .withCache}}, modelx.WithCacheConf(c), modelx.WithCacheOpts(opts...){{end}}),
        }
}

func (m *default{{.upperStartCamelObject}}Model) clone() *default{{.upperStartCamelObject}}Model {
	return &default{{.upperStartCamelObject}}Model{
		cachedConn: m.cachedConn,
		conn:       m.conn,
		table:      m.table,
		fieldNames: m.fieldNames,
		rows: m.rows,
		rowsExpectAutoFieldNames: m.rowsExpectAutoFieldNames,
		rowsExpectAutoSet: m.rowsExpectAutoSet,
	}
}
