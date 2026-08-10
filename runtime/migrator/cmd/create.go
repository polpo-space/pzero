package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	migrationFilePattern = regexp.MustCompile(`^([0-9]+)_[a-z0-9_]+\.(up|down)\.sql$`)
	migrationNamePattern = regexp.MustCompile(`[\s_-]+`)
	validMigrationName   = regexp.MustCompile(`^[a-z0-9_]+$`)
)

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

func createMigrationFiles(dir, rawName string, now time.Time) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	name := normalizeMigrationName(rawName)
	if name == "" {
		return nil, errors.New("migration name cannot be empty")
	}
	if !validMigrationName.MatchString(name) {
		return nil, fmt.Errorf("migration name %q must contain only ASCII letters, digits, spaces, hyphens, or underscores", rawName)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	version := now.UTC().Format("20060102150405")
	versions := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			return nil, fmt.Errorf("migration file has unrecognized format: %s", entry.Name())
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
