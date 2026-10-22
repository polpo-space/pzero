func (m *default{{.upperStartCamelObject}}Model) BulkInsert(ctx context.Context, session sqlx.Session, datas []*{{.upperStartCamelObject}}) error {
    if len(datas) == 0 {
        return nil
    }

    sb := sqlbuilder.InsertInto(m.table)
    sb.Cols(m.rowsExpectAutoSet)

    {{if $.expressionValues}}for _, data := range datas {
            sb.Values({{.expressionValues}})
        }{{end}}

    statement, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

    var err error
    if session != nil {
        _, err = session.ExecCtx(ctx, statement, args...)
    } else {
        _, err = m.conn.ExecCtx(ctx, statement, args...)
    }
    return err
}

func (m *default{{.upperStartCamelObject}}Model) FindFieldsByCondition(ctx context.Context, session sqlx.Session, fields []condition.Field, conditions ...condition.Condition) ([]*{{.upperStartCamelObject}}, error) {
    if len(fields) == 0 {
        fields = condition.ToFieldSlice(m.fieldNames)
    }

	statement, args := condition.BuildSelect(sqlbuilder.Select(m.withTableFields(cast.ToStringSlice(fields)...)...).From(m.table), conditions...)

	var resp []*{{.upperStartCamelObject}}
	var err error

	if session != nil {
		err = session.QueryRowsPartialCtx(ctx, &resp, statement, args...)
	} else {
	    err = m.conn.QueryRowsPartialCtx(ctx, &resp, statement, args...)
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *default{{.upperStartCamelObject}}Model) FindByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) ([]*{{.upperStartCamelObject}}, error) {
   return m.FindFieldsByCondition(ctx, session, condition.ToFieldSlice(m.fieldNames), conditions...)
}

func (m *default{{.upperStartCamelObject}}Model) CountByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) (int64, error) {
   var countconditions []condition.Condition
   for _, cond := range conditions {
    if cond.Operator != condition.Limit && cond.Operator != condition.Offset && cond.Operator != condition.OrderBy && cond.Operator != condition.OrderByDesc && cond.Operator != condition.OrderByAsc {
   	    countconditions = append(countconditions, cond)
    }
   }

   statement, args := condition.BuildSelect(sqlbuilder.Select("count(*)").From(m.table), countconditions...)

   var (
    total int64
    err error
   )

   if session != nil {
    err = session.QueryRowCtx(ctx, &total, statement, args...)
   } else {
   	err = m.conn.QueryRowCtx(ctx, &total, statement, args...)
   }
   if err != nil {
    return 0, err
   }
   return total, nil
}

func (m *default{{.upperStartCamelObject}}Model) FindOneByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) (*{{.upperStartCamelObject}}, error) {
    return m.FindOneFieldsByCondition(ctx, session, condition.ToFieldSlice(m.fieldNames), conditions...)
}

func (m *default{{.upperStartCamelObject}}Model) FindOneFieldsByCondition(ctx context.Context, session sqlx.Session, fields []condition.Field, conditions ...condition.Condition) (*{{.upperStartCamelObject}}, error) {
	statement, args := condition.BuildSelect(sqlbuilder.Select(m.withTableFields(cast.ToStringSlice(m.fieldNames)...)...).From(m.table).Limit(1), conditions...)

	var resp {{.upperStartCamelObject}}
	var err error

	if session != nil {
		err = session.QueryRowCtx(ctx, &resp, statement, args...)
	} else {
	    err = m.conn.QueryRowCtx(ctx, &resp, statement, args...)
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *default{{.upperStartCamelObject}}Model) PageByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) ([]*{{.upperStartCamelObject}}, int64 ,error) {
	statement, args := condition.BuildSelect(sqlbuilder.Select(m.withTableFields(m.fieldNames...)...).From(m.table), conditions...)

	var resp []*{{.upperStartCamelObject}}
	var err error

	if session != nil {
		err = session.QueryRowsCtx(ctx, &resp, statement, args...)
	} else {
		err = m.conn.QueryRowsCtx(ctx, &resp, statement, args...)
	}
	if err != nil {
		return nil, 0, err
	}

	total, err := m.CountByCondition(ctx, session, conditions...)
	if err != nil {
        return nil, 0, err
    }

	return resp, total, nil
}

