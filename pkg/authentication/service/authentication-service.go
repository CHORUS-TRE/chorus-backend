package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"
	userModel "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	userService "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

// Authenticator defines the authentication service API.
type Authenticator interface {
	Authenticate(ctx context.Context, username, password, totp string) (string, time.Duration, error)
	RefreshToken(ctx context.Context) (string, time.Duration, error)
	GetAuthenticationModes() []model.AuthenticationMode
	AuthenticateOAuth(ctx context.Context, providerID string) (string, error)
	OAuthCallback(ctx context.Context, providerID, state, sessionState, code string) (string, time.Duration, string, error)
	Logout(ctx context.Context) (string, error)
}

type Userer interface {
	GetUser(ctx context.Context, req userService.GetUserReq) (*userModel.User, error)
	CreateUser(ctx context.Context, req userService.CreateUserReq) (*userModel.User, error)
	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*userModel.TotpRecoveryCode, error)
	DeleteTotpRecoveryCode(ctx context.Context, req *userService.DeleteTotpRecoveryCodeReq) error
}

// AuthenticationStore groups the functions for accessing the database.
type AuthenticationStore interface {
	GetActiveUser(ctx context.Context, username, source string) (*userModel.User, error)
}

// AuthenticationService is the authentication service handler.
type AuthenticationService struct {
	cfg                 config.Config
	userer              Userer
	signingKey          string        // signingKey is the secret key with which JWT-tokens are signed.
	jwtExpirationTime   time.Duration // jwtExpirationTime is the number of minutes until a JWT-token expires.
	maxRefreshTime      time.Duration
	daemonEncryptionKey *crypto.Secret
	store               AuthenticationStore // store is the database handler object.
	oauthConfigs        map[string]*oauth2.Config
}

