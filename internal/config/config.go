package config

import (
	"time"
)

type (
	// Config is main structure holding configurations for different components.
	// All the parameters are parsed through a YAML file residing in the build path.
	Config struct {
		Daemon    Daemon              `yaml:"daemon"`
		Log       Log                 `yaml:"log"`
		Storage   Storage             `yaml:"storage"`
		Clients   Clients             `yaml:"clients"`
		Tenants   map[uint64]Tenant   `yaml:"tenants"`
		Workflows map[string]Workflow `yaml:"workflows,omitempty"`
		Services  Services            `yaml:"services"`
	}

	// Daemon holds the GRPC and HTTP server settings.
	Daemon struct {
		GRPC struct {
			Host           string `yaml:"host"`
			Port           string `yaml:"port"`
			MaxRecvMsgSize int    `yaml:"max_recv_msg_size"`
			MaxSendMsgSize int    `yaml:"max_send_msg_size"`
		} `yaml:"grpc"`

		HTTP struct {
			Host           string `yaml:"host"`
			Port           string `yaml:"port"`
			HeaderClientIP string `yaml:"header_client_ip"`
			Headers        struct {
				AccessControlAllowOrigin string `yaml:"access_control_allow_origin"`
				AccessControlMaxAge      string `yaml:"access_control_max_age"`
			} `yaml:"headers"`
			MaxCallRecvMsgSize int `yaml:"max_call_recv_msg_size"`
			MaxCallSendMsgSize int `yaml:"max_call_send_msg_size"`
		} `yaml:"http"`

		JWT struct {
			Secret         Sensitive `yaml:"secret"`
			ExpirationTime int       `yaml:"expiration_time"`
		} `yaml:"jwt"`

		TOTP struct {
			NumRecoveryCodes int `yaml:"num_recovery_codes"`
		} `yaml:"totp"`

		Jobs map[string]Job `yaml:"jobs"`

		PPROFEnabled bool   `yaml:"pprof_enabled"`
		TenantID     uint64 `yaml:"tenant_id"`

		PrivateKeyFile string `yaml:"private_key_file"`
		PrivateKey     string `yaml:"private_key"`
		PublicKeyFile  string `yaml:"public_key_file"`
		Salt           string `yaml:"salt"`
	}

	// Log bundles several logging instances.
	Log struct {
		Loggers map[string]Logger `yaml:"loggers"`
	}

	// logger holds the settings for a go.uber.org/zap logging instance.
	Logger struct {
		Enabled bool `yaml:"enabled"`

		Type     string `yaml:"type"`
		Level    string `yaml:"level"`
		Category string `yaml:"category"`

		// File
		Path       string `yaml:"path,omitempty"`
		MaxSize    int    `yaml:"max_size,omitempty"`
		MaxBackups int    `yaml:"max_backups,omitempty"`
		MaxAge     int    `yaml:"max_age,omitempty"`

		// Redis
		Host     string    `yaml:"host,omitempty"`
		Port     string    `yaml:"port,omitempty"`
		Database int       `yaml:"database,omitempty"`
		Password Sensitive `yaml:"password,omitempty"`
		Key      string    `yaml:"key,omitempty"`

		// Graylog
		GraylogTimeout                        time.Duration `yaml:"graylogtimeout,omitempty"`
		GraylogHost                           string        `yaml:"grayloghost,omitempty"`
		GraylogBulkReceiving                  bool          `yaml:"graylogbulkreceiving,omitempty"`
		GraylogAuthorizeSelfSignedCertificate bool          `yaml:"graylogauthorizeselfsignedcertificate,omitempty"`

		// OpenSearch
		OpenSearchAddresses []string  `yaml:"osaddresses,omitempty"`
		OpenSearchUsername  string    `yaml:"osusername,omitempty"`
		OpenSearchPassword  Sensitive `yaml:"ospassword,omitempty"`
		OpenSearchIndexName string    `yaml:"osindexname,omitempty"`

		// for elasticsearch logger.
		BufferSize      int  `yaml:"buffersize,omitempty"`
		RateLimit       int  `yaml:"ratelimit,omitempty"`
		DisallowDropLog bool `yaml:"disallow_drop_log,omitempty"`
	}

	Workflow struct {
		Job                    Job           `yaml:"job,omitempty"`
		DefaultStepTryDuration time.Duration `yaml:"step_try_duration"`
	}

	Clients struct {
		HelmClient HelmClient `yaml:"helm_client,omitempty"`
	}

	HelmClient struct {
		KubeConfig string `yaml:"kube_config,omitempty"` // either provide a kubeconfig

		Token     string `yaml:"token,omitempty"`      // either a service account token
		CA        string `yaml:"ca,omitempty"`         // and service account ca
		APIServer string `yaml:"api_server,omitempty"` // and service account api server

		ImagePullSecrets []ImagePullSecret `yaml:"image_pull_secrets,omitempty"`
	}

	ImagePullSecret struct {
		Registry string `yaml:"registry,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	}

	Tenant struct {
		Enabled     bool        `yaml:"enabled"`
		User        string      `yaml:"user"`
		Password    Sensitive   `yaml:"password"`
		IPWhitelist IPWhitelist `yaml:"ip_whitelist"`
		Mailing     struct {
			Sender struct {
				FromEmail string `yaml:"from_email"`
				FromName  string `yaml:"from_name"`
			} `yaml:"sender"`
			EmailAddresses map[string]string `yaml:"email_addresses"`
		} `yaml:"mailing"`
		FileStorage TenantFileStorage `yaml:"file_storage"`
	}

	TenantFileStorage struct {
		URL               string    `yaml:"url"`
		Region            string    `yaml:"region"`
		Bucket            string    `yaml:"bucket"`
		AccessKey         string    `yaml:"access_key"`
		AccessSecret      Sensitive `yaml:"access_secret"`
		EncryptionKey     Sensitive `yaml:"encryption_key"`
		SizeLimitMB       uint64    `yaml:"size_limit_mb"`
		PublicSizeLimitMB uint64    `yaml:"public_size_limit_mb"`
		RateLimitMBps     uint64    `yaml:"rate_limit_mbps"`
	}

	PublicStorage struct {
		URL          string    `yaml:"url"`
		Bucket       string    `yaml:"bucket"`
		AccessKey    string    `yaml:"access_key"`
		AccessSecret Sensitive `yaml:"access_secret"`
	}

	// IPWhitelist is a configuration to allow only a subset of IP addresses to
	// reach the HTTP endpoints.
	IPWhitelist struct {
		Enabled bool `yaml:"enabled"`
		// Subnetworks is the list of whitelisted CIDR ranges.
		Subnetworks []string `yaml:"subnetworks"`
	}

	Storage struct {
		Description string               `yaml:"description,omitempty"`
		Datastores  map[string]Datastore `yaml:"datastores,omitempty"`
	}

	Datastore struct {
		// 'postgres'
		Type           string        `yaml:"type"`
		Host           string        `yaml:"host"`
		Instance       string        `yaml:"instance"` // When instance is set, the port is not used.
		Port           string        `yaml:"port"`
		Username       string        `yaml:"username"`
		Password       Sensitive     `yaml:"password"`
		Database       string        `yaml:"database"`
		MaxConnections int           `yaml:"max_connections"`
		MaxLifetime    time.Duration `yaml:"max_lifetime"`
		SSL            struct {
			Enabled         bool   `yaml:"enabled"`
			CertificateFile string `yaml:"certificate_file"`
			KeyFile         string `yaml:"key_file"`
		} `yaml:"ssl"`
	}

	Services struct {
		MailerService struct {
			SMTP struct {
				User             string    `yaml:"user"`
				Password         Sensitive `yaml:"password"`
				Host             string    `yaml:"host"`
				Port             string    `yaml:"port"`
				Authentication   string    `yaml:"authentication"`
				InsecureMode     bool      `yaml:"insecure_mode"`
				CertificatesRepo string    `yaml:"certificates_repo,omitempty"`
				ServerName       string    `yaml:"server_name,omitempty"`
			} `yaml:"smtp"`
		} `yaml:"mailer_service"`

		AuthenticationService struct {
			Enabled        bool `yaml:"enabled"`
			DevAuthEnabled bool `yaml:"dev_auth_enabled"`
		} `yaml:"authentication_service"`

		WorkbenchService struct {
			StreamProxyEnabled bool `yaml:"stream_proxy_enabled"`
			BackendInK8S       bool `yaml:"backend_in_k8s"`
		} `yaml:"workbench_service"`
	}

	Job struct {
		Enabled  bool                   `yaml:"enabled"`
		Timeout  time.Duration          `yaml:"timeout"`
		Interval time.Duration          `yaml:"interval"`
		Options  map[string]interface{} `yaml:"options"`
	}
)