// UpdateFieldsByCondition keeps the existing error-only update contract.
func (m *default{{.upperStartCamelObject}}Model) UpdateFieldsByCondition(ctx context.Context, session sqlx.Session, data map[string]any, conditions ...condition.Condition) error {
    _, err := m.updateFieldsByCondition(ctx, session, data, conditions...)
    return err
}

// UpdateFieldsByConditionAffected returns the number of matching rows updated.
// Include the expected state in conditions to implement compare-and-set updates.
func (m *default{{.upperStartCamelObject}}Model) UpdateFieldsByConditionAffected(ctx context.Context, session sqlx.Session, data map[string]any, conditions ...condition.Condition) (int64, error) {
    result, err := m.updateFieldsByCondition(ctx, session, data, conditions...)
    if err != nil {
        return 0, err
    }
    if result == nil {
        return 0, nil
    }
    return result.RowsAffected()
}

{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) updateFieldsByCondition(ctx context.Context, session sqlx.Session, dataMap map[string]any, conditions ...condition.Condition) (sql.Result, error) {
    if len(dataMap) == 0 {
        return nil, nil
    }

    var fields []condition.Field
    fields = append(fields, {{.data.PrimaryKey.Name.ToCamel}}){{range $ui := .data.UniqueIndex}}
        {{range $uif := $ui}}fields = append(fields, {{$uif.Name.ToCamel}})
        {{end}}
    {{end}}
    fields = lo.Uniq(fields)

    datas, err := m.FindFieldsByCondition(ctx, session, fields, conditions...)
    if err != nil {
        return nil, err
    }

    var cacheKeys []string

    for _, data := range datas {
        cacheKeys = append(cacheKeys, {{.data.PrimaryCacheKey.DataKeyRight}}){{range $index, $uck := .data.UniqueCacheKey}}
        cacheKeys = append(cacheKeys, {{$uck.DataKeyRight}}){{end}}
    }

	return m.cachedConn.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		statement, args := condition.BuildUpdate(sqlbuilder.Update(m.table), dataMap, conditions...)
		if session != nil {
			return session.ExecCtx(ctx, statement, args...)
		}
		return conn.ExecCtx(ctx, statement, args...)
	}, cacheKeys...)
}
{{else}}
func (m *default{{.upperStartCamelObject}}Model) updateFieldsByCondition(ctx context.Context, session sqlx.Session, data map[string]any, conditions ...condition.Condition) (sql.Result, error) {
    if len(data) == 0 {
        return nil, nil
    }

	statement, args := condition.BuildUpdate(sqlbuilder.Update(m.table), data, conditions...)

	if session != nil {
		return session.ExecCtx(ctx, statement, args...)
	}
	return m.conn.ExecCtx(ctx, statement, args...)
}
{{end}}

{{if .withCache}}
func (m *default{{.upperStartCamelObject}}Model) DeleteByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) error {
    if len(conditions) == 0 {
		return nil
	}

    var fields []condition.Field
    fields = append(fields, {{.data.PrimaryKey.Name.ToCamel}}){{range $ui := .data.UniqueIndex}}
        {{range $uif := $ui}}fields = append(fields, {{$uif.Name.ToCamel}})
        {{end}}
    {{end}}
    fields = lo.Uniq(fields)

    datas, err := m.FindFieldsByCondition(ctx, session, fields, conditions...)
    if err != nil {
    	return err
    }

    var cacheKeys []string

    for _, data := range datas {
        cacheKeys = append(cacheKeys, {{.data.PrimaryCacheKey.DataKeyRight}}){{range $index, $uck := .data.UniqueCacheKey}}
        cacheKeys = append(cacheKeys, {{$uck.DataKeyRight}}){{end}}
    }

	_, err = m.cachedConn.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		statement, args := condition.BuildDelete(sqlbuilder.DeleteFrom(m.table), conditions...)
		if session != nil {
			return session.ExecCtx(ctx, statement, args...)
		}
		return conn.ExecCtx(ctx, statement, args...)
	}, cacheKeys...)
	return err
}
{{else}}
func (m *default{{.upperStartCamelObject}}Model) DeleteByCondition(ctx context.Context, session sqlx.Session, conditions ...condition.Condition) error {
    if len(conditions) == 0 {
		return nil
	}
	statement, args := condition.BuildDelete(sqlbuilder.DeleteFrom(m.table), conditions...)

	var err error
	if session != nil {
		_, err = session.ExecCtx(ctx, statement, args...)
	} else {
		_, err = m.conn.ExecCtx(ctx, statement, args...)
	}
	return err
}
{{end}}