// CustomClaims groups the JWT-token data fields.
type CustomClaims struct {
	ID        uint64   `json:"id"`
	TenantID  uint64   `json:"tenantId"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Username  string   `json:"username"`
	Source    string   `json:"source"`

	jwt.StandardClaims
}

// ErrUnauthorized is the error message for all validation failures to avoid being an oracle.
type ErrInvalidArgument struct{}

func (e *ErrInvalidArgument) Error() string {
	return "invalid argument"
}

type ErrUnauthorized struct{}

func (e *ErrUnauthorized) Error() string {
	return "invalid credentials"
}

type Err2FARequired struct{}

func (e *Err2FARequired) Error() string {
	return "2FA_REQUIRED"
}

// NewAuthenticationService returns a fresh authentication service instance.
func NewAuthenticationService(cfg config.Config, userer Userer, store AuthenticationStore, daemonEncryptionKey *crypto.Secret) (*AuthenticationService, error) {
	oauthConfigs := make(map[string]*oauth2.Config)

	// Initialize the OAuth2 configs for each OpenID mode
	hasMainSource := false
	for _, mode := range cfg.Services.AuthenticationService.Modes {
		if mode.MainSource {
			if hasMainSource {
				return nil, fmt.Errorf("only one authentication mode can be marked as main source")
			}
			hasMainSource = true
		}

		if mode.Type == "openid" {
			if mode.OpenID.ID == "internal" {
				return nil, fmt.Errorf("openid mode cannot be named internal")
			}

			redirectURL := mode.OpenID.ChorusBackendHost + "/api/rest/v1/authentication/oauth2/" + mode.OpenID.ID + "/redirect"

			if mode.OpenID.EnableFrontendRedirect {
				redirectURL = mode.OpenID.ChorusFrontendRedirectURL
			}

			oauthConfigs[mode.OpenID.ID] = &oauth2.Config{
				ClientID:     mode.OpenID.ClientID,
				ClientSecret: mode.OpenID.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  mode.OpenID.AuthorizeURL,
					TokenURL: mode.OpenID.TokenURL,
				},
				// RedirectURL: mode.OpenID.ChorusBackendHost + "/api/rest/v1",
				RedirectURL: redirectURL,
				Scopes:      mode.OpenID.Scopes,
			}
		}
	}

	return &AuthenticationService{
		cfg:                 cfg,
		userer:              userer,
		signingKey:          cfg.Daemon.JWT.Secret.PlainText(),
		jwtExpirationTime:   cfg.Daemon.JWT.ExpirationTime,
		maxRefreshTime:      cfg.Daemon.JWT.MaxRefreshTime,
		daemonEncryptionKey: daemonEncryptionKey,
		store:               store,
		oauthConfigs:        oauthConfigs,
	}, nil
}

func (a *AuthenticationService) GetAuthenticationModes() []model.AuthenticationMode {
	res := []model.AuthenticationMode{}
	for _, mode := range a.cfg.Services.AuthenticationService.Modes {
		if mode.Type == "internal" {
			res = append(res, model.AuthenticationMode{
				Type: mode.Type,
				Internal: model.Internal{
					PublicRegistrationEnabled: mode.PublicRegistrationEnabled,
				},
				ButtonText: mode.ButtonText,
				IconURL:    mode.IconURL,
				Order:      mode.Order,
			})
		}
		if mode.Type == "openid" {
			res = append(res, model.AuthenticationMode{
				Type: mode.Type,
				OpenID: model.OpenID{
					ID: mode.OpenID.ID,
				},
				ButtonText: mode.ButtonText,
				IconURL:    mode.IconURL,
				Order:      mode.Order,
			})
		}
	}

	// sort by order

	return res
}

// Authenticate verifies whether the user is activated and the provided password is
// correct. It then returns a fresh JWT token for further API access.
func (a *AuthenticationService) Authenticate(ctx context.Context, username, password, totp string) (string, time.Duration, error) {
	user, err := a.store.GetActiveUser(ctx, username, "internal")
	if err != nil {
		logger.SecLog.Info(ctx, "user not found", zap.String("username", username))
		return "", 0, &ErrUnauthorized{}
	}
	if user == nil {
		return "", 0, &ErrUnauthorized{}
	}

	if user.Source != "internal" {
		logger.SecLog.Info(ctx, "user from external source attempted internal password authentication", zap.String("username", username), zap.String("source", user.Source))
		return "", 0, &ErrUnauthorized{}
	}

	if !verifyPassword(user.Password, password) {
		logger.SecLog.Info(ctx, "user has entered an invalid password", zap.String("username", username))
		return "", 0, &ErrUnauthorized{}
	}

	if user.TotpEnabled && totp == "" {
		return "", 0, &Err2FARequired{}
	}

	if user.TotpEnabled && user.TotpSecret != nil {
		isTotpValid, err := crypto.VerifyTotp(totp, *user.TotpSecret, a.daemonEncryptionKey)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to verify totp", zap.Error(err))
			return "", 0, &ErrUnauthorized{}
		}
		if !isTotpValid {
			logger.SecLog.Info(ctx, "user has entered an invalid totp code", zap.String("username", username))
			// If TOTP challenge cannot be validated maybe it is a recovery code.
			codes, err := a.userer.GetTotpRecoveryCodes(ctx, user.TenantID, user.ID)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to retrieve TOTP recovery code", zap.Error(err), zap.String("username", username))
				return "", 0, &ErrUnauthorized{}
			}
			code, err := crypto.VerifyTotpRecoveryCode(ctx, totp, codes, a.daemonEncryptionKey)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to verify totp recovery code", zap.Error(err))
				return "", 0, &ErrUnauthorized{}
			}
			if code == nil {
				logger.SecLog.Info(ctx, "user has entered an invalid recovery code", zap.String("username", username))
				return "", 0, &ErrUnauthorized{}
			}

			if err := a.userer.DeleteTotpRecoveryCode(ctx, &userService.DeleteTotpRecoveryCodeReq{
				TenantID: user.TenantID,
				CodeID:   code.ID,
			}); err != nil {
				logger.TechLog.Error(ctx, "unable to delete used recovery code", zap.Error(err), zap.String("username", username), zap.Uint64("code", code.ID))
				return "", 0, &ErrUnauthorized{}
			}
		}
	}

	token, err := createJWTToken(a.signingKey, user, a.jwtExpirationTime, time.Now())
	if err != nil {
		logger.TechLog.Error(ctx, "unable to create JWT token", zap.Error(err))
		return "", 0, &ErrUnauthorized{}
	}
	return token, a.jwtExpirationTime, nil
}

func (a *AuthenticationService) AuthenticateOAuth(ctx context.Context, providerID string) (string, error) {
	oauthConfig, exists := a.oauthConfigs[providerID]
	if !exists {
		return "", fmt.Errorf("unable to find config for provider %s: %w", providerID, &ErrInvalidArgument{})
	}

	return oauthConfig.AuthCodeURL(uuid.Next()), nil
}

func (a *AuthenticationService) OAuthCallback(ctx context.Context, providerID, state, sessionState, code string) (string, time.Duration, string, error) {
	mode, err := a.getAuthMode(providerID)
	if err != nil {
		return "", 0, "", fmt.Errorf("unable to get mode: %w", err)
	}

	if mode.Type != "openid" {
		return "", 0, "", fmt.Errorf("invalid mode: %w", &ErrInvalidArgument{})
	}

	oauthConfig, exists := a.oauthConfigs[providerID]
	if !exists {
		return "", 0, "", fmt.Errorf("unable to find config for provider %s: %w", providerID, &ErrInvalidArgument{})
	}

	// Verify state (for CSRF protection) - usually, you'd compare this with a value stored in the user's session
	// if state != sessionState {
	// 	return nil, errors.New("invalid state parameter")
	// }

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to exchange token: %w", err)
	}

	client := oauthConfig.Client(ctx, token)

	userInfoResp, err := client.Get(mode.OpenID.UserInfoURL)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != http.StatusOK {
		return "", 0, "", fmt.Errorf("failed to get user info: received non-OK response: %d", userInfoResp.StatusCode)
	}

	// Decode user info response
	var userInfo map[string]string
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		return "", 0, "", fmt.Errorf("failed to decode user info response: %w", err)
	}

	// Get username claim field from config, default to "sub"
	usernameClaim := model.DEFAULT_USERNAME_CLAIM
	if mode.OpenID.UserNameClaim != "" {
		usernameClaim = mode.OpenID.UserNameClaim
	}

	username, ok := userInfo[usernameClaim]
	if !ok {
		return "", 0, "", fmt.Errorf("failed to find username claim %q in user info: %w", usernameClaim, &ErrInvalidArgument{})
	}

	user, err := a.store.GetActiveUser(ctx, username, providerID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", 0, "", nil
		}

		createUser := &userService.UserReq{
			FirstName:   userInfo["given_name"],
			LastName:    userInfo["family_name"],
			Username:    username,
			Source:      providerID,
			Password:    "",
			Status:      userModel.UserActive,
			Roles:       []userModel.UserRole{userModel.RoleAuthenticated},
			TotpEnabled: false,
		}

		_, err := a.userer.CreateUser(ctx, userService.CreateUserReq{TenantID: 1, User: createUser})
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to create user: %w", err)
		}

		user, err = a.store.GetActiveUser(ctx, username, providerID)
		if err != nil {
			return "", 0, "", fmt.Errorf("failed to create user: %w", err)
		}
	}

	jwtToken, err := createJWTToken(a.signingKey, user, a.jwtExpirationTime, time.Now())
	if err != nil {
		logger.TechLog.Error(ctx, "unable to create JWT token", zap.Error(err))
		return "", 0, "", &ErrUnauthorized{}
	}

	url := ""

	if mode.OpenID.FinalURLFormat != "" {
		url = fmt.Sprintf(mode.OpenID.FinalURLFormat, jwtToken)
	}

	return jwtToken, a.jwtExpirationTime, url, nil
}

func (a *AuthenticationService) RefreshToken(ctx context.Context) (string, time.Duration, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return "", 0, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return "", 0, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	user, err := a.userer.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       userID,
	})
	if err != nil {
		return "", 0, status.Error(codes.InvalidArgument, "could not get user")
	}

	issuedAt, err := jwt_model.ExtractIssuedAt(ctx)
	if err != nil {
		return "", 0, status.Error(codes.InvalidArgument, "could not extract issued at from jwt-token")
	}

	elapsed := time.Since(time.Unix(issuedAt, 0))
	if elapsed > a.maxRefreshTime {
		return "", 0, status.Error(codes.InvalidArgument, "too many refreshes please authenticate")
	}

	token, err := createJWTToken(a.signingKey, user, a.jwtExpirationTime, time.Unix(issuedAt, 0))
	if err != nil {
		return "", 0, status.Error(codes.Internal, "could not create jwt-token")
	}

	return token, a.jwtExpirationTime, nil
}

func (a *AuthenticationService) getAuthMode(id string) (*config.Mode, error) {
	for _, m := range a.cfg.Services.AuthenticationService.Modes {
		if m.Type == "internal" && id == "internal" {
			return &m, nil
		}
		if m.Type == "openid" && m.OpenID.ID == id {
			return &m, nil
		}
	}

	return nil, &ErrInvalidArgument{}
}

// verifyPassword checks whether the hashed password matches a provided hash.
func verifyPassword(hash, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}
	return true
}

// createJWTToken generates a fresh JWT token for a given user.
func createJWTToken(signingKey string, user *userModel.User, expirationTime time.Duration, issuedAt time.Time) (string, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = string(r)
	}
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}
	claims := CustomClaims{
		ID:        user.ID,
		TenantID:  user.TenantID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     roles,
		Username:  user.Username,
		Source:    user.Source,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationTime).Unix(),
			IssuedAt:  issuedAt.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}

func (a *AuthenticationService) Logout(ctx context.Context) (string, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return "", status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return "", status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	user, err := a.userer.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       userID,
	})
	if err != nil {
		return "", status.Error(codes.InvalidArgument, "could not get user")
	}

	mode, err := a.getAuthMode(user.Source)
	if err != nil {
		logger.SecLog.Error(ctx, "unknown user source, this should not happend, invalid config", zap.String("source", user.Source), zap.Uint64("user_id", user.ID), zap.Uint64("tenant_id", user.TenantID))
		return "", status.Error(codes.InvalidArgument, "could not get auth mode")
	}

	switch mode.Type {
	case "internal":
		return "", nil
	case "openid":
		return mode.OpenID.LogoutURL, nil
	default:
		logger.SecLog.Error(ctx, "unknown user source, this should never happend", zap.String("source", user.Source), zap.Uint64("user_id", user.ID), zap.Uint64("tenant_id", user.TenantID))
		return "", status.Error(codes.InvalidArgument, "unknown source")
	}
}
