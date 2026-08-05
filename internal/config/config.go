package config

import (
	"time"
)

type (
	// Config is main structure holding configurations for different components.
	// All the parameters are parsed through a YAML file residing in the build path.
	Config struct {
		Daemon   Daemon            `yaml:"daemon"`
		Log      Log               `yaml:"log"`
		Storage  Storage           `yaml:"storage"`
		Clients  Clients           `yaml:"clients"`
		Tenants  map[uint64]Tenant `yaml:"tenants"`
		Services Services          `yaml:"services"`
	}

	// Daemon holds the GRPC and HTTP server settings.
	Daemon struct {
		GRPC struct {
			Host           string `yaml:"host" validate:"required"`
			Port           string `yaml:"port" validate:"required"`
			MaxRecvMsgSize int    `yaml:"maxRecvMsgSize" validate:"required"`
			MaxSendMsgSize int    `yaml:"maxSendMsgSize" validate:"required"`
		} `yaml:"grpc"`

		HTTP struct {
			Host               string      `yaml:"host" validate:"required"`
			Port               string      `yaml:"port" validate:"required"`
			HeaderClientIP     string      `yaml:"headerClientIp"`
			Headers            HTTPHeaders `yaml:"headers"`
			MaxCallRecvMsgSize int         `yaml:"maxCallRecvMsgSize" validate:"required"`
			MaxCallSendMsgSize int         `yaml:"maxCallSendMsgSize" validate:"required"`
		} `yaml:"http"`

		JWT struct {
			Secret         Sensitive     `yaml:"secret" validate:"required" init:"random"`
			ExpirationTime time.Duration `yaml:"expirationTime" validate:"required"`
			MaxRefreshTime time.Duration `yaml:"maxRefreshTime" validate:"required"`
		} `yaml:"jwt"`

		TOTP struct {
			NumRecoveryCodes int `yaml:"numRecoveryCodes" validate:"required"`
		} `yaml:"totp"`

		Jobs map[string]Job `yaml:"jobs" validate:"dive"`

		Jobber Jobber `yaml:"jobber"`

		ExposeErrorStackTrace bool `yaml:"exposeErrorStackTrace"`

		PrivateKeyFile string `yaml:"privateKeyFile" validate:"required_without=PrivateKey"`
		PrivateKey     string `yaml:"privateKey" validate:"required_without=PrivateKeyFile" init:"privatekey"`
		Salt           string `yaml:"salt" validate:"required" init:"random"`

		Metrics struct {
			Enabled        bool `yaml:"enabled"`
			Authentication struct {
				Enabled  bool      `yaml:"enabled"`
				Username string    `yaml:"username" validate:"required_if=Enabled true"`
				Password Sensitive `yaml:"password" validate:"required_if=Enabled true"`
			} `yaml:"authentication"`
		} `yaml:"metrics"`
	}

	// HTTPHeaders holds the daemon's HTTP response header settings.
	HTTPHeaders struct {
		AccessControlAllowOrigins        []string `yaml:"accessControlAllowOrigins"`
		AccessControlAllowOriginWildcard bool     `yaml:"accessControlAllowOriginWildcard"`
		AccessControlMaxAge              string   `yaml:"accessControlMaxAge" validate:"required"`
		CookieDomain                     string   `yaml:"cookieDomain" validate:"required"`
	}

	// Log bundles several logging instances.
	Log struct {
		Loggers map[string]Logger `yaml:"loggers" validate:"dive"`
	}

	// logger holds the settings for a go.uber.org/zap logging instance.
	Logger struct {
		Enabled bool `yaml:"enabled"`

		Type     string `yaml:"type" validate:"oneof=stdout file redis opensearch graylog"`
		Level    string `yaml:"level" validate:"oneof=debug info warn error"`
		Category string `yaml:"category" validate:"oneof=technical business security"`

		// File
		Path       string `yaml:"path" validate:"required_if=Type file"`
		MaxSize    int    `yaml:"maxSize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`

		// Redis
		Host     string    `yaml:"host" validate:"required_if=Type redis"`
		Port     string    `yaml:"port" validate:"required_if=Type redis"`
		Database int       `yaml:"database"`
		Password Sensitive `yaml:"password"`
		Key      string    `yaml:"key" validate:"required_if=Type redis"`

		// Graylog
		GraylogTimeout                        time.Duration `yaml:"graylogTimeout"`
		GraylogHost                           string        `yaml:"graylogHost" validate:"required_if=Type graylog"`
		GraylogBulkReceiving                  bool          `yaml:"graylogBulkReceiving"`
		GraylogAuthorizeSelfSignedCertificate bool          `yaml:"graylogAuthorizeSelfSignedCertificate"`

		// OpenSearch
		OpenSearchAddresses []string  `yaml:"osAddresses" validate:"required_if=Type opensearch"`
		OpenSearchUsername  string    `yaml:"osUsername"`
		OpenSearchPassword  Sensitive `yaml:"osPassword"`
		OpenSearchIndexName string    `yaml:"osIndexName" validate:"required_if=Type opensearch"`

		// for elasticsearch logger.
		BufferSize      int  `yaml:"bufferSize"`
		RateLimit       int  `yaml:"rateLimit"`
		DisallowDropLog bool `yaml:"disallowDropLog"`
	}

	Clients struct {
		K8sClient    K8sClient    `yaml:"kubernetes"`
		DockerClient DockerClient `yaml:"docker"`
		HarborClient HarborClient `yaml:"harbor"`
	}

	K8sClient struct {
		Enabled bool `yaml:"enabled"` // if true, the client will be used to connect to the k8s cluster

		KubeConfig string `yaml:"kubeConfig"` // either provide a path to a kubeconfig file

		APIServer                string    `yaml:"apiServer"`    // or a service account api server
		ServiceAccountSecretPath string    `yaml:"saSecretPath"` // and a service account secret path
		ServiceAccountOverrideCA string    `yaml:"saOverrideCa"` // optional CA crt content to override the one provided in the service account secret, useful for private clusters with custom CAs
		Token                    Sensitive `yaml:"token"`        // or a service account token
		CA                       string    `yaml:"ca"`           // and service account ca

		ImagePullSecretName string `yaml:"imagePullSecretName"`

		ServerVersion        string `yaml:"serverVersion" validate:"required_if=Enabled true"`
		InitContainerVersion string `yaml:"initContainerVersion" validate:"required_if=Enabled true"`
		AddUserDetails       bool   `yaml:"addUserDetails"`

		InsecureTLS bool `yaml:"insecureTls"` // if true, TLS certificate verification is skipped (testing only)

		IsWatcher bool `yaml:"isWatcher"` // if true, the client will watch for changes in the cluster

		PollInterval time.Duration `yaml:"pollInterval" validate:"required_if=Enabled true"` // how often to poll while waiting for a namespace/resource deletion to complete

		DefaultRegistry   string `yaml:"defaultRegistry" validate:"required_if=Enabled true,ne=CHANGEME" init:"placeholder"`
		DefaultRepository string `yaml:"defaultRepository" validate:"required_if=Enabled true"`

		PrepullNamespace     string `yaml:"prepullNamespace" validate:"required_if=Enabled true"` // namespace the backend itself runs in, where pre-pull Jobs are created
		PrepullJobTTLSeconds int    `yaml:"prepullJobTtlSeconds" validate:"required_if=Enabled true"`
	}

	DockerClient struct {
		Enabled bool `yaml:"enabled"`
	}

	HarborClient struct {
		Enabled            bool      `yaml:"enabled"`
		URL                string    `yaml:"url" validate:"required_if=Enabled true,ne=CHANGEME" init:"placeholder"`
		Username           string    `yaml:"username"`
		Password           Sensitive `yaml:"password"`
		Project            string    `yaml:"project"`
		LabelPrefixes      []string  `yaml:"labelPrefixes"`
		PageSize           int       `yaml:"pageSize" validate:"required_if=Enabled true"`
		MaxParallelFetches uint64    `yaml:"maxParallelFetches" validate:"required_if=Enabled true"`
	}

	Tenant struct {
		Enabled     bool        `yaml:"enabled"`
		User        string      `yaml:"user"`
		Password    Sensitive   `yaml:"password"`
		IPWhitelist IPWhitelist `yaml:"ipWhitelist"`
		Mailing     struct {
			Sender struct {
				FromEmail string `yaml:"fromEmail"`
				FromName  string `yaml:"fromName"`
			} `yaml:"sender"`
			EmailAddresses map[string]string `yaml:"emailAddresses"`
		} `yaml:"mailing"`
		FileStorage TenantFileStorage `yaml:"fileStorage"`
	}

	TenantFileStorage struct {
		URL               string    `yaml:"url"`
		Region            string    `yaml:"region"`
		Bucket            string    `yaml:"bucket"`
		AccessKey         string    `yaml:"accessKey"`
		AccessSecret      Sensitive `yaml:"accessSecret"`
		EncryptionKey     Sensitive `yaml:"encryptionKey"`
		SizeLimitMB       uint64    `yaml:"sizeLimitMb"`
		PublicSizeLimitMB uint64    `yaml:"publicSizeLimitMb"`
		RateLimitMBps     uint64    `yaml:"rateLimitMbps"`
	}

	// IPWhitelist is a configuration to allow only a subset of IP addresses to
	// reach the HTTP endpoints.
	IPWhitelist struct {
		Enabled bool `yaml:"enabled"`
		// Subnetworks is the list of whitelisted CIDR ranges.
		Subnetworks []string `yaml:"subnetworks"`
	}

	Storage struct {
		Datastores map[string]Datastore `yaml:"datastores" validate:"dive"`
		FileStores map[string]FileStore `yaml:"fileStores" validate:"dive"`
	}

	Datastore struct {
		Type           string        `yaml:"type" validate:"oneof=postgres"`
		Host           string        `yaml:"host" validate:"required"`
		Port           string        `yaml:"port" validate:"required"`
		Username       string        `yaml:"username" validate:"required"`
		Password       Sensitive     `yaml:"password" validate:"required" init:"localdev=password"`
		Database       string        `yaml:"database" validate:"required"`
		MaxConnections int           `yaml:"maxConnections"`
		MaxLifetime    time.Duration `yaml:"maxLifetime"`
		SSL            struct {
			Enabled         bool   `yaml:"enabled"`
			CertificateFile string `yaml:"certificateFile" validate:"required_if=Enabled true"`
			KeyFile         string `yaml:"keyFile" validate:"required_if=Enabled true"`
		} `yaml:"ssl"`
	}

	FileStore struct {
		Type        string               `yaml:"type" validate:"oneof=minio disk"`
		MinioConfig FileStoreMinioConfig `yaml:"minioConfig"`
		DiskConfig  FileStoreDiskConfig  `yaml:"diskConfig"`
	}

	FileStoreMinioConfig struct {
		Enabled bool `yaml:"enabled"`

		Endpoint        string    `yaml:"endpoint" validate:"required_if=Enabled true"`
		AccessKeyID     string    `yaml:"accessKeyId" validate:"required_if=Enabled true"`
		SecretAccessKey Sensitive `yaml:"secretAccessKey" validate:"required_if=Enabled true" init:"localdev=minioadmin"`

		BucketName string `yaml:"bucketName" validate:"required_if=Enabled true"`
		UseSSL     bool   `yaml:"useSsl"`

		MultipartMinPartSize   uint64 `yaml:"multipartMinPartSize" validate:"required_if=Enabled true"`
		MultipartMaxPartSize   uint64 `yaml:"multipartMaxPartSize" validate:"required_if=Enabled true"`
		MultipartMaxTotalParts uint64 `yaml:"multipartMaxTotalParts" validate:"required_if=Enabled true"`
	}

	FileStoreDiskConfig struct {
		Enabled  bool   `yaml:"enabled"`
		BasePath string `yaml:"basePath" validate:"required_if=Enabled true"` // Base directory path for disk storage
	}

	Services struct {
		AuditService struct {
			Enabled       bool   `yaml:"enabled"`
			DatastoreName string `yaml:"datastoreName" validate:"required_if=Enabled true"`
		} `yaml:"auditService"`

		MailerService struct {
			SMTP struct {
				Enabled          bool      `yaml:"enabled"`
				User             string    `yaml:"user" validate:"required_unless=Authentication none"`
				Password         Sensitive `yaml:"password" validate:"required_unless=Authentication none"`
				Host             string    `yaml:"host" validate:"required_if=Enabled true"`
				Port             string    `yaml:"port" validate:"required_if=Enabled true"`
				Authentication   string    `yaml:"authentication" validate:"oneof=none plain login"`
				InsecureMode     bool      `yaml:"insecureMode"`
				CertificatesRepo string    `yaml:"certificatesRepo"`
				ServerName       string    `yaml:"serverName"`
			} `yaml:"smtp"`
		} `yaml:"mailerService"`

		AuthenticationService struct {
			Enabled       bool            `yaml:"enabled"`
			AuthUIEnabled bool            `yaml:"authUiEnabled"`
			Modes         map[string]Mode `yaml:"modes" validate:"min=1,dive"`
			SelfService   struct {
				TenantID uint64 `yaml:"tenantId" validate:"required"`
			} `yaml:"selfService"`
		} `yaml:"authenticationService"`

		OpenIDConnectProvider struct {
			Enabled                 bool                          `yaml:"enabled"`
			FrontendInteractionsURL string                        `yaml:"frontendInteractionsUrl" validate:"required_if=Enabled true"`
			JWKS                    Sensitive                     `yaml:"jwks" validate:"required_if=Enabled true" init:"jwks"`
			IssuerURL               string                        `yaml:"issuerUrl" validate:"required_if=Enabled true"`
			Scopes                  []string                      `yaml:"scopes"`
			Clients                 []OpenIDConnectProviderClient `yaml:"clients" validate:"required_if=Enabled true,dive"`
		} `yaml:"openidConnectProvider"`

		WorkbenchService struct {
			StreamProxyEnabled         bool          `yaml:"streamProxyEnabled"`
			BackendInK8S               bool          `yaml:"backendInK8s"`
			ProxyHitSaveBatchInterval  time.Duration `yaml:"proxyHitSaveBatchInterval" validate:"required"`
			WorkbenchIdleTimeout       time.Duration `yaml:"workbenchIdleTimeout"`
			WorkbenchIdleCheckInterval time.Duration `yaml:"workbenchIdleCheckInterval" validate:"required"`
			RoundTripper               struct {
				DialTimeout           time.Duration `yaml:"dialTimeout"`
				DialKeepAlive         time.Duration `yaml:"dialKeepAlive"`
				ForceAttemptHTTP2     bool          `yaml:"forceAttemptHttp2"`
				MaxIdleConns          int           `yaml:"maxIdleConns"`
				MaxIdleConnsPerHost   int           `yaml:"maxIdleConnsPerHost"`
				IdleConnTimeout       time.Duration `yaml:"idleConnTimeout"`
				TLSHandshakeTimeout   time.Duration `yaml:"tlsHandshakeTimeout"`
				ResponseHeaderTimeout time.Duration `yaml:"responseHeaderTimeout"`
				MaxTransientRetry     int           `yaml:"maxTransientRetry" validate:"required"`
			} `yaml:"roundTripper"`
		} `yaml:"workbenchService"`

		WorkspaceService struct {
			EnableKillFixedTimeout bool          `yaml:"enableKillFixedTimeout"`
			KillFixedTimeout       time.Duration `yaml:"killFixedTimeout" validate:"required_if=EnableKillFixedTimeout true"`
			KillFixedCheckInterval time.Duration `yaml:"killFixedCheckInterval" validate:"required_if=EnableKillFixedTimeout true"`
			CreatorIsAdmin         bool          `yaml:"creatorIsAdmin"`
			CreatorIsDataManager   bool          `yaml:"creatorIsDataManager"`
			GIDOffset              uint64        `yaml:"gidOffset"`
		} `yaml:"workspaceService"`

		WorkspaceFileService struct {
			Stores map[string]WorkspaceFileStore `yaml:"stores" validate:"dive"`
		} `yaml:"workspaceFileService"`

		AuthorizationService struct {
			WorkspaceAdminCanAssignDataManager bool `yaml:"workspaceAdminCanAssignDataManager"`
		} `yaml:"authorizationService"`

		ApprovalRequestService struct {
			StagingFileStoreName       string `yaml:"stagingFileStoreName" validate:"required"`
			RequireDataManagerApproval bool   `yaml:"requireDataManagerApproval"`
		} `yaml:"approvalRequestService"`

		UserService struct {
			RequireEmail bool   `yaml:"requireEmail"`
			UIDOffset    uint64 `yaml:"uidOffset"`
		} `yaml:"userService"`

		Steward struct {
			Tenant struct {
				Name string `yaml:"name" validate:"required"`
			} `yaml:"tenant"`

			User struct {
				Username string    `yaml:"username" validate:"required"`
				Password Sensitive `yaml:"password" validate:"required" init:"localdev=password"`
			} `yaml:"user"`
		} `yaml:"steward"`
	}

	WorkspaceFileStore struct {
		WorkspacePrefix string `yaml:"workspacePrefix" validate:"required"`
		Description     string `yaml:"description"`
		Order           uint   `yaml:"order"`
	}

	Mode struct {
		Type                      string `yaml:"type" validate:"oneof=internal openid"`
		Enabled                   bool   `yaml:"enabled"`
		MainSource                bool   `yaml:"mainSource"`
		PublicRegistrationEnabled bool   `yaml:"publicRegistrationEnabled"`
		OpenID                    OpenID `yaml:"openid"`
		ButtonText                string `yaml:"buttonText"`
		Order                     uint   `yaml:"order"`
	}

	OpenID struct {
		ID                        string    `yaml:"id" validate:"ne=internal"`
		ChorusBackendHost         string    `yaml:"chorusBackendHost"`
		EnableFrontendRedirect    bool      `yaml:"enableFrontendRedirect"`
		ChorusFrontendRedirectURL string    `yaml:"chorusFrontendRedirectUrl"`
		AuthorizeURL              string    `yaml:"authorizeUrl"`
		TokenURL                  string    `yaml:"tokenUrl"`
		UserInfoURL               string    `yaml:"userInfoUrl"`
		FinalURLFormat            string    `yaml:"finalUrlFormat"`
		LogoutURL                 string    `yaml:"logoutUrl"`
		UserNameClaim             string    `yaml:"userNameClaim"`
		EmailClaim                string    `yaml:"emailClaim"`
		ClientID                  string    `yaml:"clientId"`
		ClientSecret              Sensitive `yaml:"clientSecret" validate:"ne=CHANGEME" init:"placeholder"`
		Scopes                    []string  `yaml:"scopes"`
		InsecureSkipTLS           bool      `yaml:"insecureSkipTls"`
		CustomCA                  string    `yaml:"customCa"`
	}

	OpenIDConnectProviderClient struct {
		ID string `yaml:"clientId" validate:"required"`
		// Secret is used when the client authenticates with client_secret_jwt,
		// since the key used to sign the assertion is the same used to verify it.
		Secret Sensitive `yaml:"clientSecret"`
		// HashedSecret is the hash of the client secret for the client_secret_basic
		// and client_secret_post authentication methods.
		HashedSecret Sensitive `yaml:"hashedSecret"`
		// RegistrationToken is the plain text registration access token generated during
		// dynamic client registration.
		// RegistrationToken  string `yaml:"registrationToken"`
		CreatedAtTimestamp int `yaml:"createdAt"`
		ExpiresAtTimestamp int `yaml:"expiresAt"`

		OnlyPreLoggedForClient bool `yaml:"onlyPreLoggedForClient"`

		GrantAutoApproved bool           `yaml:"grantAutoApproved"`
		GrantDuration     *time.Duration `yaml:"grantDuration"`

		// UserDelegation enables a special client_credentials mode where the
		// access token is issued on behalf of a specific user instead of the
		// client itself. When enabled:
		//   - The token "sub" claim is set to the configured user ID.
		//   - Additional user claims (preferred_username, email, etc.) are
		//     embedded in the access token based on UserDelegationClaims.
		//
		// WARNING: This effectively grants the client full API access as the
		// configured user. Only enable this for trusted service accounts that
		// need to act as a specific user (e.g. CI/CD pipelines, automation
		// bots). Never expose the client secret of a user-delegation client.
		UserDelegation *UserDelegationConfig `yaml:"userDelegation"`

		IsFederated                bool     `yaml:"isFederated"`
		FederationRegistrationType string   `yaml:"federationRegistrationType"` // automatic or explicit
		FederationTrustMarkIDs     []string `yaml:"federationTrustMarkIds"`
		// OpenIDConnectProviderClientMeta `yaml:",inline"`

		Name              string   `yaml:"clientName"`
		SecretExpiresAt   *int     `yaml:"clientSecretExpiresAt"`
		ApplicationType   string   `yaml:"applicationType"` // web, native
		LogoURI           string   `yaml:"logoUri"`
		Contacts          []string `yaml:"contacts"`
		PolicyURI         string   `yaml:"policyUri"`
		TermsOfServiceURI string   `yaml:"tosUri"`
		RedirectURIs      []string `yaml:"redirectUris"`
		RequestURIs       []string `yaml:"requestUris"`
		GrantTypes        []string `yaml:"grantTypes"`    // client_credentials, authorization_code, refresh_token, implicit
		ResponseTypes     []string `yaml:"responseTypes"` // code, id_token, token, code id_token, code token, id_token token, code id_token token
		PublicJWKSURI     string   `yaml:"jwksUri"`
		// PublicJWKS        *JSONWebKeySet `yaml:"jwks"`
		// ScopeIDs contains the scopes available to the client separated by spaces.
		ScopeIDs string `yaml:"scope"`
		//...

		TokenAuthnMethod string `yaml:"tokenEndpointAuthMethod"` // none, client_secret_basic, client_secret_post, client_secret_jwt, private_key_jwt, tls_client_auth, self_signed_tls_client_auth, dpop
	}

	// UserDelegationConfig configures a client to act on behalf of a specific
	// user when using the client_credentials grant type.
	UserDelegationConfig struct {
		// UserID is the ID of the user the client acts on behalf of.
		UserID uint64 `yaml:"userId"`
		// TenantID is the tenant of the delegated user.
		TenantID uint64 `yaml:"tenantId"`
		// Claims lists extra claims to embed in the access token.
		// Supported values: preferred_username, email, name, given_name, family_name.
		Claims []string `yaml:"claims"`
	}

	Job struct {
		Enabled  bool                   `yaml:"enabled"`
		Timeout  time.Duration          `yaml:"timeout"`
		Interval time.Duration          `yaml:"interval" validate:"required_if=Enabled true"`
		Options  map[string]interface{} `yaml:"options"`
	}

	Jobber struct {
		Enabled       bool          `yaml:"enabled"`
		CheckInterval time.Duration `yaml:"checkInterval" validate:"required_if=Enabled true"`
		Jitter        float64       `yaml:"jitter"` // the actual interval is ±jitter * interval * uniform(0,1)
		LockStore     string        `yaml:"lockStore" validate:"required,oneof=memory postgres"`
	}
)
