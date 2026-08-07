package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newVersionCommand(loadSQLConf SQLConfigLoader, options *commandOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the migration version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := loadMigrator(cmd, loadSQLConf, options)
			if err != nil {
				return err
			}
			version, dirty, versionErr := m.Version()
			if versionErr == nil {
				_, versionErr = fmt.Fprintf(cmd.OutOrStdout(), "Current version: %d, dirty: %t\n", version, dirty)
			}
			return closeAfter(versionErr, m)
		},
	}
}

func parseVersion(raw string) (uint, error) {
	version, err := strconv.ParseUint(raw, 10, strconv.IntSize)
	if err != nil {
		return 0, err
	}
	maxInt := uint64(^uint(0) >> 1)
	if version > maxInt {
		return 0, fmt.Errorf("migration version %s overflows int on %d-bit architecture", raw, strconv.IntSize)
	}
	return uint(version), nil
}

func parseForceVersion(raw string) (int, error) {
	version, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if version < 0 {
		return 0, fmt.Errorf("migration version %q: invalid syntax", raw)
	}
	return version, nil
}
