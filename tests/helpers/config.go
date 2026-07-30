//go:build unit || integration || acceptance

package helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/spf13/viper"
)

var cfg config.Config

const TEST_CONFIG_FILE = "TEST_CONFIG_FILE"
const TEST_CONFIG_SET = "TEST_CONFIG_SET"

func TestConfigFiles() []string {
	raw := os.Getenv(TEST_CONFIG_FILE)
	if raw == "" {
		fmt.Printf("%s must be set\n", TEST_CONFIG_FILE)
		os.Exit(1)
	}
	return strings.Split(raw, ",")
}

func Setup() {
	for i, configFile := range TestConfigFiles() {
		viper.SetConfigFile(configFile)

		var err error
		if i == 0 {
			err = viper.ReadInConfig()
		} else {
			err = viper.MergeInConfig()
		}
		if err != nil {
			fmt.Println("config file not found:", configFile)
			os.Exit(1)
		}
		fmt.Println("using config file:", configFile)
	}

	if raw := os.Getenv(TEST_CONFIG_SET); raw != "" {
		for _, kv := range strings.Split(raw, ",") {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				fmt.Printf("invalid %s entry %q: expected key=value\n", TEST_CONFIG_SET, kv)
				os.Exit(1)
			}
			viper.Set(k, v)
		}
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
