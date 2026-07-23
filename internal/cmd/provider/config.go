package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"

	"github.com/go-viper/mapstructure/v2"
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

		if err := validateConfig(cfg); err != nil {
			fmt.Printf("config validation failed: %v\n", err)
			os.Exit(1)
		}
	})
	return cfg
}

// validateConfig fails fast on configuration that would otherwise boot the
// server with silently empty required fields (JWT secret, signing keys).
// Datastore credentials are intentionally not checked here: the primary
// "chorus" datastore is looked up by name (provider.ProvideMainDB ->
// ProvideDB("chorus")) rather than being its own named struct field, so it
// can't carry `validate` tags directly — that failure surfaces later, in
// ProvideDB, when the datastore is actually used.
func validateConfig(cfg config.Config) error {
	return ProvideValidator().Struct(cfg)
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
	v.SetDefault("daemon.http.host", "0.0.0.0")
	v.SetDefault("daemon.http.port", "5000")
	v.SetDefault("daemon.http.headers.access_control_allow_origins", []string{"http://localhost:3000"})
	v.SetDefault("daemon.http.headers.access_control_allow_origin_wildcard", true)
	v.SetDefault("daemon.http.headers.access_control_max_age", "600")
	v.SetDefault("daemon.http.headers.cookie_domain", "localhost")
	v.SetDefault("daemon.http.max_call_recv_msg_size", 1073741824) // 1 GiB
	v.SetDefault("daemon.http.max_call_send_msg_size", 1073741824) // 1 GiB
	v.SetDefault("daemon.grpc.host", "localhost")
	v.SetDefault("daemon.grpc.port", "5555")
	v.SetDefault("daemon.grpc.max_recv_msg_size", 1073741824) // 1 GiB
	v.SetDefault("daemon.grpc.max_send_msg_size", 1073741824) // 1 GiB
	v.SetDefault("daemon.jwt.expiration_time", "72h")
	v.SetDefault("daemon.jwt.max_refresh_time", "4320h") // 180 days
	v.SetDefault("daemon.totp.num_recovery_codes", 10)
	v.SetDefault("daemon.pprof_enabled", false)
	v.SetDefault("daemon.expose_error_stack_trace", true)
	v.SetDefault("daemon.metrics.enabled", true)
	v.SetDefault("daemon.metrics.authentication.enabled", false)

	// Jobber
	v.SetDefault("daemon.jobber.enabled", true)
	v.SetDefault("daemon.jobber.check_interval", 30*time.Second)
	v.SetDefault("daemon.jobber.jitter", 0.2)
	v.SetDefault("daemon.jobber.lock_store", "postgres")

	// Jobs
	v.SetDefault("daemon.jobs.app-sync.enabled", true)
	v.SetDefault("daemon.jobs.app-sync.interval", 30*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.timeout", 10*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.options", map[string]interface{}{"tenant_id": 1, "user_id": 1})

	// Storage
	v.SetDefault("storage.datastores.chorus.type", "postgres")
	v.SetDefault("storage.datastores.chorus.host", "127.0.0.1")
	v.SetDefault("storage.datastores.chorus.port", "5432")
	v.SetDefault("storage.datastores.chorus.username", "admin")
	v.SetDefault("storage.datastores.chorus.database", "chorus")
	v.SetDefault("storage.datastores.chorus.max_connections", 10)
	v.SetDefault("storage.datastores.chorus.max_lifetime", 10*time.Second)
	v.SetDefault("storage.datastores.chorus.ssl.enabled", false)
	v.SetDefault("storage.datastores.audit.type", "postgres")
	v.SetDefault("storage.datastores.audit.host", "127.0.0.1")
	v.SetDefault("storage.datastores.audit.port", "5432")
	v.SetDefault("storage.datastores.audit.username", "admin")
	v.SetDefault("storage.datastores.audit.database", "audit")
	v.SetDefault("storage.datastores.audit.max_connections", 10)
	v.SetDefault("storage.datastores.audit.max_lifetime", 10*time.Second)
	v.SetDefault("storage.datastores.audit.ssl.enabled", false)
	v.SetDefault("storage.file_stores.s3.type", "minio")
	v.SetDefault("storage.file_stores.s3.minio_config.enabled", true)
	v.SetDefault("storage.file_stores.s3.minio_config.endpoint", "localhost:9000")
	v.SetDefault("storage.file_stores.s3.minio_config.access_key_id", "minioadmin")
	v.SetDefault("storage.file_stores.s3.minio_config.bucket_name", "chorus-data")
	v.SetDefault("storage.file_stores.s3.minio_config.use_ssl", false)
	v.SetDefault("storage.file_stores.s3.minio_config.multipart_min_part_size", 5242880)    // 5MB
	v.SetDefault("storage.file_stores.s3.minio_config.multipart_max_part_size", 5368709120) // 5GB
	v.SetDefault("storage.file_stores.s3.minio_config.multipart_max_total_parts", 10000)
	v.SetDefault("storage.file_stores.disk.type", "disk")
	v.SetDefault("storage.file_stores.disk.disk_config.enabled", true)
	v.SetDefault("storage.file_stores.disk.disk_config.base_path", "docker/.diskfilestore")

	// Services
	v.SetDefault("services.audit_service.enabled", true)
	v.SetDefault("services.audit_service.datastore_name", "audit")
	v.SetDefault("services.authentication_service.enabled", true)
	v.SetDefault("services.authentication_service.auth_ui_enabled", true)
	v.SetDefault("services.authentication_service.self_service.tenant_id", 1)
	v.SetDefault("services.authentication_service.modes.internal.type", "internal")
	v.SetDefault("services.authentication_service.modes.internal.enabled", true)
	v.SetDefault("services.authentication_service.modes.internal.main_source", true)
	v.SetDefault("services.authentication_service.modes.internal.public_registration_enabled", true)
	v.SetDefault("services.authentication_service.modes.internal.button_text", "Login via local DB")
	v.SetDefault("services.authentication_service.modes.internal.order", 1)
	v.SetDefault("services.authentication_service.modes.keycloak.type", "openid")
	v.SetDefault("services.authentication_service.modes.keycloak.enabled", true)
	v.SetDefault("services.authentication_service.modes.keycloak.button_text", "Login with via Keycloak")
	v.SetDefault("services.authentication_service.modes.keycloak.order", 2)
	v.SetDefault("services.authentication_service.modes.keycloak.openid.id", "keycloak")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.chorus_backend_host", "http://localhost:5000")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.enable_frontend_redirect", true)
	v.SetDefault("services.authentication_service.modes.keycloak.openid.chorus_frontend_redirect_url", "http://localhost:3000/oauthredirect")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.authorize_url", "http://localhost:8080/realms/chorus/protocol/openid-connect/auth")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.token_url", "http://localhost:8080/realms/chorus/protocol/openid-connect/token")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.user_info_url", "http://localhost:8080/realms/chorus/protocol/openid-connect/userinfo")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.logout_url", "http://localhost:8080/realms/chorus/protocol/openid-connect/logout?client_id=chorus&post_logout_redirect_uri=http://localhost:3000")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.user_name_claim", "preferred_username")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.client_id", "chorus")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.scopes", []string{"openid", "profile", "email", "roles"})
	v.SetDefault("services.approval_request_service.staging_file_store_name", "disk")
	v.SetDefault("services.mailer_service.smtp.host", "smtp-relay.sendinblue.com")
	v.SetDefault("services.mailer_service.smtp.port", "587")
	v.SetDefault("services.mailer_service.smtp.user", "smtpUser")
	v.SetDefault("services.mailer_service.smtp.authentication", "none")
	v.SetDefault("services.mailer_service.smtp.insecure_mode", false)
	v.SetDefault("services.openid_connect_provider.enabled", true)
	v.SetDefault("services.openid_connect_provider.issuer_url", "http://localhost:5000/openid-connect")
	v.SetDefault("services.openid_connect_provider.frontend_interactions_url", "http://localhost:5000/auth-ui")
	v.SetDefault("services.openid_connect_provider.scopes", []string{"openid", "profile", "email", "roles"})
	v.SetDefault("services.workspace_service.enable_kill_fixed_timeout", false)
	v.SetDefault("services.workspace_service.kill_fixed_timeout", time.Hour)
	v.SetDefault("services.workspace_service.kill_fixed_check_interval", time.Hour)
	v.SetDefault("services.workspace_service.creator_is_admin", true)
	v.SetDefault("services.workspace_service.creator_is_data_manager", true)
	v.SetDefault("services.workspace_service.gid_offset", 2000)
	v.SetDefault("services.workspace_file_service.stores.archive.file_store_name", "s3")
	v.SetDefault("services.workspace_file_service.stores.archive.workspace_prefix", "workspaces/%s")
	v.SetDefault("services.workspace_file_service.stores.archive.description", "Long-term object storage (MinIO).")
	v.SetDefault("services.workspace_file_service.stores.archive.order", 1)
	v.SetDefault("services.workspace_file_service.stores.disk.file_store_name", "disk")
	v.SetDefault("services.workspace_file_service.stores.disk.workspace_prefix", "workspaces/%s")
	v.SetDefault("services.workspace_file_service.stores.disk.description", "Local disk storage.")
	v.SetDefault("services.workspace_file_service.stores.disk.order", 2)
	v.SetDefault("services.user_service.require_email", false)
	v.SetDefault("services.user_service.uid_offset", 2000)
	v.SetDefault("services.authorization_service.workspace_admin_can_assign_data_manager", true)
	v.SetDefault("services.approval_request_service.require_data_manager_approval", false)
	v.SetDefault("services.workbench_service.stream_proxy_enabled", true)
	v.SetDefault("services.workbench_service.backend_in_k8s", false)
	v.SetDefault("services.workbench_service.proxy_hit_save_batch_interval", 30*time.Second)
	v.SetDefault("services.workbench_service.workbench_idle_notification", nil)
	v.SetDefault("services.workbench_service.workbench_idle_timeout", 24*time.Hour)
	v.SetDefault("services.workbench_service.workbench_idle_check_interval", 10*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.dial_timeout", 5*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.dial_keep_alive", 30*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.force_attempt_http2", false)
	v.SetDefault("services.workbench_service.round_tripper.max_idle_conns", 256)
	v.SetDefault("services.workbench_service.round_tripper.max_idle_conns_per_host", 256)
	v.SetDefault("services.workbench_service.round_tripper.idle_conn_timeout", 90*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.tls_handshake_timeout", 10*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.response_header_timeout", 15*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.max_transient_retry", 3)
	v.SetDefault("services.steward.tenant.name", "default")
	v.SetDefault("services.steward.user.username", "chorus")

	// Clients
	v.SetDefault("clients.docker_client.enabled", true)
	v.SetDefault("clients.k8s_client.enabled", true)
	v.SetDefault("clients.k8s_client.insecure_tls", true)
	v.SetDefault("clients.k8s_client.is_watcher", true)
	v.SetDefault("clients.k8s_client.server_version", "6.3.6-r0-3")
	v.SetDefault("clients.k8s_client.init_container_version", "0.0.2-4")
	v.SetDefault("clients.k8s_client.add_user_details", false)
	v.SetDefault("clients.k8s_client.image_pull_secret_name", "regcred")
	v.SetDefault("clients.k8s_client.default_registry", "")
	v.SetDefault("clients.k8s_client.default_repository", "apps")
	// Default to the conventional kubeconfig path,
	// leave unset if file does not exist
	if home, err := os.UserHomeDir(); err == nil {
		kubeConfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeConfigPath); err == nil {
			v.SetDefault("clients.k8s_client.kube_config", kubeConfigPath)
		}
	}

	// Harbor
	v.SetDefault("clients.harbor_client.enabled", true)
	v.SetDefault("clients.harbor_client.url", "")
	v.SetDefault("clients.harbor_client.project", "apps")
	v.SetDefault("clients.harbor_client.label_prefixes", []string{"ch.chorus-tre.", "org.opencontainers.image."})
	v.SetDefault("clients.harbor_client.page_size", 100)
	v.SetDefault("clients.harbor_client.max_parallel_fetches", 16)
	v.SetDefault("clients.harbor_client.username", "")

	// Loggers
	v.SetDefault("log.loggers.stdout_technical.enabled", true)
	v.SetDefault("log.loggers.stdout_technical.type", "stdout")
	v.SetDefault("log.loggers.stdout_technical.level", "debug")
	v.SetDefault("log.loggers.stdout_technical.category", "technical")
	v.SetDefault("log.loggers.stdout_business.enabled", true)
	v.SetDefault("log.loggers.stdout_business.type", "stdout")
	v.SetDefault("log.loggers.stdout_business.level", "debug")
	v.SetDefault("log.loggers.stdout_business.category", "business")
	v.SetDefault("log.loggers.stdout_security.enabled", true)
	v.SetDefault("log.loggers.stdout_security.type", "stdout")
	v.SetDefault("log.loggers.stdout_security.level", "debug")
	v.SetDefault("log.loggers.stdout_security.category", "security")
}
