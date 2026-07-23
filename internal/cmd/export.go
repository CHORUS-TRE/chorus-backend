package cmd

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var exportDefaultConfigCmd = &cobra.Command{
	Use:   "export-default-config",
	Short: "export the code-level default configuration",
	Long:  `export the full configuration options, with the default values only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportDefaultConfig()
	},
}

var exportCurrentConfigCmd = &cobra.Command{
	Use:     "export-current-config",
	Short:   "export the currently resolved configuration",
	Long:    `export the configuration as it resolves from the given --config file(s) merged with the code-level defaults`,
	PreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportCurrentConfig()
	},
}

func init() {
	rootCmd.AddCommand(exportDefaultConfigCmd)
	rootCmd.AddCommand(exportCurrentConfigCmd)
}

// runExportDefaultConfig outputs the default configuration structure to stdout.
func runExportDefaultConfig() error {
	out, err := yaml.Marshal(provider.ProvideDefaultConfig())
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

// runExportCurrentConfig outputs the loaded configuration to stdout.
func runExportCurrentConfig() error {
	cfg := provider.ProvideConfig()

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
