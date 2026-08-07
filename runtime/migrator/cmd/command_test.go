package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/polpo-space/pzero/runtime/migrator"
	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type fakeMigrator struct {
	upSteps       []uint
	downSteps     []uint
	gotoVersions  []uint
	forceVersions []int
	version       uint
	dirty         bool
	err           error
	closed        bool
}

func (f *fakeMigrator) Up(steps ...uint) error {
	f.upSteps = append(f.upSteps, steps...)
	return f.err
}

func (f *fakeMigrator) Down(steps ...uint) error {
	f.downSteps = append(f.downSteps, steps...)
	return f.err
}

func (f *fakeMigrator) Goto(version uint) error {
	f.gotoVersions = append(f.gotoVersions, version)
	return f.err
}

func (f *fakeMigrator) Force(version int) error {
	f.forceVersions = append(f.forceVersions, version)
	return f.err
}

func (f *fakeMigrator) Version() (uint, bool, error) {
	return f.version, f.dirty, f.err
}

func (f *fakeMigrator) Close() (error, error) {
	f.closed = true
	return nil, nil
}

func TestMigrateUpDefaultsToAllPending(t *testing.T) {
	t.Parallel()

	fake := &fakeMigrator{}
	cmd, _ := newCommandForTest(t, fake, fixedTime(), t.TempDir())
	cmd.SetArgs([]string{"migrate", "up"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("up command failed: %v", err)
	}
	if len(fake.upSteps) != 0 {
		t.Fatalf("expected default up with no steps, got %v", fake.upSteps)
	}
	if !fake.closed {
		t.Fatal("expected migrator to be closed")
	}
}

func TestMigrateUpAndDownSteps(t *testing.T) {
	t.Parallel()

	fake := &fakeMigrator{}
	cmd, _ := newCommandForTest(t, fake, fixedTime(), t.TempDir())
	cmd.SetArgs([]string{"migrate", "up", "--steps", "3"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("up command failed: %v", err)
	}
	if len(fake.upSteps) != 1 || fake.upSteps[0] != 3 {
		t.Fatalf("expected up steps [3], got %v", fake.upSteps)
	}

	fake.closed = false
	cmd.SetArgs([]string{"migrate", "down"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("down command failed: %v", err)
	}
	if len(fake.downSteps) != 1 || fake.downSteps[0] != 1 {
		t.Fatalf("expected down steps [1], got %v", fake.downSteps)
	}
}

func TestMigrateStepsRejectNonPositiveValues(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"migrate", "up", "--steps", "0"},
		{"migrate", "down", "--steps", "-1"},
	} {
		cmd, _ := newCommandForTest(t, &fakeMigrator{}, fixedTime(), t.TempDir())
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("expected %v to reject non-positive steps", args)
		}
	}
}

func TestMigrateGotoAndForceParseVersions(t *testing.T) {
	t.Parallel()

	fake := &fakeMigrator{}
	cmd, _ := newCommandForTest(t, fake, fixedTime(), t.TempDir())
	cmd.SetArgs([]string{"migrate", "goto", "12"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("goto command failed: %v", err)
	}
	cmd.SetArgs([]string{"migrate", "force", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("force command failed: %v", err)
	}
	if len(fake.gotoVersions) != 1 || fake.gotoVersions[0] != 12 {
		t.Fatalf("expected goto version [12], got %v", fake.gotoVersions)
	}
	if len(fake.forceVersions) != 1 || fake.forceVersions[0] != 7 {
		t.Fatalf("expected force version [7], got %v", fake.forceVersions)
	}
}

func TestMigrateGotoAndForceRejectInvalidVersions(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"migrate", "goto", "not-a-version"},
		{"migrate", "force", "--", "-1"},
	} {
		cmd, _ := newCommandForTest(t, &fakeMigrator{}, fixedTime(), t.TempDir())
		cmd.SetArgs(args)
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "invalid syntax") {
			t.Fatalf("expected invalid version error for %v, got %v", args, err)
		}
	}
}

