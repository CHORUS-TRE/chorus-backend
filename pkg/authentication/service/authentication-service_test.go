//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	userModel "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	userService "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

// --- fakes ---

type fakeUserer struct {
	createUserFn             func(ctx context.Context, req userService.CreateUserReq) (*userModel.User, error)
	createUserRolesFn        func(ctx context.Context, tenantID, userID uint64, roles []userModel.UserRole) error
	getUserFn                func(ctx context.Context, req userService.GetUserReq) (*userModel.User, error)
	getTotpRecoveryCodesFn   func(ctx context.Context, tenantID, userID uint64) ([]*userModel.TotpRecoveryCode, error)
	deleteTotpRecoveryCodeFn func(ctx context.Context, req *userService.DeleteTotpRecoveryCodeReq) error
}

func (f *fakeUserer) CreateUser(ctx context.Context, req userService.CreateUserReq) (*userModel.User, error) {
	return f.createUserFn(ctx, req)
}
func (f *fakeUserer) CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []userModel.UserRole) error {
	return f.createUserRolesFn(ctx, tenantID, userID, roles)
}
func (f *fakeUserer) GetUser(ctx context.Context, req userService.GetUserReq) (*userModel.User, error) {
	return f.getUserFn(ctx, req)
}
func (f *fakeUserer) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*userModel.TotpRecoveryCode, error) {
	return f.getTotpRecoveryCodesFn(ctx, tenantID, userID)
}
func (f *fakeUserer) DeleteTotpRecoveryCode(ctx context.Context, req *userService.DeleteTotpRecoveryCodeReq) error {
	return f.deleteTotpRecoveryCodeFn(ctx, req)
}

type fakeStore struct {
	getActiveUserFn func(ctx context.Context, username, source string) (*userModel.User, error)
}

func (f *fakeStore) GetActiveUser(ctx context.Context, username, source string) (*userModel.User, error) {
	return f.getActiveUserFn(ctx, username, source)
}

// --- helpers ---

// oauthTestServer returns a test HTTP server handling the token exchange and
// user info endpoints of a minimal OAuth2/OIDC provider.
func oauthTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sub":         "testuser",
			"given_name":  "Test",
			"family_name": "User",
			"email":       "test@example.com",
		})
	})
	return httptest.NewServer(mux)
}

// oauthSvc builds a minimal AuthenticationService wired to the given test server.
func oauthSvc(selfServiceTenantID uint64, srv *httptest.Server, userer Userer, store AuthenticationStore) *AuthenticationService {
	const providerID = "test-provider"
	return &AuthenticationService{
		selfServiceTenantID: selfServiceTenantID,
		signingKey:          "test-signing-key",
		jwtExpirationTime:   time.Hour,
		oauthHTTPClients:    map[string]*http.Client{},
		modes: map[string]config.Mode{
			providerID: {
				Type:    "openid",
				Enabled: true,
				OpenID: config.OpenID{
					ID:          providerID,
					UserInfoURL: srv.URL + "/userinfo",
				},
			},
		},
		oauthConfigs: map[string]*oauth2.Config{
			providerID: {
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				Endpoint: oauth2.Endpoint{
					AuthURL:  srv.URL + "/auth",
					TokenURL: srv.URL + "/token",
				},
				RedirectURL: "http://localhost/callback",
			},
		},
		userer: userer,
		store:  store,
	}
}

// --- tests ---

func TestVerifyPassword(t *testing.T) {
	const password = "superpassword"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.Nil(t, err)
	require.True(t, verifyPassword(string(hash), password))
}

func TestNewAuthenticationService_SetsSelfServiceTenantID(t *testing.T) {
	const tenantID = uint64(99)
	cfg := config.Config{}
	cfg.Services.AuthenticationService.SelfService.TenantID = tenantID

	svc, err := NewAuthenticationService(cfg, &fakeUserer{}, &fakeStore{}, nil)
	require.NoError(t, err)
	require.Equal(t, tenantID, svc.selfServiceTenantID)
}

func TestOAuthCallback_ExistingUser_SkipsUserCreation(t *testing.T) {
	srv := oauthTestServer(t)
	defer srv.Close()

	createCalled := false
	svc := oauthSvc(42, srv,
		&fakeUserer{
			createUserFn: func(_ context.Context, _ userService.CreateUserReq) (*userModel.User, error) {
				createCalled = true
				return nil, nil
			},
		},
		&fakeStore{
			getActiveUserFn: func(_ context.Context, _, _ string) (*userModel.User, error) {
				return &userModel.User{ID: 1, TenantID: 42, Username: "testuser", Source: "test-provider"}, nil
			},
		},
	)

	_, _, _, err := svc.OAuthCallback(context.Background(), "test-provider", "state", "state", "code")
	require.NoError(t, err)
	require.False(t, createCalled, "CreateUser must not be called for an existing user")
}
