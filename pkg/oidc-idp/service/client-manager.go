package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type clientManager struct {
	Clients map[string]*goidc.Client
}

var _ goidc.ClientManager = &clientManager{}

func NewClientManager(cfg config.Config) (*clientManager, error) {
	cs := make(map[string]*goidc.Client, len(cfg.Services.OpenIDConnectProvider.Clients))

	for _, c := range cfg.Services.OpenIDConnectProvider.Clients {
		gt := make([]goidc.GrantType, len(c.GrantTypes))
		for i, grantType := range c.GrantTypes {
			gt[i] = goidc.GrantType(grantType)
		}

		rt := make([]goidc.ResponseType, len(c.ResponseTypes))
		for i, responseType := range c.ResponseTypes {
			rt[i] = goidc.ResponseType(responseType)
		}

		h, err := bcrypt.GenerateFromPassword([]byte(c.Secret.PlainText()), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("unable to byrcpt secret: %w", err)
		}

		client := &goidc.Client{
			ID:                         c.ID,
			Secret:                     c.Secret.PlainText(),
			HashedSecret:               string(h),
			RegistrationToken:          c.RegistrationToken,
			CreatedAtTimestamp:         c.CreatedAtTimestamp,
			ExpiresAtTimestamp:         c.ExpiresAtTimestamp,
			IsFederated:                c.IsFederated,
			FederationRegistrationType: goidc.ClientRegistrationType(c.FederationRegistrationType),
			FederationTrustMarkIDs:     c.FederationTrustMarkIDs,
			ClientMeta: goidc.ClientMeta{
				Name:              c.Name,
				SecretExpiresAt:   c.SecretExpiresAt,
				ApplicationType:   goidc.ApplicationType(c.ApplicationType),
				LogoURI:           c.LogoURI,
				Contacts:          c.Contacts,
				PolicyURI:         c.PolicyURI,
				TermsOfServiceURI: c.TermsOfServiceURI,
				RedirectURIs:      c.RedirectURIs,
				RequestURIs:       c.RequestURIs,
				GrantTypes:        gt,
				ResponseTypes:     rt,
				PublicJWKSURI:     c.PublicJWKSURI,
				ScopeIDs:          c.ScopeIDs,
				TokenAuthnMethod:  goidc.ClientAuthnType(c.TokenAuthnMethod),
			},
		}
		cs[c.ID] = client
	}

	return &clientManager{
		Clients: cs,
	}, nil
}

func (m *clientManager) Save(ctx context.Context, c *goidc.Client) error {
	logger.TechLog.Error(ctx, "Save should not be called on client manager", zap.Any("client", c))
	return errors.New("not implemented")
}

func (m *clientManager) Client(ctx context.Context, id string) (*goidc.Client, error) {
	logger.TechLog.Info(ctx, "Fetching client from in-memory storage", zap.String("client_id", id))

	for k, c := range m.Clients {
		logger.TechLog.Debug(ctx, "Available client", zap.String("client_id", k), zap.String("client_name", c.ClientMeta.Name))
	}

	c, exists := m.Clients[id]
	if !exists {
		logger.TechLog.Warn(ctx, "Client not found in in-memory storage", zap.String("client_id", id))
		return nil, errors.New("entity not found")
	}

	// Make sure the content of jwks_uri is cleared from jwks when fetching the
	// client from the in memory storaged.
	if c.PublicJWKSURI != "" {
		c.PublicJWKS = nil
	}

	return c, nil
}

func (m *clientManager) Delete(_ context.Context, id string) error {
	logger.TechLog.Error(context.Background(), "Should not be called, deleting client from in-memory storage", zap.String("client_id", id))
	return errors.New("not implemented")
}
