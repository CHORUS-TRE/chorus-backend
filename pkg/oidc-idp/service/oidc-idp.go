package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/oidc-idp/service/authutil"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
)

type OIDCProviderService interface {
	AddOIDCMiddleware(h http.Handler) http.Handler
}

type oidcProviderService struct {
	cfg                  config.Config
	op                   *provider.Provider
	authorizer           authorization_service.Authorizer
	authnSessionManager  goidc.AuthnSessionManager
	logoutSessionManager goidc.LogoutSessionManager
	clientManager        goidc.ClientManager
	grantSessionManager  goidc.GrantSessionManager
	userService          authutil.Userer
}

func NewOIDCProviderService(cfg config.Config, authorizer authorization_service.Authorizer, authnSessionManager goidc.AuthnSessionManager, clientManager goidc.ClientManager, logoutSessionManager goidc.LogoutSessionManager, grantSessionManager goidc.GrantSessionManager, userService authutil.Userer) (OIDCProviderService, error) {
	s := &oidcProviderService{
		cfg:                  cfg,
		authorizer:           authorizer,
		authnSessionManager:  authnSessionManager,
		clientManager:        clientManager,
		logoutSessionManager: logoutSessionManager,
		grantSessionManager:  grantSessionManager,
		userService:          userService,
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

	scopes := []goidc.Scope{}
	for _, scopeStr := range s.cfg.Services.OpenIDConnectProvider.Scopes {
		scopes = append(scopes, goidc.NewScope(scopeStr))
	}

	op, err := provider.New(
		goidc.ProfileOpenID,
		s.cfg.Services.OpenIDConnectProvider.IssuerURL,
		authutil.PrivateJWKSFunc(s.cfg),
		provider.WithScopes(scopes...),
		provider.WithIDTokenSignatureAlgs(goidc.RS256, goidc.None),
		provider.WithUserInfoSignatureAlgs(goidc.RS256, goidc.None),
		provider.WithPAR(nil, 10),
		provider.WithJAR(goidc.RS256, goidc.None),
		provider.WithJARByReference(false),
		provider.WithJARM(goidc.RS256),
		provider.WithTokenAuthnMethods(
			goidc.ClientAuthnSecretBasic,
			goidc.ClientAuthnSecretPost,
			goidc.ClientAuthnPrivateKeyJWT,
		),
		provider.WithPrivateKeyJWTSignatureAlgs(goidc.RS256),
		provider.WithIssuerResponseParameter(),
		provider.WithClaimsParameter(),
		provider.WithPKCE(goidc.CodeChallengeMethodSHA256),
		provider.WithImplicitGrant(),
		provider.WithAuthorizationCodeGrant(),
		provider.WithRefreshTokenGrant(authutil.IssueRefreshToken, 600),
		provider.WithClaims(authutil.Claims[0], authutil.Claims...),
		provider.WithACRs(authutil.ACRs[0], authutil.ACRs...),
		// provider.WithDCR(authutil.DCRFunc, authutil.ValidateInitialTokenFunc),
		provider.WithTokenOptions(authutil.TokenOptionsFunc(goidc.RS256)),
		provider.WithHTTPClientFunc(authutil.HTTPClient),
		provider.WithPolicies(authutil.Policy(s.cfg, s.userService, s.authorizer)),
		provider.WithNotifyErrorFunc(authutil.ErrorLoggingFunc),
		provider.WithRenderErrorFunc(authutil.RenderError(s.cfg)),
		provider.WithDisplayValues(authutil.DisplayValues[0], authutil.DisplayValues...),
		provider.WithSubIdentifierTypes(goidc.SubIdentifierPublic, goidc.SubIdentifierPairwise),
		provider.WithLogout(authutil.HandleLogout(s.cfg), authutil.LogoutPolicy(s.cfg)),
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
