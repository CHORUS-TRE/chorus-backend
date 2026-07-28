package cmd

import (
	"fmt"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
)

var checkConfigCmd = &cobra.Command{
	Use:     "check-config",
	Short:   "validate the currently resolved configuration",
	Long:    `validate the configuration as it resolves from the given --config file(s), --set overrides and CHORUS_* env vars, merged with the code-level defaults`,
	PreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckConfig()
	},
}

func init() {
	rootCmd.AddCommand(checkConfigCmd)
}

// runCheckConfig resolves the configuration and reports its validation
// errors, one per line.
func runCheckConfig() error {
	if err := provider.CheckConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("config is valid")
	return nil
}
