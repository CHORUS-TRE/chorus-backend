package provider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"

	val "github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var configOnce sync.Once
var cfg config.Config

const (
	MiB = 1 << (10 * (iota + 2))
	GiB
)

// ProvideConfig returns the user-provided config structure. A field will
// take a default value, as given by the default config structure, if it is
// not specified by the user.
func ProvideConfig() config.Config {
	configOnce.Do(func() {
		resolved, err := resolveConfig()
		if err != nil {
			fmt.Printf("config error: unable to ProvideConfig: %v", err)
			os.Exit(1)
		}
		cfg = resolved

		if err := validateConfig(cfg); err != nil {
			fmt.Printf("config validation failed: %v\n", err)
			os.Exit(1)
		}
	})
	return cfg
}

// resolveConfig unmarshals viper's current settings into a Config, applying
// code-level defaults for anything left unset.
func resolveConfig() (config.Config, error) {
	var c config.Config
	decode := func(dc *mapstructure.DecoderConfig) { dc.TagName = "yaml" }

	if err := viper.GetViper().Unmarshal(&c, decode); err != nil {
		return c, err
	}

	SetDefaultConfig(viper.GetViper())

	if err := viper.GetViper().Unmarshal(&c, decode); err != nil {
		return c, err
	}

	return c, nil
}

// validateConfig fails fast on empty required fields (JWT secret, signing keys).
// Datastore credentials aren't struct fields (looked up by name via ProvideDB),
// so they can't carry `validate` tags and fail later instead.
func validateConfig(cfg config.Config) error {
	return ProvideValidator().Struct(cfg)
}

// CheckConfig resolves the configuration and reports its validation errors,
// one per line, in a human-readable form -- unlike ProvideConfig, it never
// exits the process.
func CheckConfig() error {
	cfg, err := resolveConfig()
	if err != nil {
		return err
	}

	err = validateConfig(cfg)
	if err == nil {
		return nil
	}

	var validationErrs val.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	lines := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		lines = append(lines, formatValidationError(cfg, fe))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func formatValidationError(cfg config.Config, fe val.FieldError) string {
	// Namespace() is rooted at the Config struct's Go type name (e.g.
	// "Config.daemon.grpc.host"); strip it to get the plain dotted path
	// used everywhere else (--set, CHORUS_* env vars, config.yaml).
	_, path, _ := strings.Cut(fe.Namespace(), ".")

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("FAIL '%s' is missing", path)
	case "required_without":
		sibling := siblingPath(cfg, fe, fe.Param())
		return fmt.Sprintf("FAIL '%s' is missing (required unless '%s' is set)", path, sibling)
	case "required_if":
		field, value, _ := strings.Cut(fe.Param(), " ")
		sibling := siblingPath(cfg, fe, field)
		return fmt.Sprintf("FAIL '%s' is missing (required when '%s' is %s)", path, sibling, value)
	case "required_unless":
		field, value, _ := strings.Cut(fe.Param(), " ")
		sibling := siblingPath(cfg, fe, field)
		return fmt.Sprintf("FAIL '%s' is missing (required unless '%s' is %s)", path, sibling, value)
	case "oneof":
		return fmt.Sprintf("FAIL '%s' must be one of: %s", path, fe.Param())
	case "ne":
		return fmt.Sprintf("FAIL '%s' must not be '%s'", path, fe.Param())
	case "min":
		return fmt.Sprintf("FAIL '%s' must contain at least %s entry/entries", path, fe.Param())
	case "required_openid":
		return fmt.Sprintf("FAIL '%s' is missing (required for openid-type authentication modes)", path)
	default:
		return fmt.Sprintf("FAIL '%s' failed '%s' validation", path, fe.Tag())
	}
}

