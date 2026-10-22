package {{.Package}}

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestPostgresInsert(t *testing.T) {
	original := sqlbuilder.DefaultFlavor
	sqlbuilder.DefaultFlavor = sqlbuilder.MySQL
	t.Cleanup(func() { sqlbuilder.DefaultFlavor = original })

	if want := []string{`"device_id"`, `"state"`}; !slices.Equal(deviceStatesFieldNames, want) {
		t.Fatalf("field metadata before construction = %v, want %v", deviceStatesFieldNames, want)
	}
	fields := &deviceStatesFieldNames[0]
	const cached, autoIncrement = {{.Cached}}, {{.AutoIncrement}}
	databaseError, cacheError := errors.New("database failure"), errors.New("cache failure")
	for _, withSession := range []bool{false, true} {
		for _, failure := range []error{nil, databaseError, cacheError} {
			t.Run(fmt.Sprintf("session=%t/error=%v", withSession, failure), func(t *testing.T) {
				db, cachedDB, tx := &recordingConn{}, &recordingConn{}, &recordingConn{}
				c := &recordingCache{}
				cachedConn := sqlc.NewConnWithCache(cachedDB, c)
				m := NewDeviceStatesModel(db, modelx.WithCachedConn(cachedConn), modelx.WithFlavor(sqlbuilder.MySQL))
				if fields != &deviceStatesFieldNames[0] {
					t.Fatal("constructor replaced package-level field metadata")
				}
				active := db
				if cached {
					active = cachedDB
				}
				var session sqlx.Session
				if withSession {
					session, active = tx, tx
				}
				if failure == databaseError {
					active.err = failure
				}
				if failure == cacheError {
					c.err = failure
				}
				data := &DeviceStates{DeviceId: 7, State: 3}
				err := m.Insert(context.Background(), session, data)
				wantErr := failure
				if !cached && failure == cacheError {
					wantErr = nil
				}
				if !errors.Is(err, wantErr) {
					t.Fatalf("Insert error = %v, want %v", err, wantErr)
				}

				wantSQL := `INSERT INTO "device_states" ("device_id","state") VALUES ($1, $2)`
				wantArgs := []any{int64(7), int64(3)}
				if autoIncrement {
					wantSQL = `INSERT INTO "device_states" ("state") VALUES ($1) RETURNING "device_id"`
					wantArgs = []any{int64(3)}
				}
				if active.query != wantSQL || !reflect.DeepEqual(active.args, wantArgs) {
					t.Fatalf("SQL = %s %v, want %s %v", active.query, active.args, wantSQL, wantArgs)
				}
				for _, conn := range []*recordingConn{db, cachedDB, tx} {
					if conn != active && conn.calls != 0 {
						t.Fatal("Insert used the wrong connection")
					}
				}
				if active.calls != 1 || active.returning != autoIncrement {
					t.Fatalf("calls = %d, returning = %t", active.calls, active.returning)
				}
				wantID := int64(7)
				if autoIncrement && failure != databaseError {
					wantID = 42
				}
				if data.DeviceId != wantID {
					t.Fatalf("primary key = %d, want %d", data.DeviceId, wantID)
				}
				var wantKeys []string
				if cached && failure != databaseError {
					wantKeys = []string{fmt.Sprintf("cache:public:deviceStates:deviceId:%d", wantID), "cache:public:deviceStates:state:3"}
				}
				if !slices.Equal(c.deleted, wantKeys) {
					t.Fatalf("deleted keys = %v, want %v", c.deleted, wantKeys)
				}
			})
		}
	}
}

func TestPostgresQueries(t *testing.T) {
	original := sqlbuilder.DefaultFlavor
	sqlbuilder.DefaultFlavor = sqlbuilder.MySQL
	t.Cleanup(func() { sqlbuilder.DefaultFlavor = original })
	probe := errors.New("SQL captured")
	conn := &recordingConn{err: probe}
	m := NewDeviceStatesModel(conn)
	ctx := context.Background()
	data := &DeviceStates{DeviceId: 7, State: 3}
	where := condition.Condition{Field: State, Operator: condition.Equal, Value: 3}
	checks := map[string]func() error{
		"bulk insert":         func() error { return m.BulkInsert(ctx, nil, []*DeviceStates{data}) },
		"select by condition": func() error { _, err := m.FindFieldsByCondition(ctx, nil, []condition.Field{State}, where); return err },
		"count":               func() error { _, err := m.CountByCondition(ctx, nil, where); return err },
	}
	if !{{.Cached}} {
		checks["find one"] = func() error { _, err := m.FindOne(ctx, nil, 7); return err }
		checks["update"] = func() error { return m.Update(ctx, nil, data) }
		checks["delete"] = func() error { return m.Delete(ctx, nil, 7) }
		checks["update by condition"] = func() error { return m.UpdateFieldsByCondition(ctx, nil, map[string]any{"state": 4}, where) }
		checks["delete by condition"] = func() error { return m.DeleteByCondition(ctx, nil, where) }
	}
	for name, run := range checks {
		t.Run(name, func(t *testing.T) {
			conn.query = ""
			if err := run(); !errors.Is(err, probe) {
				t.Fatalf("error = %v", err)
			}
			if !strings.Contains(conn.query, `"device_states"`) || !strings.Contains(conn.query, "$1") || strings.ContainsAny(conn.query, "`?") {
				t.Fatalf("expected PostgreSQL SQL, got %s", conn.query)
			}
		})
	}
}

type recordingConn struct {
	sqlx.SqlConn
	query     string
	args      []any
	calls     int
	returning bool
	err       error
}

func (c *recordingConn) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	c.query, c.args = query, args
	c.calls++
	return nil, c.err
}

func (c *recordingConn) QueryRowCtx(_ context.Context, dest any, query string, args ...any) error {
	c.query, c.args, c.returning = query, args, true
	c.calls++
	if c.err == nil {
		*dest.(*int64) = 42
	}
	return c.err
}

type recordingCache struct {
	cache.Cache
	deleted []string
	err     error
}

func (c *recordingCache) DelCtx(_ context.Context, keys ...string) error {
	c.deleted = append(c.deleted, keys...)
	return c.err
}

func (c *recordingConn) QueryRowsPartialCtx(ctx context.Context, _ any, query string, args ...any) error {
	_, err := c.ExecCtx(ctx, query, args...)
	return err
}
