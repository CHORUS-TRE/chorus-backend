package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/rest/middleware/oidc-idp/authutil"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
)

type OIDCProviderService interface {
	AddOIDCMiddleware(h http.Handler) http.Handler
}

type oidcProviderService struct {
	cfg                  config.Config
	op                   *provider.Provider
	authnSessionManager  goidc.AuthnSessionManager
	logoutSessionManager goidc.LogoutSessionManager
	clientManager        goidc.ClientManager
	grantSessionManager  goidc.GrantSessionManager
}

func NewOIDCProviderService(cfg config.Config, authnSessionManager goidc.AuthnSessionManager, clientManager goidc.ClientManager, logoutSessionManager goidc.LogoutSessionManager, grantSessionManager goidc.GrantSessionManager) (OIDCProviderService, error) {
	s := &oidcProviderService{
		cfg:                  cfg,
		authnSessionManager:  authnSessionManager,
		clientManager:        clientManager,
		logoutSessionManager: logoutSessionManager,
		grantSessionManager:  grantSessionManager,
	}

	err := s.init()
	if err != nil {
		return nil, fmt.Errorf("error initianting service: %w", err)
	}

	return s, nil
}

func (s *oidcProviderService) init() error {
	if !s.cfg.Services.OpenIDConnectProvider.Enabled {
		return nil
	}

	// authnSessionStorage := storage.NewAuthnSessionManager(1000)
	// clientStorage, err := storage.NewClientManager(s.cfg)
	// if err != nil {
	// 	logger.TechLog.Fatal(context.Background(), "unable to instantiate client storage for OIDC provider", zap.Error(err))
	// }
	// grantSessionStorage := storage.NewGrantSessionManager(1000)
	// logoutSessionStorage := storage.NewLogoutSessionManager(1000)

	op, err := provider.New(
		goidc.ProfileOpenID,
		authutil.Issuer,
		authutil.PrivateJWKSFunc(s.cfg),
		provider.WithScopes(authutil.Scopes...),
		provider.WithIDTokenSignatureAlgs(goidc.HS256, goidc.None),
		provider.WithUserInfoSignatureAlgs(goidc.HS256, goidc.None),
		provider.WithPAR(nil, 10),
		provider.WithJAR(goidc.HS256, goidc.None),
		provider.WithJARByReference(false),
		provider.WithJARM(goidc.HS256),
		provider.WithTokenAuthnMethods(
			goidc.ClientAuthnSecretBasic,
			goidc.ClientAuthnSecretPost,
			goidc.ClientAuthnPrivateKeyJWT,
		),
		// provider.WithPrivateKeyJWTSignatureAlgs(goidc.HS256),
		provider.WithSecretJWTSignatureAlgs(goidc.HS256),
		provider.WithIssuerResponseParameter(),
		provider.WithClaimsParameter(),
		provider.WithPKCE(goidc.CodeChallengeMethodSHA256),
		provider.WithImplicitGrant(),
		provider.WithAuthorizationCodeGrant(),
		provider.WithRefreshTokenGrant(authutil.IssueRefreshToken, 600),
		provider.WithClaims(authutil.Claims[0], authutil.Claims...),
		provider.WithACRs(authutil.ACRs[0], authutil.ACRs...),
		// provider.WithDCR(authutil.DCRFunc, authutil.ValidateInitialTokenFunc),
		provider.WithTokenOptions(authutil.TokenOptionsFunc(goidc.HS256)),
		provider.WithHTTPClientFunc(authutil.HTTPClient),
		provider.WithPolicies(authutil.Policy(s.cfg)),
		provider.WithNotifyErrorFunc(authutil.ErrorLoggingFunc),
		provider.WithRenderErrorFunc(authutil.RenderError()),
		provider.WithDisplayValues(authutil.DisplayValues[0], authutil.DisplayValues...),
		provider.WithSubIdentifierTypes(goidc.SubIdentifierPublic, goidc.SubIdentifierPairwise),
		provider.WithLogout(authutil.HandleLogout(), authutil.LogoutPolicy()),
		provider.WithAuthnSessionStorage(s.authnSessionManager),
		provider.WithClientStorage(s.clientManager),
		provider.WithGrantSessionStorage(s.grantSessionManager),
		provider.WithLogoutSessionManager(s.logoutSessionManager),
	)
	if err != nil {
		return fmt.Errorf("unable to instantiate OIDC provider: %w", err)
	}

	s.op = op

	return nil
}

func (s *oidcProviderService) AddOIDCMiddleware(h http.Handler) http.Handler {
	if !s.cfg.Services.OpenIDConnectProvider.Enabled {
		return h
	}

	openidIDPHandler := s.op.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if strings.HasPrefix(r.RequestURI, "/openid-connect") {
			handler := http.StripPrefix("/openid-connect", openidIDPHandler)
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If not "/openid-connect", passing to the next middleware
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
