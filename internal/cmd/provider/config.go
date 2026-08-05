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
	case "wildcard_with_allowlist":
		return fmt.Sprintf("FAIL '%s' must be false when 'accessControlAllowOrigins' is non-empty", path)
	default:
		return fmt.Sprintf("FAIL '%s' failed '%s' validation", path, fe.Tag())
	}
}

// siblingPath resolves a required_if/required_without Param (a Go field name
// within the same parent struct) to its full dotted yaml path, e.g. "Enabled"
// on Config.Clients.K8sClient.DefaultRegistry becomes "clients.kubernetes.enabled".
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
	v.SetDefault("daemon.grpc.maxRecvMsgSize", 1*GiB)
	v.SetDefault("daemon.grpc.maxSendMsgSize", 1*GiB)

	// Daemon - HTTP
	v.SetDefault("daemon.http.host", "0.0.0.0")
	v.SetDefault("daemon.http.port", "5000")
	v.SetDefault("daemon.http.headers.accessControlAllowOrigins", []string{"http://localhost:3000"})
	v.SetDefault("daemon.http.headers.accessControlAllowOriginWildcard", false)
	v.SetDefault("daemon.http.headers.accessControlMaxAge", "600")
	v.SetDefault("daemon.http.headers.cookieDomain", "localhost")
	v.SetDefault("daemon.http.maxCallRecvMsgSize", 1*GiB)
	v.SetDefault("daemon.http.maxCallSendMsgSize", 1*GiB)

	// Daemon - JWT
	v.SetDefault("daemon.jwt.secret", "")
	v.SetDefault("daemon.jwt.expirationTime", 3*24*time.Hour)
	v.SetDefault("daemon.jwt.maxRefreshTime", 180*24*time.Hour)

	// Daemon - TOTP
	v.SetDefault("daemon.totp.numRecoveryCodes", 10)

	// Daemon - Jobs
	v.SetDefault("daemon.jobs.app-sync.enabled", true)
	v.SetDefault("daemon.jobs.app-sync.interval", 30*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.timeout", 10*time.Minute)
	v.SetDefault("daemon.jobs.app-sync.options", map[string]interface{}{"tenant_id": 1, "user_id": 1})

	// Daemon - Jobber
	v.SetDefault("daemon.jobber.enabled", true)
	v.SetDefault("daemon.jobber.checkInterval", 30*time.Second)
	v.SetDefault("daemon.jobber.jitter", 0.2)
	v.SetDefault("daemon.jobber.lockStore", "postgres")

	// Daemon - Error Stack Trace
	v.SetDefault("daemon.exposeErrorStackTrace", false)

	// Daemon - Private Key
	v.SetDefault("daemon.privateKeyFile", "")
	v.SetDefault("daemon.privateKey", "")
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
	v.SetDefault("clients.kubernetes.enabled", true)
	v.SetDefault("clients.kubernetes.kubeConfig", "")
	v.SetDefault("clients.kubernetes.apiServer", "https://kubernetes.default.svc")
	v.SetDefault("clients.kubernetes.saSecretPath", "/var/run/secrets/kubernetes.io/serviceaccount")
	v.SetDefault("clients.kubernetes.saOverrideCa", "")
	v.SetDefault("clients.kubernetes.token", "")
	v.SetDefault("clients.kubernetes.ca", "")
	v.SetDefault("clients.kubernetes.imagePullSecretName", "regcred")
	v.SetDefault("clients.kubernetes.serverVersion", "6.3.6-r0-3")
	v.SetDefault("clients.kubernetes.initContainerVersion", "0.0.2-4")
	v.SetDefault("clients.kubernetes.addUserDetails", false)
	v.SetDefault("clients.kubernetes.insecureTls", false)
	v.SetDefault("clients.kubernetes.isWatcher", true)
	v.SetDefault("clients.kubernetes.pollInterval", 500*time.Millisecond)
	v.SetDefault("clients.kubernetes.defaultRegistry", "")
	v.SetDefault("clients.kubernetes.defaultRepository", "apps")
	v.SetDefault("clients.kubernetes.prepullNamespace", "backend")
	v.SetDefault("clients.kubernetes.prepullJobTtlSeconds", 60)
	// Default to the conventional kubeconfig path,
	// leave unset if file does not exist
	if home, err := os.UserHomeDir(); err == nil {
		kubeConfigPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeConfigPath); err == nil {
			v.SetDefault("clients.kubernetes.kubeConfig", kubeConfigPath)
		}
	}

	// Clients - Docker
	v.SetDefault("clients.docker.enabled", true)

	// Clients - Harbor
	v.SetDefault("clients.harbor.enabled", true)
	v.SetDefault("clients.harbor.url", "")
	v.SetDefault("clients.harbor.project", "apps")
	v.SetDefault("clients.harbor.labelPrefixes", []string{"ch.chorus-tre.", "org.opencontainers.image."})
	v.SetDefault("clients.harbor.pageSize", 100)
	v.SetDefault("clients.harbor.maxParallelFetches", 16)
	v.SetDefault("clients.harbor.username", "")
	v.SetDefault("clients.harbor.password", "")

	// Storage - Datastores
	v.SetDefault("storage.datastores.chorus.type", "postgres")
	v.SetDefault("storage.datastores.chorus.host", "127.0.0.1")
	v.SetDefault("storage.datastores.chorus.port", "5432")
	v.SetDefault("storage.datastores.chorus.username", "admin")
	v.SetDefault("storage.datastores.chorus.database", "chorus")
	v.SetDefault("storage.datastores.chorus.maxConnections", 10)
	v.SetDefault("storage.datastores.chorus.maxLifetime", 10*time.Second)
	v.SetDefault("storage.datastores.chorus.ssl.enabled", false)
	v.SetDefault("storage.datastores.chorus.password", "")

	v.SetDefault("storage.datastores.audit.type", "postgres")
	v.SetDefault("storage.datastores.audit.host", "127.0.0.1")
	v.SetDefault("storage.datastores.audit.port", "5432")
	v.SetDefault("storage.datastores.audit.username", "admin")
	v.SetDefault("storage.datastores.audit.database", "audit")
	v.SetDefault("storage.datastores.audit.maxConnections", 10)
	v.SetDefault("storage.datastores.audit.maxLifetime", 10*time.Second)
	v.SetDefault("storage.datastores.audit.ssl.enabled", false)
	v.SetDefault("storage.datastores.audit.password", "")

	// Storage - File Stores
	v.SetDefault("storage.fileStores.archive.type", "minio")
	v.SetDefault("storage.fileStores.archive.minioConfig.enabled", true)
	v.SetDefault("storage.fileStores.archive.minioConfig.endpoint", "localhost:9000")
	v.SetDefault("storage.fileStores.archive.minioConfig.accessKeyId", "minioadmin")
	v.SetDefault("storage.fileStores.archive.minioConfig.secretAccessKey", "")
	v.SetDefault("storage.fileStores.archive.minioConfig.bucketName", "chorus-data")
	v.SetDefault("storage.fileStores.archive.minioConfig.useSsl", false)
	v.SetDefault("storage.fileStores.archive.minioConfig.multipartMinPartSize", 5*MiB)
	v.SetDefault("storage.fileStores.archive.minioConfig.multipartMaxPartSize", 5*GiB)
	v.SetDefault("storage.fileStores.archive.minioConfig.multipartMaxTotalParts", 10000)

	v.SetDefault("storage.fileStores.disk.type", "disk")
	v.SetDefault("storage.fileStores.disk.diskConfig.enabled", true)
	v.SetDefault("storage.fileStores.disk.diskConfig.basePath", "docker/.diskfilestore")

	// Services - Audit
	v.SetDefault("services.auditService.enabled", true)
	v.SetDefault("services.auditService.datastoreName", "audit")

	// Services - Mailer
	v.SetDefault("services.mailerService.smtp.enabled", false)
	v.SetDefault("services.mailerService.smtp.host", "")
	v.SetDefault("services.mailerService.smtp.port", "")
	v.SetDefault("services.mailerService.smtp.user", "")
	v.SetDefault("services.mailerService.smtp.password", "")
	v.SetDefault("services.mailerService.smtp.authentication", "none")
	v.SetDefault("services.mailerService.smtp.insecureMode", false)
	v.SetDefault("services.mailerService.smtp.certificatesRepo", "")
	v.SetDefault("services.mailerService.smtp.serverName", "")

	// Services - Authentication
	v.SetDefault("services.authenticationService.enabled", true)
	v.SetDefault("services.authenticationService.authUiEnabled", true)
	v.SetDefault("services.authenticationService.selfService.tenantId", 1)

	v.SetDefault("services.authenticationService.modes.internal.type", "internal")
	v.SetDefault("services.authenticationService.modes.internal.enabled", true)
	v.SetDefault("services.authenticationService.modes.internal.mainSource", true)
	v.SetDefault("services.authenticationService.modes.internal.publicRegistrationEnabled", true)
	v.SetDefault("services.authenticationService.modes.internal.buttonText", "Login via local DB")
	v.SetDefault("services.authenticationService.modes.internal.order", 1)

	v.SetDefault("services.authenticationService.modes.keycloak.type", "openid")
	v.SetDefault("services.authenticationService.modes.keycloak.enabled", false)
	v.SetDefault("services.authenticationService.modes.keycloak.mainSource", false)
	v.SetDefault("services.authenticationService.modes.keycloak.buttonText", "Login with Keycloak")
	v.SetDefault("services.authenticationService.modes.keycloak.order", 2)
	v.SetDefault("services.authenticationService.modes.keycloak.openid.id", "keycloak")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.chorusBackendHost", "http://localhost:5000")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.enableFrontendRedirect", true)
	v.SetDefault("services.authenticationService.modes.keycloak.openid.chorusFrontendRedirectUrl", "http://localhost:3000/oauthredirect")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.authorizeUrl", "http://localhost:8080/realms/chorus/protocol/openid-connect/auth")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.tokenUrl", "http://localhost:8080/realms/chorus/protocol/openid-connect/token")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.userInfoUrl", "http://localhost:8080/realms/chorus/protocol/openid-connect/userinfo")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.logoutUrl", "http://localhost:8080/realms/chorus/protocol/openid-connect/logout?client_id=chorus&post_logout_redirect_uri=http://localhost:3000")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.userNameClaim", "preferred_username")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.clientId", "chorus")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.clientSecret", "")
	v.SetDefault("services.authenticationService.modes.keycloak.openid.scopes", []string{"openid", "profile", "email", "roles"})

	// Services - OpenID Connect Provider
	v.SetDefault("services.openidConnectProvider.enabled", false)
	v.SetDefault("services.openidConnectProvider.issuerUrl", "")               // e.g. http://localhost:5000/openid-connect
	v.SetDefault("services.openidConnectProvider.frontendInteractionsUrl", "") // e.g. http://localhost:5000/auth-ui
	v.SetDefault("services.openidConnectProvider.jwks", "")
	v.SetDefault("services.openidConnectProvider.scopes", []string{"openid", "profile", "email", "roles"})

	// Services - Workbench
	v.SetDefault("services.workbenchService.streamProxyEnabled", true)
	v.SetDefault("services.workbenchService.backendInK8s", false)
	v.SetDefault("services.workbenchService.proxyHitSaveBatchInterval", 30*time.Second)
	v.SetDefault("services.workbenchService.workbenchIdleTimeout", 24*time.Hour)
	v.SetDefault("services.workbenchService.workbenchIdleCheckInterval", 10*time.Second)
	v.SetDefault("services.workbenchService.roundTripper.dialTimeout", 5*time.Second)
	// If zero, keep-alive probes are sent with a default value (currently 15 seconds)
	// If negative, keep-alive probes are disabled.
	v.SetDefault("services.workbenchService.roundTripper.dialKeepAlive", 30*time.Second)
	v.SetDefault("services.workbenchService.roundTripper.forceAttemptHttp2", false)
	v.SetDefault("services.workbenchService.roundTripper.maxIdleConns", 256)
	v.SetDefault("services.workbenchService.roundTripper.maxIdleConnsPerHost", 256)
	v.SetDefault("services.workbenchService.roundTripper.idleConnTimeout", 90*time.Second)
	v.SetDefault("services.workbenchService.roundTripper.tlsHandshakeTimeout", 10*time.Second)
	v.SetDefault("services.workbenchService.roundTripper.responseHeaderTimeout", 15*time.Second)
	v.SetDefault("services.workbenchService.roundTripper.maxTransientRetry", 3)

	// Services - Workspace
	v.SetDefault("services.workspaceService.enableKillFixedTimeout", false)
	v.SetDefault("services.workspaceService.killFixedTimeout", 1*time.Hour)
	v.SetDefault("services.workspaceService.killFixedCheckInterval", 1*time.Hour)
	v.SetDefault("services.workspaceService.creatorIsAdmin", true)
	v.SetDefault("services.workspaceService.creatorIsDataManager", true)
	v.SetDefault("services.workspaceService.gidOffset", 0)

	// Services - Workspace File
	v.SetDefault("services.workspaceFileService.stores.archive.workspacePrefix", "workspaces/%s")
	v.SetDefault("services.workspaceFileService.stores.archive.description", "Object storage backed file store")
	v.SetDefault("services.workspaceFileService.stores.archive.order", 1)

	v.SetDefault("services.workspaceFileService.stores.disk.workspacePrefix", "workspaces/%s")
	v.SetDefault("services.workspaceFileService.stores.disk.description", "Local disk storage - intended for local development only")
	v.SetDefault("services.workspaceFileService.stores.disk.order", 2)

	// Services - Authorization
	v.SetDefault("services.authorizationService.workspaceAdminCanAssignDataManager", true)

	// Services - Approval Request
	v.SetDefault("services.approvalRequestService.stagingFileStoreName", "disk")
	v.SetDefault("services.approvalRequestService.requireDataManagerApproval", true)

	// Services - User
	v.SetDefault("services.userService.requireEmail", false)
	v.SetDefault("services.userService.uidOffset", 0)

	// Services - Steward
	v.SetDefault("services.steward.tenant.name", "default")
	v.SetDefault("services.steward.user.username", "chorus")
	v.SetDefault("services.steward.user.password", "")

}
