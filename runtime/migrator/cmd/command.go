package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/polpo-space/pzero/runtime/migrator"
)

const MigrationDir = "desc/sql_migration"

type SQLConfigLoader func(configPath string) (sqlx.SqlConf, error)

type commandOptions struct {
	migrationDir string
	now          func() time.Time
	newMigrator  func(sqlx.SqlConf) (migrator.Migrator, error)
}

type commandOption func(*commandOptions)

func withMigratorFactory(factory func(sqlx.SqlConf) (migrator.Migrator, error)) commandOption {
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
		newMigrator: func(conf sqlx.SqlConf) (migrator.Migrator, error) {
			return migrator.New(conf)
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
				return errors.New("up steps must be greater than 0")
			}

			m, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("steps") {
				return closeAfter(m.Up(uint(steps)), m)
			}
			return closeAfter(m.Up(), m)
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
				return errors.New("down steps must be greater than 0")
			}

			m, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(m.Down(uint(steps)), m)
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
			m, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(m.Goto(version), m)
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
			version, err := parseForceVersion(args[0])
			if err != nil {
				return err
			}
			m, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			return closeAfter(m.Force(version), m)
		},
	}
}

func loadMigrator(cmd *cobra.Command, loadSQLConf SQLConfigLoader, options *commandOptions) (migrator.Migrator, error) {
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

func closeAfter(primaryErr error, m migrator.Migrator) error {
	sourceErr, databaseErr := m.Close()
	return errors.Join(primaryErr, sourceErr, databaseErr)
}
