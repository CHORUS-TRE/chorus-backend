package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

type authnSessionRow struct {
	ID                 string  `db:"id"`
	TenantID           *int64  `db:"tenantid"`
	SessionData        []byte  `db:"session_data"`
	CallbackID         *string `db:"callbackid"`
	AuthCode           *string `db:"authcode"`
	PushedAuthReqID    *string `db:"pushedauthreqid"`
	CIBAAuthID         *string `db:"cibaauthid"`
	CreatedAtTimestamp int     `db:"createdattimestamp"`
}

func (m *authnSessionManager) Save(ctx context.Context, session *goidc.AuthnSession) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("unable to marshal session: %w", err)
	}

	const query = `
		INSERT INTO authn_sessions (id, tenantid, session_data, callbackid, authcode, pushedauthreqid, cibaauthid, createdattimestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			callbackid = EXCLUDED.callbackid,
			authcode = EXCLUDED.authcode,
			pushedauthreqid = EXCLUDED.pushedauthreqid,
			cibaauthid = EXCLUDED.cibaauthid,
			updatedat = now()
	`

	var callbackID, authCode, pushedAuthReqID, cibaAuthID *string
	if session.CallbackID != "" {
		callbackID = &session.CallbackID
	}
	if session.AuthCode != "" {
		authCode = &session.AuthCode
	}
	if session.PushedAuthReqID != "" {
		pushedAuthReqID = &session.PushedAuthReqID
	}
	if session.CIBAAuthID != "" {
		cibaAuthID = &session.CIBAAuthID
	}

	_, err = m.db.ExecContext(ctx, query, session.ID, nil, sessionData, callbackID, authCode, pushedAuthReqID, cibaAuthID, session.CreatedAtTimestamp)
	if err != nil {
		return fmt.Errorf("unable to save authn session: %w", err)
	}

	return nil
}

func (m *authnSessionManager) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.AuthnSession, error) {
	const query = `SELECT id, tenantid, session_data, callbackid, authcode, pushedauthreqid, cibaauthid, createdattimestamp 
		FROM authn_sessions WHERE callbackid = $1`

	var row authnSessionRow
	if err := m.db.GetContext(ctx, &row, query, callbackID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get authn session by callback id: %w", err)
	}

	return unmarshalAuthnSession(&row)
}

func (m *authnSessionManager) SessionByAuthCode(ctx context.Context, authorizationCode string) (*goidc.AuthnSession, error) {
	const query = `SELECT id, tenantid, session_data, callbackid, authcode, pushedauthreqid, cibaauthid, createdattimestamp 
		FROM authn_sessions WHERE authcode = $1`

	var row authnSessionRow
	if err := m.db.GetContext(ctx, &row, query, authorizationCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get authn session by auth code: %w", err)
	}

	return unmarshalAuthnSession(&row)
}

func (m *authnSessionManager) SessionByPushedAuthReqID(ctx context.Context, requestURI string) (*goidc.AuthnSession, error) {
	const query = `SELECT id, tenantid, session_data, callbackid, authcode, pushedauthreqid, cibaauthid, createdattimestamp 
		FROM authn_sessions WHERE pushedauthreqid = $1`

	var row authnSessionRow
	if err := m.db.GetContext(ctx, &row, query, requestURI); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get authn session by pushed auth req id: %w", err)
	}

	return unmarshalAuthnSession(&row)
}

func (m *authnSessionManager) SessionByCIBAAuthID(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	const query = `SELECT id, tenantid, session_data, callbackid, authcode, pushedauthreqid, cibaauthid, createdattimestamp 
		FROM authn_sessions WHERE cibaauthid = $1`

	var row authnSessionRow
	if err := m.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get authn session by ciba auth id: %w", err)
	}

	return unmarshalAuthnSession(&row)
}

func (m *authnSessionManager) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM authn_sessions WHERE id = $1`

	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("unable to delete authn session: %w", err)
	}

	return nil
}

func unmarshalAuthnSession(row *authnSessionRow) (*goidc.AuthnSession, error) {
	var session goidc.AuthnSession
	if err := json.Unmarshal(row.SessionData, &session); err != nil {
		return nil, fmt.Errorf("unable to unmarshal session data: %w", err)
	}
	return &session, nil
}