func TestMigrateVersionPrintsCurrentState(t *testing.T) {
	t.Parallel()

	cmd, out := newCommandForTest(t, &fakeMigrator{version: 42, dirty: true}, fixedTime(), t.TempDir())
	cmd.SetArgs([]string{"migrate", "version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "Current version: 42, dirty: true") {
		t.Fatalf("unexpected version output %q", got)
	}
}

func TestMigrateCreateWritesTimestampedPairWithoutLoadingConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	loaded := false
	cmd, _ := newCommandForTestWithLoader(t, &fakeMigrator{}, fixedTime(), dir, func(string) (sqlx.SqlConf, error) {
		loaded = true
		return sqlx.SqlConf{}, errors.New("config should not be loaded")
	})
	cmd.SetArgs([]string{"migrate", "create", "add local check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("create command failed: %v", err)
	}
	if loaded {
		t.Fatal("create must not load database config")
	}
	version := strconv.FormatInt(fixedTime().Unix(), 10)
	for _, name := range []string{
		version + "_add_local_check.up.sql",
		version + "_add_local_check.down.sql",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected migration file %s: %v", name, err)
		}
	}
}

func TestMigrateCreateRejectsDuplicateVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existing := strconv.FormatInt(fixedTime().Unix(), 10) + "_existing.up.sql"
	if err := os.WriteFile(filepath.Join(dir, existing), []byte("-- existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd, _ := newCommandForTest(t, &fakeMigrator{}, fixedTime(), dir)
	cmd.SetArgs([]string{"migrate", "create", "add local check"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected existing file %s to be rejected", existing)
	}
}

func TestMigrateCreateIgnoresUnrelatedSQLAndAcceptsUpstreamNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, existing := range []string{"notes.sql", "20260101_Add-Index.up.sql"} {
		if err := os.WriteFile(filepath.Join(dir, existing), []byte("-- existing\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cmd, _ := newCommandForTest(t, &fakeMigrator{}, fixedTime(), dir)
	cmd.SetArgs([]string{"migrate", "create", "add local check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("create should ignore unrelated SQL and accept upstream-compatible names: %v", err)
	}
}

func TestGeneratedVersionFits32BitMigrateRange(t *testing.T) {
	t.Parallel()

	version := fixedTime().Unix()
	if version > int64(^uint32(0)>>1) {
		t.Fatalf("generated version %d does not fit a 32-bit int", version)
	}
}

func TestMigrateUsesInheritedConfigFlagAndPropagatesErrors(t *testing.T) {
	t.Parallel()

	expected := errors.New("migration is dirty")
	cmd, _ := newCommandForTest(t, &fakeMigrator{err: expected}, fixedTime(), t.TempDir())
	cmd.SetArgs([]string{"migrate", "up", "--config", "custom.yaml"})
	err := cmd.Execute()
	if !errors.Is(err, expected) {
		t.Fatalf("expected migration error, got %v", err)
	}
}

func TestMigrateRejectsNonPgxConfig(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "app", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("config", "etc/etc.yaml", "config file")
	root.AddCommand(NewCommand(func(string) (sqlx.SqlConf, error) {
		return sqlx.SqlConf{DriverName: "mysql", DataSource: "unused"}, nil
	}))
	root.SetArgs([]string{"migrate", "up"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "only pgx is supported") {
		t.Fatalf("expected pgx-only error, got %v", err)
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 26, 9, 45, 30, 0, time.UTC)
}

func newCommandForTest(t *testing.T, fake *fakeMigrator, now time.Time, migrationDir string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	return newCommandForTestWithLoader(t, fake, now, migrationDir, func(configPath string) (sqlx.SqlConf, error) {
		if configPath != "etc/etc.yaml" && configPath != "custom.yaml" {
			t.Fatalf("unexpected config path: %s", configPath)
		}
		return sqlx.SqlConf{DriverName: "pgx", DataSource: "postgres://localhost/test"}, nil
	})
}

func newCommandForTestWithLoader(
	t *testing.T,
	fake *fakeMigrator,
	now time.Time,
	migrationDir string,
	loader SQLConfigLoader,
) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	out := &bytes.Buffer{}
	root := &cobra.Command{Use: "app", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("config", "etc/etc.yaml", "config file")
	root.SetOut(out)
	root.SetErr(out)
	root.AddCommand(newCommand(
		loader,
		withMigratorFactory(func(sqlx.SqlConf) (migrator.Migrator, error) { return fake, nil }),
		withNow(func() time.Time { return now }),
		withMigrationDir(migrationDir),
	))
	return root, out
}
