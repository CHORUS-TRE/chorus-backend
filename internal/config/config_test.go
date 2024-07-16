package config

import (
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const testConfig = "./chorus-test.yml"

func TestDecodeDaemon(t *testing.T) {
	cfg := config(t)

	// Daemon
	require.Equal(t, "localhost", cfg.Daemon.HTTP.Host)
	require.Equal(t, "5000", cfg.Daemon.HTTP.Port)
	require.Equal(t, "127.0.0.1", cfg.Daemon.GRPC.Host)
	require.Equal(t, "5555", cfg.Daemon.GRPC.Port)
	require.Equal(t, Sensitive("jwt_secret"), cfg.Daemon.JWT.Secret)
	require.Equal(t, 10, cfg.Daemon.JWT.ExpirationTime)
	require.Equal(t, 10, cfg.Daemon.TOTP.NumRecoveryCodes)
	require.Equal(t, true, cfg.Daemon.PPROFEnabled)
	require.Equal(t, "True-Client-IP", cfg.Daemon.HTTP.HeaderClientIP)
}

func TestDecodeStorage(t *testing.T) {
	cfg := config(t)

	// Storage
	require.Equal(t, "This is a description", cfg.Storage.Description)

	s := cfg.Storage.Datastores["chorus"]
	require.Equal(t, "postgres", s.Type)
	require.Equal(t, "localhost", s.Host)
	require.Equal(t, "40657", s.Port)
	require.Equal(t, "root", s.Username)
	require.Equal(t, Sensitive("password"), s.Password)
	require.Equal(t, "chorus", s.Database)
	require.Equal(t, 10, s.MaxConnections)
	require.Equal(t, 10*time.Second, s.MaxLifetime)
	require.Equal(t, true, s.SSL.Enabled)
	require.Equal(t, "/chorus/postgres-certs/client.crt", s.SSL.CertificateFile)
	require.Equal(t, "/chorus/postgres-certs/client.key", s.SSL.KeyFile)
}

func TestDecodeLog(t *testing.T) {
	cfg := config(t)
	require.Equal(t, 6, len(cfg.Log.Loggers))

	// Loggers
	l := cfg.Log.Loggers["stdout_technical"]
	require.True(t, l.Enabled)
	require.Equal(t, "stdout", l.Type)
	require.Equal(t, "info", l.Level)
	require.Equal(t, "technical", l.Category)

	l = cfg.Log.Loggers["stdout_business"]
	require.True(t, l.Enabled)
	require.Equal(t, "stdout", l.Type)
	require.Equal(t, "info", l.Level)
	require.Equal(t, "business", l.Category)

	l = cfg.Log.Loggers["stdout_security"]
	require.True(t, l.Enabled)
	require.Equal(t, "stdout", l.Type)
	require.Equal(t, "warn", l.Level)
	require.Equal(t, "security", l.Category)

	l = cfg.Log.Loggers["file_technical"]
	require.True(t, l.Enabled)
	require.Equal(t, "file", l.Type)
	require.Equal(t, "error", l.Level)
	require.Equal(t, "technical", l.Category)
	require.Equal(t, "/var/log/chorus/technical.log", l.Path)
	require.Equal(t, 7, l.MaxAge)
	require.Equal(t, 20, l.MaxBackups)
	require.Equal(t, 50, l.MaxSize)

	l = cfg.Log.Loggers["redis_technical"]
	require.True(t, l.Enabled)
	require.Equal(t, "redis", l.Type)
	require.Equal(t, "info", l.Level)
	require.Equal(t, "technical", l.Category)
	require.Equal(t, "redis", l.Host)
	require.Equal(t, "6379", l.Port)
	require.Equal(t, 0, l.Database)
	require.Equal(t, Sensitive("redis_password"), l.Password)
	require.Equal(t, "log", l.Key)

	l = cfg.Log.Loggers["graylog_technical"]
	require.True(t, l.Enabled)
	require.Equal(t, "graylog", l.Type)
	require.Equal(t, "debug", l.Level)
	require.Equal(t, "technical", l.Category)
	require.Equal(t, "http://local.chorus-tre.ch:12201/gelf", l.GraylogHost)
	require.Equal(t, 5*time.Second, l.GraylogTimeout)
	require.True(t, l.GraylogBulkReceiving)
	require.True(t, l.GraylogAuthorizeSelfSignedCertificate)
}

func TestDecodeTenants(t *testing.T) {
	cfg := config(t)
	require.Len(t, cfg.Tenants, 1)

	tenant := cfg.Tenants[88888]

	// Ip Whitelist
	require.True(t, tenant.IPWhitelist.Enabled)
	require.Len(t, tenant.IPWhitelist.Subnetworks, 2)
	require.Equal(t, "127.0.0.1/32", tenant.IPWhitelist.Subnetworks[0])
	require.Equal(t, "10.1.0.0/16", tenant.IPWhitelist.Subnetworks[1])
}

func config(t *testing.T) Config {
	v := viper.New()
	v.SetConfigFile(testConfig)
	err := v.ReadInConfig()
	require.Nil(t, err)

	cfg := Config{}
	err = v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
	})
	require.Nil(t, err)

	return cfg
}
