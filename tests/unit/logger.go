package unit

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

func InitTestLogger() {
	//nolint:errcheck
	logger.InitLoggers(config.Config{
		Log: config.Log{
			Loggers: map[string]config.Logger{
				"stdout_technical": {Enabled: true, Type: "stdout", Level: "debug", Category: "technical"},
				"stdout_business":  {Enabled: true, Type: "stdout", Level: "debug", Category: "business"},
				"stdout_security":  {Enabled: true, Type: "stdout", Level: "debug", Category: "security"},
			},
		}})
}