// siblingPath resolves a required_if/required_without Param (a Go field name
// within the same parent struct) to its full dotted yaml path, e.g. "Enabled"
// on Config.Clients.K8sClient.DefaultRegistry becomes "clients.k8s_client.enabled".
// Parent segments may be map-indexed (e.g. "Jobs[app-sync]") when the field
// lives inside a map, as with Daemon.Jobs. Falls back to the raw Go field name
// if anything doesn't resolve cleanly.
func siblingPath(cfg config.Config, fe val.FieldError, goFieldName string) string {
	structSegments := strings.Split(fe.StructNamespace(), ".")
	if len(structSegments) < 2 {
		return goFieldName
	}
	parentSegments := structSegments[1 : len(structSegments)-1] // drop leading "Config" and the leaf field

	v := reflect.ValueOf(cfg)
	for _, seg := range parentSegments {
		fieldName, mapKey, hasMapKey := strings.Cut(seg, "[")
		v = v.FieldByName(fieldName)
		if !v.IsValid() {
			return goFieldName
		}
		if hasMapKey {
			key, ok := convertMapKey(v.Type().Key(), strings.TrimSuffix(mapKey, "]"))
			if !ok {
				return goFieldName
			}
			v = v.MapIndex(key)
			if !v.IsValid() {
				return goFieldName
			}
		}
	}

	sf, ok := v.Type().FieldByName(goFieldName)
	if !ok {
		return goFieldName
	}
	yamlName, _, _ := strings.Cut(sf.Tag.Get("yaml"), ",")
	if yamlName == "" || yamlName == "-" {
		return goFieldName
	}

	_, path, _ := strings.Cut(fe.Namespace(), ".")
	parentPath := path
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		parentPath = path[:idx]
	}
	return parentPath + "." + yamlName
}

// convertMapKey parses raw (as extracted from a "Field[key]" namespace
// segment) into a reflect.Value of the given map key type.
func convertMapKey(keyType reflect.Type, raw string) (reflect.Value, bool) {
	switch keyType.Kind() {
	case reflect.String:
		return reflect.ValueOf(raw).Convert(keyType), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(n).Convert(keyType), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(n).Convert(keyType), true
	default:
		return reflect.Value{}, false
	}
}

