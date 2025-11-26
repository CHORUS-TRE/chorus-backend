package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

var _ goidc.LogoutSessionManager = NewLogoutSessionManager(0)

type LogoutSessionManager struct {
	db *sqlx.DB
}

func NewLogoutSessionManager(db *sqlx.DB) *LogoutSessionManager {
	return &LogoutSessionManager{
		db: db,
	}
}

func (m *LogoutSessionManager) Save(_ context.Context, session *goidc.LogoutSession) error {

}

func (m *LogoutSessionManager) SessionByCallbackID(_ context.Context, callbackID string) (*goidc.LogoutSession, error) {

}

func (m *LogoutSessionManager) Delete(_ context.Context, id string) error {

}

func (m *LogoutSessionManager) firstSession(condition func(*goidc.LogoutSession) bool) (*goidc.LogoutSession, bool) {

}
