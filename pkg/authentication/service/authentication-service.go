package service

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Authenticator defines the authentication service API.
type Authenticator interface {
	Authenticate(ctx context.Context, username, password string, totp string) (string, error)
}

type Userer interface {
	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*user_model.TotpRecoveryCode, error)
	DeleteTotpRecoveryCode(ctx context.Context, req *service.DeleteTotpRecoveryCodeReq) error
}

// AuthenticationStore groups the functions for accessing the database.
type AuthenticationStore interface {
	GetActiveUser(ctx context.Context, username string) (*model.User, error)
}

// AuthenticationService is the authentication service handler.
type AuthenticationService struct {
	userer              Userer
	signingKey          string // signingKey is the secret key with which JWT-tokens are signed.
	jwtExpirationTime   int    // jwtExpirationTime is the number of minutes until a JWT-token expires.
	daemonEncryptionKey *crypto.Secret
	store               AuthenticationStore // store is the database handler object.
}

// CustomClaims groups the JWT-token data fields.
type CustomClaims struct {
	ID        uint64   `json:"id"`
	TenantID  uint64   `json:"tenantId"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Username  string   `json:"username"`

	jwt.StandardClaims
}

// ErrUnauthorized is the error message for all validation failures to avoid being an oracle.
type ErrUnauthorized struct{}

func (e *ErrUnauthorized) Error() string {
	return "invalid credentials"
}

type Err2FARequired struct{}

func (e *Err2FARequired) Error() string {
	return "2FA_REQUIRED"
}

// NewAuthenticationService returns a fresh authentication service instance.
func NewAuthenticationService(cfg *config.Daemon, userer Userer, store AuthenticationStore, daemonEncryptionKey *crypto.Secret) *AuthenticationService {
	return &AuthenticationService{
		userer:              userer,
		signingKey:          cfg.JWT.Secret.PlainText(),
		jwtExpirationTime:   cfg.JWT.ExpirationTime,
		daemonEncryptionKey: daemonEncryptionKey,
		store:               store,
	}
}

// Authenticate verifies whether the user is activated and the provided password is
// correct. It then returns a fresh JWT token for further API access.
func (a *AuthenticationService) Authenticate(ctx context.Context, username, password, totp string) (string, error) {
	user, err := a.store.GetActiveUser(ctx, username)
	if err != nil {
		logger.SecLog.Info(ctx, "user not found", zap.String("username", username))
		return "", &ErrUnauthorized{}
	}
	if user == nil {
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

			if err := a.userer.DeleteTotpRecoveryCode(ctx, &service.DeleteTotpRecoveryCodeReq{
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(expirationTime)).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}
