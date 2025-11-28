package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

type grantSessionRow struct {
	ID                 string  `db:"id"`
	TenantID           *int64  `db:"tenantid"`
	SessionData        []byte  `db:"session_data"`
	TokenID            *string `db:"tokenid"`
	RefreshToken       *string `db:"refreshtoken"`
	AuthCode           *string `db:"authcode"`
	CreatedAtTimestamp int     `db:"createdattimestamp"`
}

func (m *GrantSessionManager) Save(ctx context.Context, grantSession *goidc.GrantSession) error {
	sessionData, err := json.Marshal(grantSession)
	if err != nil {
		return fmt.Errorf("unable to marshal grant session: %w", err)
	}

	const query = `
		INSERT INTO grant_sessions (id, tenantid, session_data, tokenid, refreshtoken, authcode, createdattimestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			tokenid = EXCLUDED.tokenid,
			refreshtoken = EXCLUDED.refreshtoken,
			authcode = EXCLUDED.authcode,
			updatedat = now()
	`

	var tokenID, refreshToken, authCode *string
	if grantSession.TokenID != "" {
		tokenID = &grantSession.TokenID
	}
	if grantSession.RefreshToken != "" {
		refreshToken = &grantSession.RefreshToken
	}
	if grantSession.AuthCode != "" {
		authCode = &grantSession.AuthCode
	}

	_, err = m.db.ExecContext(ctx, query, grantSession.ID, nil, sessionData, tokenID, refreshToken, authCode, grantSession.CreatedAtTimestamp)
	if err != nil {
		return fmt.Errorf("unable to save grant session: %w", err)
	}

	return nil
}

func (m *GrantSessionManager) SessionByTokenID(ctx context.Context, tokenID string) (*goidc.GrantSession, error) {
	const query = `SELECT id, tenantid, session_data, tokenid, refreshtoken, authcode, createdattimestamp 
		FROM grant_sessions WHERE tokenid = $1`

	var row grantSessionRow
	if err := m.db.GetContext(ctx, &row, query, tokenID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get grant session by token id: %w", err)
	}

	return unmarshalGrantSession(&row)
}

func (m *GrantSessionManager) SessionByRefreshToken(ctx context.Context, tkn string) (*goidc.GrantSession, error) {
	const query = `SELECT id, tenantid, session_data, tokenid, refreshtoken, authcode, createdattimestamp 
		FROM grant_sessions WHERE refreshtoken = $1`

	var row grantSessionRow
	if err := m.db.GetContext(ctx, &row, query, tkn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get grant session by refresh token: %w", err)
	}

	return unmarshalGrantSession(&row)
}

func (m *GrantSessionManager) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM grant_sessions WHERE id = $1`

	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("unable to delete grant session: %w", err)
	}

	return nil
}

func (m *GrantSessionManager) DeleteByAuthCode(ctx context.Context, code string) error {
	const query = `DELETE FROM grant_sessions WHERE authcode = $1`

	_, err := m.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("unable to delete grant session by auth code: %w", err)
	}

	return nil
}

func unmarshalGrantSession(row *grantSessionRow) (*goidc.GrantSession, error) {
	var session goidc.GrantSession
	if err := json.Unmarshal(row.SessionData, &session); err != nil {
		return nil, fmt.Errorf("unable to unmarshal grant session data: %w", err)
	}
	return &session, nil
}
