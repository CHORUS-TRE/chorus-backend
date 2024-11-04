package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"
	userModel "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	userService "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/pkg/errors"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// Authenticator defines the authentication service API.
type Authenticator interface {
	Authenticate(ctx context.Context, username, password, totp string) (string, error)
	AuthenticateOAuth(ctx context.Context, providerID string) (string, error)
	OAuthCallback(ctx context.Context, providerID, state, sessionState, code string) (string, error)
}

type Userer interface {
	CreateUser(ctx context.Context, req userService.CreateUserReq) (uint64, error)
	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*userModel.TotpRecoveryCode, error)
	DeleteTotpRecoveryCode(ctx context.Context, req *userService.DeleteTotpRecoveryCodeReq) error
}

// AuthenticationStore groups the functions for accessing the database.
type AuthenticationStore interface {
	GetActiveUser(ctx context.Context, username, source string) (*model.User, error)
}

// AuthenticationService is the authentication service handler.
type AuthenticationService struct {
	cfg                 config.Config
	userer              Userer
	signingKey          string // signingKey is the secret key with which JWT-tokens are signed.
	jwtExpirationTime   int    // jwtExpirationTime is the number of minutes until a JWT-token expires.
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
func NewAuthenticationService(cfg config.Config, userer Userer, store AuthenticationStore, daemonEncryptionKey *crypto.Secret) *AuthenticationService {
	oauthConfigs := make(map[string]*oauth2.Config)

	// Initialize the OAuth2 configs for each OpenID mode
	for _, mode := range cfg.Services.AuthenticationService.Modes {
		if mode.Type == "openid" {
			if mode.OpenID.ID == "internal" {
				log.Fatal("openid mode cannot be named internal")
			}

			oauthConfigs[mode.OpenID.ID] = &oauth2.Config{
				ClientID:     mode.OpenID.ClientID,
				ClientSecret: mode.OpenID.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  mode.OpenID.AuthorizeURL,
					TokenURL: mode.OpenID.TokenURL,
				},
				// RedirectURL: mode.OpenID.ChorusBackendHost + "/api/rest/v1",
				RedirectURL: mode.OpenID.ChorusBackendHost + "/api/rest/v1/authentication/oauth2/" + mode.OpenID.ID + "/redirect",
				Scopes:      mode.OpenID.Scopes,
			}
		}
	}

	return &AuthenticationService{
		cfg:                 cfg,
		userer:              userer,
		signingKey:          cfg.Daemon.JWT.Secret.PlainText(),
		jwtExpirationTime:   cfg.Daemon.JWT.ExpirationTime,
		daemonEncryptionKey: daemonEncryptionKey,
		store:               store,
		oauthConfigs:        oauthConfigs,
	}
}

// Authenticate verifies whether the user is activated and the provided password is
// correct. It then returns a fresh JWT token for further API access.
func (a *AuthenticationService) Authenticate(ctx context.Context, username, password, totp string) (string, error) {
	user, err := a.store.GetActiveUser(ctx, username, "internal")
	if err != nil {
		logger.SecLog.Info(ctx, "user not found", zap.String("username", username))
		return "", &ErrUnauthorized{}
	}
	if user == nil {
		return "", &ErrUnauthorized{}
	}

	if user.Source != "internal" {
		logger.SecLog.Info(ctx, "user from external source attempted internal password authentication", zap.String("username", username), zap.String("source", user.Source))
		return "", &ErrUnauthorized{}
	}

	if !verifyPassword(user.Password, password) {
		logger.SecLog.Info(ctx, "user has entered an invalid password", zap.String("username", username))
		return "", &ErrUnauthorized{}
	}

	if user.TotpEnabled && totp == "" {
		return "", &Err2FARequired{}
	}

	if user.TotpEnabled && user.TotpSecret != nil {
		isTotpValid, err := crypto.VerifyTotp(totp, *user.TotpSecret, a.daemonEncryptionKey)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to verify totp", zap.Error(err))
			return "", &ErrUnauthorized{}
		}
		if !isTotpValid {
			logger.SecLog.Info(ctx, "user has entered an invalid totp code", zap.String("username", username))
			// If TOTP challenge cannot be validated maybe it is a recovery code.
			codes, err := a.userer.GetTotpRecoveryCodes(ctx, user.TenantID, user.ID)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to retrieve TOTP recovery code", zap.Error(err), zap.String("username", username))
				return "", &ErrUnauthorized{}
			}
			code, err := crypto.VerifyTotpRecoveryCode(ctx, totp, codes, a.daemonEncryptionKey)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to verify totp recovery code", zap.Error(err))
				return "", &ErrUnauthorized{}
			}
			if code == nil {
				logger.SecLog.Info(ctx, "user has entered an invalid recovery code", zap.String("username", username))
				return "", &ErrUnauthorized{}
			}

			if err := a.userer.DeleteTotpRecoveryCode(ctx, &userService.DeleteTotpRecoveryCodeReq{
				TenantID: user.TenantID,
				CodeID:   code.ID,
			}); err != nil {
				logger.TechLog.Error(ctx, "unable to delete used recovery code", zap.Error(err), zap.String("username", username), zap.Uint64("code", code.ID))
				return "", &ErrUnauthorized{}
			}
		}
	}

	token, err := createJWTToken(a.signingKey, user, a.jwtExpirationTime)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to create JWT token", zap.Error(err))
		return "", &ErrUnauthorized{}
	}
	return token, nil
}

