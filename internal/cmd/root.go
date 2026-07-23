package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/component"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	componentName    = "chorus"
	descriptionShort = "chorus is the backend for the chorus platform."
	descriptionLong  = `chorus is the backend for the chorus platform.`
)

// configFilenames holds every --config occurrence, in order.
// If empty, no file is loaded at all and the server runs
// on provider.SetDefaultConfig()'s code-level defaults alone.
var configFilenames = []string{}

var rootCmd = &cobra.Command{
	Use:     componentName,
	Short:   descriptionShort,
	Long:    descriptionLong,
	RunE:    startCmd.RunE,
	PreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
}

func init() {
	rootCmd.Version = "v"
	rootCmd.SetVersionTemplate(getVersion())

	rootCmd.PersistentFlags().StringArrayVar(
		&configFilenames,
		"config",
		[]string{},
		"config file path, repeatable (later files override earlier ones); omit entirely to run on code-level defaults only",
	)
	rootCmd.PersistentFlags().StringVar(
		&component.RuntimeEnvironment,
		"runtime-environment",
		"",
		"the runtime environment, e.g. INT, ACC, PROD...",
	)
	err := viper.BindPFlag("runtime-environment", rootCmd.PersistentFlags().Lookup("runtime-environment"))
	if err != nil {
		panic(err)
	}

	// Environment variables always apply on top of whatever files were (or
	// weren't) loaded, e.g. CHORUS_DAEMON_JWT_SECRET -> daemon.jwt.secret.
	// This is optional, not required: config files may still hold secret
	// values directly where that's simpler (local dev, CI).
	viper.SetEnvPrefix("CHORUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// initConfig loads every file in configFilenames, in order: the first is
// read as the base, the rest are merged on top.
func initConfig() error {
	if len(configFilenames) == 0 {
		fmt.Println("No --config passed, running on code-level defaults only")
		return nil
	}

	for i, f := range configFilenames {
		viper.SetConfigFile(f)

		var err error
		if i == 0 {
			err = viper.ReadInConfig()
		} else {
			err = viper.MergeInConfig()
		}
		if err != nil {
			return fmt.Errorf("unable to load config file %v: %w", f, err)
		}
		fmt.Println("Using config file:", f)
	}

	return nil
}

func getVersion() string {
	version, _ := json.Marshal(provider.ProvideComponentInfo())
	return string(version)
}

func Execute() {
	defer logPanicRecovery()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func logPanicRecovery() {
	if r := recover(); r != nil {
		logger.TechLog.Fatal(context.Background(), "goodbye world, panic occurred", zap.String("panic_error", fmt.Sprintf("%v", r)), zap.Stack("panic_stack_trace"))
	}
}
