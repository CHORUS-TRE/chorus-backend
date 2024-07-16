package integration

import (
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

const (
	UserPassword   = "testpassword"
	UserEmail      = "tenant1@chorus-tre.ch"
	DefaultTimeout = 30 * time.Second
	TenantID       = 1
)

var cfg config.Config

func TestSetup() {
	logConf := config.Log{
		Loggers: map[string]config.Logger{
			"stdout_technical": {Enabled: true, Type: "stdout", Level: "debug", Category: "technical"},
			"stdout_business":  {Enabled: true, Type: "stdout", Level: "debug", Category: "business"},
			"stdout_security":  {Enabled: true, Type: "stdout", Level: "debug", Category: "security"},
		},
	}

	tenant := config.Tenant{
		Password: UserPassword,
		User:     UserEmail,
	}

	cfg = config.Config{
		Log: logConf, Tenants: map[uint64]config.Tenant{TenantID: tenant},
	}
	//nolint:errcheck
	logger.InitLoggers(cfg)
}

func Conf() config.Config {
	return cfg
}
