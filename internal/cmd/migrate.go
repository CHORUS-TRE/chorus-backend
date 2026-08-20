package cmd

import (
	"fmt"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate <db-name>",
	Short:   "run the embedded migrations for a datastore",
	Long:    `run the embedded migrations for the given datastore ID (e.g. "chorus" or "audit") to completion, then exit`,
	Args:    cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMigrate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(datastoreID string) error {
	if err := provider.Migrate(datastoreID); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("datastore %q migrated\n", datastoreID)
	return nil
}
