//go:build unit || integration || acceptance

package helpers

import (
	"fmt"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/spf13/viper"
)

var cfg config.Config

const TEST_CONFIG_FILE = "TEST_CONFIG_FILE"
const LOCAL_DEV_CONFIG_FILE = "./../../../configs/dev/chorus.yaml"

func TestConfigFile() string {
	if os.Getenv(TEST_CONFIG_FILE) != "" {
		return os.Getenv(TEST_CONFIG_FILE)
	}

	return LOCAL_DEV_CONFIG_FILE
}

func Setup() {
	configFile := TestConfigFile()

	viper.BindEnv("storage.datastores.chorus.host", "DB_HOST")

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("config file not found:", viper.GetString("config"))
		os.Exit(1)
	} else {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}

	cfg = provider.ProvideConfig()
	if _, err := logger.InitLoggers(cfg); err != nil {
		fmt.Println("unable to initialize loggers:", err.Error())
		os.Exit(1)
	}
}

func Conf() config.Config {
	return cfg
}