// FieldTag resolves a FieldError's own failing field (not a sibling, unlike
// siblingPath) and returns the value of the given struct tag on it, e.g.
// FieldTag(cfg, fe, "init") reads the field's `init:"..."` tag. Works
// identically for plain validate-tag failures and custom struct-level
// validator failures (e.g. validateOpenIDMode's sl.ReportError calls),
// since both produce the same StructNamespace() format.
func FieldTag(cfg config.Config, fe val.FieldError, tagName string) (string, bool) {
	segments := strings.Split(fe.StructNamespace(), ".")
	if len(segments) < 2 {
		return "", false
	}
	segments = segments[1:] // drop leading "Config"

	v := reflect.ValueOf(cfg)
	var sf reflect.StructField
	for _, seg := range segments {
		fieldName, mapKey, hasMapKey := strings.Cut(seg, "[")

		if v.Kind() != reflect.Struct {
			return "", false
		}
		var ok bool
		sf, ok = v.Type().FieldByName(fieldName)
		if !ok {
			return "", false
		}
		v = v.FieldByName(fieldName)
		if !v.IsValid() {
			return "", false
		}

		if hasMapKey {
			key, ok := convertMapKey(v.Type().Key(), strings.TrimSuffix(mapKey, "]"))
			if !ok {
				return "", false
			}
			v = v.MapIndex(key)
			if !v.IsValid() {
				return "", false
			}
		}
	}
	return sf.Tag.Get(tagName), true
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

	// Daemon - GRPC
	v.SetDefault("daemon.grpc.host", "localhost")
	v.SetDefault("daemon.grpc.port", "5555")
	v.SetDefault("daemon.grpc.max_recv_msg_size", 1*GiB)
	v.SetDefault("daemon.grpc.max_send_msg_size", 1*GiB)

	// Daemon - HTTP
	v.SetDefault("daemon.http.host", "0.0.0.0")
	v.SetDefault("daemon.http.port", "5000")
	v.SetDefault("daemon.http.headers.access_control_allow_origins", []string{"http://localhost:3000"})
	v.SetDefault("daemon.http.headers.access_control_allow_origin_wildcard", false)
	v.SetDefault("daemon.http.headers.access_control_max_age", "600")
	v.SetDefault("daemon.http.headers.cookie_domain", "localhost")
	v.SetDefault("daemon.http.max_call_recv_msg_size", 1*GiB)
	v.SetDefault("daemon.http.max_call_send_msg_size", 1*GiB)

	// Daemon - JWT
	v.SetDefault("daemon.jwt.secret", "")
	v.SetDefault("daemon.jwt.expiration_time", 3*24*time.Hour)
	v.SetDefault("daemon.jwt.max_refresh_time", 180*24*time.Hour)

	// Daemon - TOTP
	v.SetDefault("daemon.totp.num_recovery_codes", 10)

	// Daemon - Jobs
	v.SetDefault("daemon.jobs.app-sync.enabled", true)
	v.SetDefault("daemon.jobs.app-sync.interval", 30*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.timeout", 10*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.options", map[string]interface{}{"tenant_id": 1, "user_id": 1})

	// Daemon - Jobber
	v.SetDefault("daemon.jobber.enabled", true)
	v.SetDefault("daemon.jobber.check_interval", 30*time.Second)
	v.SetDefault("daemon.jobber.jitter", 0.2)
	v.SetDefault("daemon.jobber.lock_store", "postgres")

	// Daemon - Error Stack Trace
	v.SetDefault("daemon.expose_error_stack_trace", false)

	// Daemon - Private Key
	v.SetDefault("daemon.private_key_file", "")
	v.SetDefault("daemon.private_key", "")
	v.SetDefault("daemon.salt", "")

	// Daemon - Metrics
	v.SetDefault("daemon.metrics.enabled", true)
	v.SetDefault("daemon.metrics.authentication.enabled", false)
	v.SetDefault("daemon.metrics.authentication.username", "prometheus")
	v.SetDefault("daemon.metrics.authentication.password", "")

	// Log and Loggers
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

	// Clients - Kubernetes
	v.SetDefault("clients.k8s_client.enabled", true)
	v.SetDefault("clients.k8s_client.kube_config", "")
	v.SetDefault("clients.k8s_client.api_server", "https://kubernetes.default.svc")
	v.SetDefault("clients.k8s_client.sa_secret_path", "/var/run/secrets/kubernetes.io/serviceaccount")
	v.SetDefault("clients.k8s_client.sa_override_ca", "")
	v.SetDefault("clients.k8s_client.token", "")
	v.SetDefault("clients.k8s_client.ca", "")
	v.SetDefault("clients.k8s_client.image_pull_secret_name", "regcred")
	v.SetDefault("clients.k8s_client.server_version", "6.3.6-r0-3")
	v.SetDefault("clients.k8s_client.init_container_version", "0.0.2-4")
	v.SetDefault("clients.k8s_client.add_user_details", false)
	v.SetDefault("clients.k8s_client.insecure_tls", false)
	v.SetDefault("clients.k8s_client.is_watcher", true)
	v.SetDefault("clients.k8s_client.poll_interval", 500*time.Millisecond)
	v.SetDefault("clients.k8s_client.default_registry", "")
	v.SetDefault("clients.k8s_client.default_repository", "apps")
	v.SetDefault("clients.k8s_client.prepull_namespace", "backend")
	v.SetDefault("clients.k8s_client.prepull_job_ttl_seconds", 60)
	// Default to the conventional kubeconfig path,
	// leave unset if file does not exist
	if home, err := os.UserHomeDir(); err == nil {
		kubeConfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeConfigPath); err == nil {
			v.SetDefault("clients.k8s_client.kube_config", kubeConfigPath)
		}
	}

	// Clients - Docker
	v.SetDefault("clients.docker_client.enabled", true)

	// Clients - Harbor
	v.SetDefault("clients.harbor_client.enabled", true)
	v.SetDefault("clients.harbor_client.url", "")
	v.SetDefault("clients.harbor_client.project", "apps")
	v.SetDefault("clients.harbor_client.label_prefixes", []string{"ch.chorus-tre.", "org.opencontainers.image."})
	v.SetDefault("clients.harbor_client.page_size", 100)
	v.SetDefault("clients.harbor_client.max_parallel_fetches", 16)
	v.SetDefault("clients.harbor_client.username", "")
	v.SetDefault("clients.harbor_client.password", "")

	// Storage - Datastores
	v.SetDefault("storage.datastores.chorus.type", "postgres")
	v.SetDefault("storage.datastores.chorus.host", "127.0.0.1")
	v.SetDefault("storage.datastores.chorus.port", "5432")
	v.SetDefault("storage.datastores.chorus.username", "admin")
	v.SetDefault("storage.datastores.chorus.database", "chorus")
	v.SetDefault("storage.datastores.chorus.max_connections", 10)
	v.SetDefault("storage.datastores.chorus.max_lifetime", 10*time.Second)
	v.SetDefault("storage.datastores.chorus.ssl.enabled", false)
	v.SetDefault("storage.datastores.chorus.password", "")

	v.SetDefault("storage.datastores.audit.type", "postgres")
	v.SetDefault("storage.datastores.audit.host", "127.0.0.1")
	v.SetDefault("storage.datastores.audit.port", "5432")
	v.SetDefault("storage.datastores.audit.username", "admin")
	v.SetDefault("storage.datastores.audit.database", "audit")
	v.SetDefault("storage.datastores.audit.max_connections", 10)
	v.SetDefault("storage.datastores.audit.max_lifetime", 10*time.Second)
	v.SetDefault("storage.datastores.audit.ssl.enabled", false)
	v.SetDefault("storage.datastores.audit.password", "")

	// Storage - File Stores
	v.SetDefault("storage.file_stores.archive.type", "minio")
	v.SetDefault("storage.file_stores.archive.minio_config.enabled", true)
	v.SetDefault("storage.file_stores.archive.minio_config.endpoint", "localhost:9000")
	v.SetDefault("storage.file_stores.archive.minio_config.access_key_id", "minioadmin")
	v.SetDefault("storage.file_stores.archive.minio_config.secret_access_key", "")
	v.SetDefault("storage.file_stores.archive.minio_config.bucket_name", "chorus-data")
	v.SetDefault("storage.file_stores.archive.minio_config.use_ssl", false)
	v.SetDefault("storage.file_stores.archive.minio_config.multipart_min_part_size", 5*MiB)
	v.SetDefault("storage.file_stores.archive.minio_config.multipart_max_part_size", 5*GiB)
	v.SetDefault("storage.file_stores.archive.minio_config.multipart_max_total_parts", 10000)

	v.SetDefault("storage.file_stores.disk.type", "disk")
	v.SetDefault("storage.file_stores.disk.disk_config.enabled", true)
	v.SetDefault("storage.file_stores.disk.disk_config.base_path", "docker/.diskfilestore")

	// Services - Audit
	v.SetDefault("services.audit_service.enabled", true)
	v.SetDefault("services.audit_service.datastore_name", "audit")

	// Services - Mailer
	v.SetDefault("services.mailer_service.smtp.enabled", false)
	v.SetDefault("services.mailer_service.smtp.host", "")
	v.SetDefault("services.mailer_service.smtp.port", "")
	v.SetDefault("services.mailer_service.smtp.user", "")
	v.SetDefault("services.mailer_service.smtp.password", "")
	v.SetDefault("services.mailer_service.smtp.authentication", "none")
	v.SetDefault("services.mailer_service.smtp.insecure_mode", false)
	v.SetDefault("services.mailer_service.smtp.certificates_repo", "")
	v.SetDefault("services.mailer_service.smtp.server_name", "")

	// Services - Authentication
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
	v.SetDefault("services.authentication_service.modes.keycloak.enabled", false)
	v.SetDefault("services.authentication_service.modes.keycloak.main_source", false)
	v.SetDefault("services.authentication_service.modes.keycloak.button_text", "Login with Keycloak")
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
	v.SetDefault("services.authentication_service.modes.keycloak.openid.client_secret", "")
	v.SetDefault("services.authentication_service.modes.keycloak.openid.scopes", []string{"openid", "profile", "email", "roles"})

	// Services - OpenID Connect Provider
	v.SetDefault("services.openid_connect_provider.enabled", false)
	v.SetDefault("services.openid_connect_provider.issuer_url", "")                // e.g. http://localhost:5000/openid-connect
	v.SetDefault("services.openid_connect_provider.frontend_interactions_url", "") // e.g. http://localhost:5000/auth-ui
	v.SetDefault("services.openid_connect_provider.jwks", "")
	v.SetDefault("services.openid_connect_provider.scopes", []string{"openid", "profile", "email", "roles"})

	// Services - Workbench
	v.SetDefault("services.workbench_service.stream_proxy_enabled", true)
	v.SetDefault("services.workbench_service.backend_in_k8s", false)
	v.SetDefault("services.workbench_service.proxy_hit_save_batch_interval", 30*time.Second)
	v.SetDefault("services.workbench_service.workbench_idle_timeout", 24*time.Hour)
	v.SetDefault("services.workbench_service.workbench_idle_check_interval", 10*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.dial_timeout", 5*time.Second)
	// If zero, keep-alive probes are sent with a default value (currently 15 seconds)
	// If negative, keep-alive probes are disabled.
	v.SetDefault("services.workbench_service.round_tripper.dial_keep_alive", 30*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.force_attempt_http2", false)
	v.SetDefault("services.workbench_service.round_tripper.max_idle_conns", 256)
	v.SetDefault("services.workbench_service.round_tripper.max_idle_conns_per_host", 256)
	v.SetDefault("services.workbench_service.round_tripper.idle_conn_timeout", 90*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.tls_handshake_timeout", 10*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.response_header_timeout", 15*time.Second)
	v.SetDefault("services.workbench_service.round_tripper.max_transient_retry", 3)

	// Services - Workspace
	v.SetDefault("services.workspace_service.enable_kill_fixed_timeout", false)
	v.SetDefault("services.workspace_service.kill_fixed_timeout", 1*time.Hour)
	v.SetDefault("services.workspace_service.kill_fixed_check_interval", 1*time.Hour)
	v.SetDefault("services.workspace_service.creator_is_admin", true)
	v.SetDefault("services.workspace_service.creator_is_data_manager", true)
	v.SetDefault("services.workspace_service.gid_offset", 0)

	// Services - Workspace File
	v.SetDefault("services.workspace_file_service.stores.archive.workspace_prefix", "workspaces/%s")
	v.SetDefault("services.workspace_file_service.stores.archive.description", "Object storage backed file store")
	v.SetDefault("services.workspace_file_service.stores.archive.order", 1)

	v.SetDefault("services.workspace_file_service.stores.disk.workspace_prefix", "workspaces/%s")
	v.SetDefault("services.workspace_file_service.stores.disk.description", "Local disk storage - intended for local development only")
	v.SetDefault("services.workspace_file_service.stores.disk.order", 2)

	// Services - Authorization
	v.SetDefault("services.authorization_service.workspace_admin_can_assign_data_manager", true)

	// Services - Approval Request
	v.SetDefault("services.approval_request_service.staging_file_store_name", "disk")
	v.SetDefault("services.approval_request_service.require_data_manager_approval", true)

	// Services - User
	v.SetDefault("services.user_service.require_email", false)
	v.SetDefault("services.user_service.uid_offset", 0)

	// Services - Steward
	v.SetDefault("services.steward.tenant.name", "default")
	v.SetDefault("services.steward.user.username", "chorus")
	v.SetDefault("services.steward.user.password", "")

}