func (a *AuthenticationService) AuthenticateOAuth(ctx context.Context, providerID string) (string, error) {
	oauthConfig, exists := a.oauthConfigs[providerID]
	if !exists {
		return "", errors.Wrap(&ErrInvalidArgument{}, "unable to find config for provider "+providerID)
	}

	return oauthConfig.AuthCodeURL(uuid.Next()), nil
}

func (a *AuthenticationService) OAuthCallback(ctx context.Context, providerID, state, sessionState, code string) (string, error) {
	oauthConfig, exists := a.oauthConfigs[providerID]
	if !exists {
		return "", errors.Wrap(&ErrInvalidArgument{}, "unable to find config for provider "+providerID)
	}

	// Verify state (for CSRF protection) - usually, you'd compare this with a value stored in the user's session
	// if state != sessionState {
	// 	return nil, errors.New("invalid state parameter")
	// }

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %v", err)
	}

	client := oauthConfig.Client(ctx, token)

	mode, err := a.getAuthMode(providerID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get mode")
	}

	userInfoResp, err := client.Get(mode.OpenID.UserInfoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %v", err)
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info: received non-OK response: %d", userInfoResp.StatusCode)
	}

	type OAuthUser struct {
		Username  string `json:"sub"`
		Email     string `json:"email"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
	}

	// var userInfo map[string]string
	var userInfo OAuthUser
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info response: %v", err)
	}

	user, err := a.store.GetActiveUser(ctx, userInfo.Username, providerID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}

		createUser := &userService.UserReq{
			FirstName:   userInfo.FirstName,
			LastName:    userInfo.LastName,
			Username:    userInfo.Username,
			Source:      providerID,
			Password:    "",
			Status:      userModel.UserActive,
			Roles:       []userModel.UserRole{userModel.RoleAuthenticated},
			TotpEnabled: false,
		}

		_, err := a.userer.CreateUser(ctx, userService.CreateUserReq{TenantID: 1, User: createUser})
		if err != nil {
			return "", fmt.Errorf("failed to create user: %v", err)
		}

		user, err = a.store.GetActiveUser(ctx, userInfo.Username, providerID)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %v", err)
		}
	}

	jwtToken, err := createJWTToken(a.signingKey, user, a.jwtExpirationTime)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to create JWT token", zap.Error(err))
		return "", &ErrUnauthorized{}
	}

	return jwtToken, nil
}

func (a *AuthenticationService) getAuthMode(id string) (*config.Mode, error) {
	for _, m := range a.cfg.Services.AuthenticationService.Modes {
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
func createJWTToken(signingKey string, user *model.User, expirationTime int) (string, error) {
	claims := CustomClaims{
		ID:        user.ID,
		TenantID:  user.TenantID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     user.Roles,
		Username:  user.Username,
		Source:    user.Source,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(expirationTime)).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}
