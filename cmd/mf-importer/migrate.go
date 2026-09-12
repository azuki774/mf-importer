package main

import (
	"fmt"
	"mf-importer/internal/logger"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

func newMigrateCommand() *cobra.Command {
	command := &cobra.Command{Use: "migrate", Short: "apply embedded database migrations"}
	for _, operation := range []struct {
		name      string
		direction migrate.MigrationDirection
		limit     int
	}{
		{"up", migrate.Up, 0}, {"down", migrate.Down, 1},
	} {
		limit := operation.limit
		child := &cobra.Command{
			Use: operation.name, Short: operation.name + " database migrations", Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				if limit < 0 {
					return fmt.Errorf("--limit must be non-negative")
				}
				db, err := openImporterDatabase()
				if err != nil {
					return err
				}
				defer db.CloseDB()
				return applyMigrations(cmd.Context(), logger.NewLogger(), db, operation.direction, limit)
			},
		}
		child.Flags().IntVar(&limit, "limit", operation.limit, "maximum migrations to apply (0 means all)")
		command.AddCommand(child)
	}
	return command
}

func init() { rootCmd.AddCommand(newMigrateCommand()) }
