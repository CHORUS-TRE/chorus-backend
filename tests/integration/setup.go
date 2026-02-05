package integration

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var cfg config.Config
var setupOnce sync.Once

func TestSetup() {
	setupOnce.Do(func() {
		logConf := config.Log{
			Loggers: map[string]config.Logger{
				"stdout_technical": {Enabled: true, Type: "stdout", Level: "debug", Category: "technical"},
				"stdout_business":  {Enabled: true, Type: "stdout", Level: "debug", Category: "business"},
				"stdout_security":  {Enabled: true, Type: "stdout", Level: "debug", Category: "security"},
			},
		}

		cfg = config.Config{
			Log: logConf,
		}
		//nolint:errcheck
		logger.InitLoggers(cfg)
	})
}
