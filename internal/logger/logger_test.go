package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

func TestInitLoggers(t *testing.T) {
	cfg := config.Config{
		Log: config.Log{
			Loggers: map[string]config.Logger{
				"stdout_technical": {Enabled: true, Type: "stdout", Level: "info", Category: "technical"},
				"stdout_business":  {Enabled: true, Type: "stdout", Level: "info", Category: "business"},
				"stdout_security":  {Enabled: true, Type: "stdout", Level: "info", Category: "security"},
			},
		},
	}

	_, err := InitLoggers(cfg)

	assert.Nil(t, err)
}
