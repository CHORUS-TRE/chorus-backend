package cmd

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var exportConfigCmd = &cobra.Command{
	Use:   "export-default-config",
	Short: "export configuration",
	Long:  `export the full configuration options, with the default values`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportDefaultConfig()
	},
}

func init() {
	rootCmd.AddCommand(exportConfigCmd)
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

// runExportConfig outputs the loaded configuration to stdout.
func runExportConfig() error {
	cfg := provider.ProvideConfig()

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
