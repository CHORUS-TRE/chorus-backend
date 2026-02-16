//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	integration "github.com/CHORUS-TRE/chorus-backend/tests/integration/postgres"
)

const (
	testTenantID = uint64(88888)
	testUserID   = uint64(90000)
	testUsername = "testuser"
)

func setupAuditDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := integration.GetDB()
	require.NoError(t, err)

	// Run audit-specific migration (separate from main chorus migration).
	migrations, tableName, err := migration.GetAuditMigration("postgres")
	require.NoError(t, err)
	_, err = migration.Migrate("postgres", migrations, tableName, db)
	require.NoError(t, err)

	t.Cleanup(func() {
		integration.TruncateTables(db, "audit")
	})

	return db
}

func newTestEntry(action model.AuditAction, description string, details model.AuditDetails) *model.AuditEntry {
	return &model.AuditEntry{
		TenantID:    testTenantID,
		UserID:      testUserID,
		Username:    testUsername,
		Action:      action,
		Description: description,
		Details:     details,
		CreatedAt:   time.Now().UTC(),
	}
}

func TestAuditStorage_Record(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	entry := newTestEntry(
		model.AuditActionUserCreate,
		"Created user with ID 123.",
		model.AuditDetails{"user_id": float64(123), "username": "jdoe"},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.NotZero(t, result.ID)
	require.Equal(t, testTenantID, result.TenantID)
	require.Equal(t, testUserID, result.UserID)
	require.Equal(t, testUsername, result.Username)
	require.Equal(t, model.AuditActionUserCreate, result.Action)
	require.Equal(t, "Created user with ID 123.", result.Description)
	require.Equal(t, "jdoe", result.Details["username"])
	require.InDelta(t, float64(123), result.Details["user_id"], 0)
}

func TestAuditStorage_Record_UserUpdate(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	entry := newTestEntry(
		model.AuditActionUserUpdate,
		"Updated user with ID 90000.",
		model.AuditDetails{"user_id": float64(90000), "username": "hmoto"},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.Equal(t, model.AuditActionUserUpdate, result.Action)
	require.Contains(t, result.Description, "Updated user")
	require.Equal(t, "hmoto", result.Details["username"])
	require.NotNil(t, result.Details["user_id"])
}

func TestAuditStorage_Record_UserDelete(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	entry := newTestEntry(
		model.AuditActionUserDelete,
		"Deleted user with ID 90001.",
		model.AuditDetails{"user_id": float64(90001)},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.Equal(t, model.AuditActionUserDelete, result.Action)
	require.Contains(t, result.Description, "Deleted user")
	require.NotNil(t, result.Details["user_id"])
}

func TestAuditStorage_Record_PasswordChange_NoSensitiveData(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	// Simulate what the middleware records on success: no password fields.
	entry := newTestEntry(
		model.AuditActionUserPasswordChange,
		"Successfully changed password.",
		model.AuditDetails{},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.Equal(t, model.AuditActionUserPasswordChange, result.Action)
	require.Contains(t, result.Description, "Successfully changed password")
	require.NotContains(t, result.Details, "current_password")
	require.NotContains(t, result.Details, "new_password")
}

func TestAuditStorage_Record_PasswordChangeFailed(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	// Simulate what the middleware records on failure: error_message but no passwords.
	entry := newTestEntry(
		model.AuditActionUserPasswordChange,
		"Failed to change password.",
		model.AuditDetails{
			"error_message":    "rpc error: code = Unauthenticated",
			"grpc_status_code": float64(16),
		},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.Contains(t, result.Description, "Failed")
	require.Contains(t, result.Details, "error_message")
	require.NotContains(t, result.Details, "current_password")
	require.NotContains(t, result.Details, "new_password")
}

func TestAuditStorage_Record_TotpReset_NoSensitiveData(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	// Simulate what the middleware records: no TOTP secrets.
	entry := newTestEntry(
		model.AuditActionUserTotpReset,
		"Successfully reset TOTP.",
		model.AuditDetails{},
	)

	result, err := store.Record(ctx, entry)
	require.NoError(t, err)
	require.Equal(t, model.AuditActionUserTotpReset, result.Action)
	require.Contains(t, result.Description, "TOTP")
	require.NotContains(t, result.Details, "totp_secret")
	require.NotContains(t, result.Details, "totp_recovery_codes")
}

func TestAuditStorage_List_FilterByAction(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	// Record entries with different actions.
	_, err := store.Record(ctx, newTestEntry(model.AuditActionUserCreate, "Created user.", model.AuditDetails{}))
	require.NoError(t, err)
	_, err = store.Record(ctx, newTestEntry(model.AuditActionUserUpdate, "Updated user.", model.AuditDetails{}))
	require.NoError(t, err)
	_, err = store.Record(ctx, newTestEntry(model.AuditActionUserDelete, "Deleted user.", model.AuditDetails{}))
	require.NoError(t, err)

	entries, pagination, err := store.List(ctx, testTenantID, nil, &model.AuditFilter{
		Action: model.AuditActionUserCreate,
	})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, uint64(1), pagination.Total)
	require.Equal(t, model.AuditActionUserCreate, entries[0].Action)
}

func TestAuditStorage_Count(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	_, err := store.Record(ctx, newTestEntry(model.AuditActionUserCreate, "Created user 1.", model.AuditDetails{}))
	require.NoError(t, err)
	_, err = store.Record(ctx, newTestEntry(model.AuditActionUserCreate, "Created user 2.", model.AuditDetails{}))
	require.NoError(t, err)

	count, err := store.Count(ctx, testTenantID, &model.AuditFilter{
		Action: model.AuditActionUserCreate,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestAuditStorage_DetailsJSONBRoundTrip(t *testing.T) {
	db := setupAuditDB(t)
	store := NewAuditStorage(db)
	ctx := context.Background()

	details := model.AuditDetails{
		"user_id":          float64(90000),
		"username":         "jdoe",
		"grpc_status_code": float64(0),
		"nested_info":      "some value",
	}

	entry := newTestEntry(model.AuditActionUserCreate, "Created user.", details)
	result, err := store.Record(ctx, entry)
	require.NoError(t, err)

	// Verify JSONB round-trip preserves all keys and values.
	require.Len(t, result.Details, 4)
	require.InDelta(t, float64(90000), result.Details["user_id"], 0)
	require.Equal(t, "jdoe", result.Details["username"])
	require.InDelta(t, float64(0), result.Details["grpc_status_code"], 0)
	require.Equal(t, "some value", result.Details["nested_info"])
}
