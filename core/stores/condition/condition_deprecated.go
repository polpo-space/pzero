package condition

import (
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/spf13/cast"
)

// AdaptTable: return quote table
// Deprecated use Quote instead
func AdaptTable(table string) string {
	return Quote(table)
}

// AdaptField: return quote field
// Deprecated use Quote instead
func AdaptField(field string) string {
	return Quote(field)
}

// Select return select with condition builder
// Deprecated: Use BuildSelect instead
func Select(builder sqlbuilder.SelectBuilder, conditions ...Condition) sqlbuilder.SelectBuilder {
	builder.SetFlavor(sqlbuilder.PostgreSQL)
	clause := whereClause(conditions...)
	for _, c := range conditions {
		if c.SkipFunc != nil {
			c.Skip = c.SkipFunc()
		}
		if c.Skip {
			continue
		}
		if c.ValueFunc != nil {
			c.Value = c.ValueFunc()
		}
		switch Operator(strings.ToUpper(string(c.Operator))) {
		case Limit:
			builder.Limit(cast.ToInt(c.Value))
		case Offset:
			builder.Offset(cast.ToInt(c.Value))
		case OrderBy:
			builder.OrderBy(cast.ToStringSlice(ToSlice(c.Value))...)
		case OrderByDesc:
			builder.OrderByDesc(Quote(string(c.Field)))
		case OrderByAsc:
			builder.OrderByAsc(Quote(string(c.Field)))
		case GroupBy:
			// compatibility with old version
			if c.Value != nil {
				builder.GroupBy(cast.ToStringSlice(ToSlice(c.Value))...)
			} else {
				builder.GroupBy(Quote(string(c.Field)))
			}
		case Join:
			builder.JoinWithOption(c.JoinCondition.Option, c.JoinCondition.Table, cast.ToStringSlice(ToSlice(c.JoinCondition.OnExpr))...)
		}
	}
	if clause != nil {
		builder = *builder.AddWhereClause(clause)
	}
	return builder
}

// Update return update with condition builder
// Deprecated: Use BuildUpdate instead
func Update(builder sqlbuilder.UpdateBuilder, conditions ...Condition) sqlbuilder.UpdateBuilder {
	builder.SetFlavor(sqlbuilder.PostgreSQL)
	clause := whereClause(conditions...)
	for _, c := range conditions {
		if c.SkipFunc != nil {
			c.Skip = c.SkipFunc()
		}
		if c.Skip {
			continue
		}
		if c.ValueFunc != nil {
			c.Value = c.ValueFunc()
		}
		switch Operator(strings.ToUpper(string(c.Operator))) {
		case Limit:
			builder.Limit(cast.ToInt(c.Value))
		case OrderBy:
			builder.OrderBy(cast.ToStringSlice(ToSlice(c.Value))...)
		case OrderByDesc:
			builder.OrderByDesc(Quote(string(c.Field)))
		case OrderByAsc:
			builder.OrderByAsc(Quote(string(c.Field)))
		}
	}
	if clause != nil {
		builder = *builder.AddWhereClause(clause)
	}
	return builder
}

// Delete return delete with condition builder
// Deprecated: Use BuildDelete instead
func Delete(builder sqlbuilder.DeleteBuilder, conditions ...Condition) sqlbuilder.DeleteBuilder {
	builder.SetFlavor(sqlbuilder.PostgreSQL)
	clause := whereClause(conditions...)
	for _, c := range conditions {
		if c.SkipFunc != nil {
			c.Skip = c.SkipFunc()
		}
		if c.Skip {
			continue
		}
		if c.ValueFunc != nil {
			c.Value = c.ValueFunc()
		}
		switch Operator(strings.ToUpper(string(c.Operator))) {
		case Limit:
			builder.Limit(cast.ToInt(c.Value))
		case OrderBy:
			builder.OrderBy(cast.ToStringSlice(ToSlice(c.Value))...)
		case OrderByDesc:
			builder.OrderByDesc(Quote(string(c.Field)))
		case OrderByAsc:
			builder.OrderByAsc(Quote(string(c.Field)))
		}
	}
	if clause != nil {
		builder = *builder.AddWhereClause(clause)
	}
	return builder
}
