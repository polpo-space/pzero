package condition

import (
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/stretchr/testify/assert"
)

func TestSelectByWhereRawSql(t *testing.T) {
	sb := sqlbuilder.NewSelectBuilder()
	SelectByWhereRawSql(sb, "`sys_user_id` = ? and `sys_authority_authority_id` = ?", 1, 2)
	query, args := sb.Build()
	assert.Equal(t, `WHERE "sys_user_id" = $1 AND "sys_authority_authority_id" = $2`, query)
	assert.Equal(t, []any{1, 2}, args)
}

func TestQuote(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"manage_role.id", `"manage_role"."id"`},
		{"id", `"id"`},
		{"`manage_role`.`id`", `"manage_role"."id"`},
		{`"manage_role"."id"`, `"manage_role"."id"`},
	} {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, Quote(tc.input))
		})
	}
}
