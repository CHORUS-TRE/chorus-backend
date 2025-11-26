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

var _ goidc.LogoutSessionManager = NewLogoutSessionManager(nil)

type LogoutSessionManager struct {
	db *sqlx.DB
}

func NewLogoutSessionManager(db *sqlx.DB) *LogoutSessionManager {
	return &LogoutSessionManager{
		db: db,
	}
}

type logoutSessionRow struct {
	ID                 string  `db:"id"`
	TenantID           *int64  `db:"tenantid"`
	SessionData        []byte  `db:"session_data"`
	CallbackID         *string `db:"callbackid"`
	CreatedAtTimestamp int     `db:"createdattimestamp"`
}

func (m *LogoutSessionManager) Save(ctx context.Context, session *goidc.LogoutSession) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("unable to marshal logout session: %w", err)
	}

	const query = `
		INSERT INTO logout_sessions (id, tenantid, session_data, callbackid, createdattimestamp)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			callbackid = EXCLUDED.callbackid,
			updatedat = now()
	`

	var callbackID *string
	if session.CallbackID != "" {
		callbackID = &session.CallbackID
	}

	_, err = m.db.ExecContext(ctx, query, session.ID, nil, sessionData, callbackID, session.CreatedAtTimestamp)
	if err != nil {
		return fmt.Errorf("unable to save logout session: %w", err)
	}

	return nil
}

func (m *LogoutSessionManager) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.LogoutSession, error) {
	const query = `SELECT id, tenantid, session_data, callbackid, createdattimestamp 
		FROM logout_sessions WHERE callbackid = $1`

	var row logoutSessionRow
	if err := m.db.GetContext(ctx, &row, query, callbackID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("entity not found")
		}
		return nil, fmt.Errorf("unable to get logout session by callback id: %w", err)
	}

	return unmarshalLogoutSession(&row)
}

func (m *LogoutSessionManager) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM logout_sessions WHERE id = $1`

	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("unable to delete logout session: %w", err)
	}

	return nil
}

func unmarshalLogoutSession(row *logoutSessionRow) (*goidc.LogoutSession, error) {
	var session goidc.LogoutSession
	if err := json.Unmarshal(row.SessionData, &session); err != nil {
		return nil, fmt.Errorf("unable to unmarshal logout session data: %w", err)
	}
	return &session, nil
}
