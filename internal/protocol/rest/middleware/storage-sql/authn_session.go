package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

var _ goidc.AuthnSessionManager = &authnSessionManager{}

type authnSessionManager struct {
	db *sqlx.DB
}

func NewAuthnSessionManager(db *sqlx.DB) *authnSessionManager {
	return &authnSessionManager{
		db: db,
	}
}

func (m *authnSessionManager) Save(_ context.Context, session *goidc.AuthnSession) error {

	return nil
}

func (m *authnSessionManager) SessionByCallbackID(_ context.Context, callbackID string) (*goidc.AuthnSession, error) {

	return nil, nil
}

func (m *authnSessionManager) SessionByAuthCode(_ context.Context, authorizationCode string) (*goidc.AuthnSession, error) {

	return nil, nil
}

func (m *authnSessionManager) SessionByPushedAuthReqID(_ context.Context, requestURI string) (*goidc.AuthnSession, error) {

	return nil, nil
}

func (m *authnSessionManager) SessionByCIBAAuthID(_ context.Context, id string) (*goidc.AuthnSession, error) {

	return nil, nil
}

func (m *authnSessionManager) Delete(_ context.Context, id string) error {
	return nil
}
