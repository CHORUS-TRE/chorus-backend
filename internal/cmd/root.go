package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

const (
	configDevPath     = "./configs/dev"
	configDevFilename = "chorus"
)

var configFilename = ""

var rootCmd = &cobra.Command{
	Use:   componentName,
	Short: descriptionShort,
	Long:  descriptionLong,
	Run:   startCmd.Run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = "v"
	rootCmd.SetVersionTemplate(getVersion())

	rootCmd.PersistentFlags().StringVar(
		&configFilename,
		"config",
		"",
		"config file (default is ./configs/dev/chorus.yml)",
	)
	rootCmd.PersistentFlags().StringVar(
		&component.RuntimeEnvironment,
		"runtime-environment",
		"",
		"the runtime environment, e.g. INT, ACC, PROD...",
	)
	err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	if err != nil {
		panic(err)
	}
	err = viper.BindPFlag("runtime-environment", rootCmd.PersistentFlags().Lookup("runtime-environment"))
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	if configFilename != "" {
		viper.SetConfigFile(configFilename)
	} else {
		viper.AddConfigPath(configDevPath)
		viper.SetConfigName(configDevFilename)
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintf(os.Stderr, "Error reading config file %v: %v", viper.ConfigFileUsed(), err)
		os.Exit(1)
	}
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
