package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

var _ goidc.GrantSessionManager = NewGrantSessionManager(nil)

type GrantSessionManager struct {
	db *sqlx.DB
}

func NewGrantSessionManager(db *sqlx.DB) *GrantSessionManager {
	return &GrantSessionManager{
		db: db,
	}
}

func (m *GrantSessionManager) Save(_ context.Context, grantSession *goidc.GrantSession) error {

}

func (m *GrantSessionManager) SessionByTokenID(_ context.Context, tokenID string) (*goidc.GrantSession, error) {

}

func (m *GrantSessionManager) SessionByRefreshToken(_ context.Context, tkn string) (*goidc.GrantSession, error) {

}

func (m *GrantSessionManager) Delete(_ context.Context, id string) error {

}

func (m *GrantSessionManager) DeleteByAuthCode(ctx context.Context, code string) error {

}
