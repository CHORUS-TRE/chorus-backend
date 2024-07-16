package provider

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var configOnce sync.Once
var cfg config.Config

// ProvideConfig returns the user-provided config structure. A field will
// take a default value, as given by the default config structure, if it is
// not specified by the user.
func ProvideConfig() config.Config {
	configOnce.Do(func() {
		if err := viper.GetViper().Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" }); err != nil {
			fmt.Printf("config error: unable to ProvideConfig: %v", err)
			os.Exit(1)
		}

		SetDefaultConfig(viper.GetViper())

		if err := viper.GetViper().Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" }); err != nil {
			fmt.Printf("config error: unable to ProvideConfig: %v", err)
			os.Exit(1)
		}
	})
	return cfg
}

var defaultConfigOnce sync.Once
var defaultCfg config.Config

func ProvideDefaultConfig() config.Config {
	defaultConfigOnce.Do(func() {
		v := viper.New()

		if err := v.Unmarshal(&defaultCfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" }); err != nil {
			fmt.Printf("config error: unable to ProvideDefaultConfig: %v", err)
			os.Exit(1)
		}

		SetDefaultConfig(v)

		if err := v.Unmarshal(&defaultCfg, func(c *mapstructure.DecoderConfig) { c.TagName = "yaml" }); err != nil {
			fmt.Printf("config error: unable to ProvideDefaultConfig: %v", err)
			os.Exit(1)
		}
	})
	return defaultCfg
}

func SetDefaultConfig(v *viper.Viper) {
	// Daemon
	v.SetDefault("daemon.http.host", "localhost")
	v.SetDefault("daemon.http.port", "5000")
	v.SetDefault("daemon.http.headers.access_control_allow_origin", "*")
	v.SetDefault("daemon.http.max_call_recv_msg_size", 10485760)  // 10 MiB
	v.SetDefault("daemon.http.max_call_send_msg_size", 104857600) // 100 MiB
	v.SetDefault("daemon.grpc.host", "localhost")
	v.SetDefault("daemon.grpc.port", "5555")
	v.SetDefault("daemon.grpc.max_recv_msg_size", 10485760)  // 10 MiB
	v.SetDefault("daemon.grpc.max_send_msg_size", 104857600) // 100 MiB
	v.SetDefault("daemon.tenant_id", 9999999)
	v.SetDefault("daemon.jwt.expiration_time", 10*time.Minute)
	v.SetDefault("daemon.totp.num_recovery_codes", 10)
	v.SetDefault("daemon.pprof_enabled", false)
	v.SetDefault("daemon.jobs.job_status_gc.enabled", true)
	v.SetDefault("daemon.jobs.job_status_gc.timeout", 10*time.Minute)
	v.SetDefault("daemon.jobs.job_status_gc.interval", 15*time.Minute)
	v.SetDefault("daemon.jobs.job_status_gc.options.successes", 90)
	v.SetDefault("daemon.jobs.job_status_gc.options.failures", 90)

	// Storage
	v.SetDefault("storage.description", "Type can be 'postgres'")
	v.SetDefault("storage.datastores.chorus.type", "postgres")
	v.SetDefault("storage.datastores.chorus.host", "localhost")
	v.SetDefault("storage.datastores.chorus.port", "5432")
	v.SetDefault("storage.datastores.chorus.username", "admin")
	v.SetDefault("storage.datastores.chorus.database", "chorus")
	v.SetDefault("storage.datastores.chorus.max_connections", 0)
	v.SetDefault("storage.datastores.chorus.max_lifetime", 10*time.Second)
	v.SetDefault("storage.datastores.chorus.ssl.enabled", false)
	v.SetDefault("storage.datastores.chorus.ssl.certificate_file", "/chorus/postgres-certs/client.crt")
	v.SetDefault("storage.datastores.chorus.ssl.key_file", "/chorus/postgres-certs/client.key")

	// Services
	v.SetDefault("services.authentication_service.enabled", false)

	// Loggers
	v.SetDefault("log.description", "Type can be either 'stdout', 'file' or 'redis'. Level can be either 'debug', 'info', 'warn', or 'error'. Category can be either 'technical', 'business' or 'security'.")
	v.SetDefault("log.loggers.stdout_technical.enabled", true)
	v.SetDefault("log.loggers.stdout_technical.type", "stdout")
	v.SetDefault("log.loggers.stdout_technical.level", "info")
	v.SetDefault("log.loggers.stdout_technical.category", "technical")
	v.SetDefault("log.loggers.stdout_business.enabled", true)
	v.SetDefault("log.loggers.stdout_business.type", "stdout")
	v.SetDefault("log.loggers.stdout_business.level", "info")
	v.SetDefault("log.loggers.stdout_business.category", "business")
	v.SetDefault("log.loggers.stdout_security.enabled", true)
	v.SetDefault("log.loggers.stdout_security.type", "stdout")
	v.SetDefault("log.loggers.stdout_security.level", "info")
	v.SetDefault("log.loggers.stdout_security.category", "security")
}
