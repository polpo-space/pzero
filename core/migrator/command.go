package migrator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const MigrationDir = "desc/sql_migration"

var (
	migrationFilePattern = regexp.MustCompile(`^([0-9]+)_(.*)\.(up|down)\.sql$`)
	migrationNamePattern = regexp.MustCompile(`[\s_-]+`)
	validMigrationName   = regexp.MustCompile(`^[a-z0-9_]+$`)
)

type SQLConfigLoader func(configPath string) (sqlx.SqlConf, error)

type commandOptions struct {
	migrationDir string
	now          func() time.Time
	newMigrator  func(sqlx.SqlConf) (Migrator, error)
}

type commandOption func(*commandOptions)

func withMigratorFactory(factory func(sqlx.SqlConf) (Migrator, error)) commandOption {
	return func(options *commandOptions) {
		options.newMigrator = factory
	}
}

func withNow(now func() time.Time) commandOption {
	return func(options *commandOptions) {
		options.now = now
	}
}

func withMigrationDir(dir string) commandOption {
	return func(options *commandOptions) {
		options.migrationDir = dir
	}
}

// NewCommand builds a migrate command for a service root command. The service
// root must expose an inherited string flag named "config".
func NewCommand(loadSQLConf SQLConfigLoader) *cobra.Command {
	return newCommand(loadSQLConf)
}

func newCommand(loadSQLConf SQLConfigLoader, opts ...commandOption) *cobra.Command {
	options := &commandOptions{
		migrationDir: MigrationDir,
		now:          time.Now,
		newMigrator: func(conf sqlx.SqlConf) (Migrator, error) {
			return New(conf)
		},
	}
	for _, opt := range opts {
		opt(options)
	}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "manage SQL migrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newUpCommand(loadSQLConf, options))
	cmd.AddCommand(newDownCommand(loadSQLConf, options))
	cmd.AddCommand(newGotoCommand(loadSQLConf, options))
	cmd.AddCommand(newForceCommand(loadSQLConf, options))
	cmd.AddCommand(newVersionCommand(loadSQLConf, options))
	cmd.AddCommand(newCreateCommand(options))
	return cmd
}

func newUpCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "apply database migrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			steps, err := cmd.Flags().GetInt("steps")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("steps") && steps <= 0 {
				return fmt.Errorf("up steps must be greater than 0")
			}

			migrator, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("steps") {
				return closeAfter(migrator.Up(uint(steps)), migrator)
			}
			return closeAfter(migrator.Up(), migrator)
		},
	}
	cmd.Flags().IntP("steps", "n", 0, "apply at most this many migrations")
	return cmd
}

func newDownCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "roll back database migrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			steps, err := cmd.Flags().GetInt("steps")
			if err != nil {
				return err
			}
			if steps <= 0 {
				return fmt.Errorf("down steps must be greater than 0")
			}

			migrator, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(migrator.Down(uint(steps)), migrator)
		},
	}
	cmd.Flags().IntP("steps", "n", 1, "rollback count")
	return cmd
}

func newGotoCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "goto [version]",
		Short: "migrate to a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := parseVersion(args[0])
			if err != nil {
				return err
			}
			migrator, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(migrator.Goto(version), migrator)
		},
	}
}

func newForceCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "force [version]",
		Short: "set a migration version and clear dirty state",
		Long:  "Force sets the database migration version without running migrations. Verify the database state manually before using it.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := parseVersion(args[0])
			if err != nil {
				return err
			}
			migrator, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(migrator.Force(version), migrator)
		},
	}
}

func newVersionCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the migration version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			migrator, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			version, dirty, versionErr := migrator.Version()
			if versionErr == nil {
				_, versionErr = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %d, dirty: %t\n", version, dirty)
			}
			return closeAfter(versionErr, migrator)
		},
	}
}

func newCreateCommand(options *commandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "create a migration file pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			created, err := createMigrationFiles(options.migrationDir, args[0], options.now())
			if err != nil {
				return err
			}
			for _, filename := range created {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), filename); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func loadMigrator(cmd *cobra.Command, loadSQLConf SQLConfigLoader, options *commandOptions) (Migrator, error) {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("read inherited config flag: %w", err)
	}
	conf, err := loadSQLConf(configPath)
	if err != nil {
		return nil, err
	}
	return options.newMigrator(conf)
}

func closeAfter(primaryErr error, migrator Migrator) error {
	sourceErr, databaseErr := migrator.Close()
	return errors.Join(primaryErr, sourceErr, databaseErr)
}

func parseVersion(raw string) (uint, error) {
	version, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	maxInt := uint64(^uint(0) >> 1)
	if version > maxInt {
		return 0, fmt.Errorf("migration version %s overflows int on %d-bit architecture", raw, strconv.IntSize)
	}
	return uint(version), nil
}

func createMigrationFiles(dir, rawName string, now time.Time) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	name := normalizeMigrationName(rawName)
	if name == "" {
		return nil, fmt.Errorf("migration name cannot be empty")
	}
	if !validMigrationName.MatchString(name) {
		return nil, fmt.Errorf("migration name %q must contain only ASCII letters, digits, spaces, hyphens, or underscores", rawName)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	version := strconv.FormatInt(now.UTC().Unix(), 10)
	versions := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			continue
		}
		versions[matches[1]] = struct{}{}
	}
	if _, exists := versions[version]; exists {
		return nil, fmt.Errorf("migration version %s already exists", version)
	}

	up := filepath.Join(dir, version+"_"+name+".up.sql")
	down := filepath.Join(dir, version+"_"+name+".down.sql")
	if err := writeExclusive(up, "-- Write your up migration SQL here.\n"); err != nil {
		return nil, err
	}
	if err := writeExclusive(down, "-- Write your down migration SQL here.\n"); err != nil {
		_ = os.Remove(up)
		return nil, err
	}
	return []string{up, down}, nil
}

func normalizeMigrationName(name string) string {
	name = strings.ToLower(name)
	name = migrationNamePattern.ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}

func writeExclusive(path, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	return file.Close()
}
