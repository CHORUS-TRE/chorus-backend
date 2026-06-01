package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
)

type mockStore struct {
	getSettings    *model.PlatformSettings
	getErr         error
	upsertSettings *model.PlatformSettings
	upsertErr      error
	capturedGet    uint64
	capturedUpsert *model.PlatformSettings
}

func (m *mockStore) GetPlatformSettings(_ context.Context, tenantID uint64) (*model.PlatformSettings, error) {
	m.capturedGet = tenantID
	return m.getSettings, m.getErr
}

func (m *mockStore) UpsertPlatformSettings(_ context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	m.capturedUpsert = settings
	return m.upsertSettings, m.upsertErr
}

func ctxWithTenant(tenantID uint64) context.Context {
	return context.WithValue(context.Background(), jwt_model.JWTClaimsContextKey, &jwt_model.JWTClaims{
		TenantID: tenantID,
	})
}

func TestGetPlatformSettings_WithTenantFromJWT(t *testing.T) {
	want := &model.PlatformSettings{TenantID: 42, Title: "My Platform"}
	store := &mockStore{getSettings: want}
	svc := NewPlatformSettingsService(store)

	got, err := svc.GetPlatformSettings(ctxWithTenant(42))

	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, uint64(42), store.capturedGet)
}

func TestGetPlatformSettings_FallsBackToTenant1WhenNoJWT(t *testing.T) {
	want := &model.PlatformSettings{TenantID: 1}
	store := &mockStore{getSettings: want}
	svc := NewPlatformSettingsService(store)

	got, err := svc.GetPlatformSettings(context.Background())

	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, uint64(1), store.capturedGet, "should fall back to tenant 1 for public endpoint")
}

func TestGetPlatformSettings_PropagatesStoreError(t *testing.T) {
	store := &mockStore{getErr: errors.New("db down")}
	svc := NewPlatformSettingsService(store)

	_, err := svc.GetPlatformSettings(ctxWithTenant(1))

	require.Error(t, err)
}

func TestUpdatePlatformSettings_SetsTenantIDFromJWT(t *testing.T) {
	returned := &model.PlatformSettings{TenantID: 7, Title: "Updated"}
	store := &mockStore{upsertSettings: returned}
	svc := NewPlatformSettingsService(store)

	got, err := svc.UpdatePlatformSettings(ctxWithTenant(7), &model.PlatformSettings{Title: "Updated"})

	require.NoError(t, err)
	assert.Equal(t, returned, got)
	assert.Equal(t, uint64(7), store.capturedUpsert.TenantID, "service must set TenantID from JWT before calling store")
}

func TestUpdatePlatformSettings_ReturnsErrorWhenNoJWT(t *testing.T) {
	store := &mockStore{}
	svc := NewPlatformSettingsService(store)

	_, err := svc.UpdatePlatformSettings(context.Background(), &model.PlatformSettings{})

	require.Error(t, err)
	assert.Nil(t, store.capturedUpsert, "store should not be called without a valid JWT")
}

func TestUpdatePlatformSettings_PropagatesStoreError(t *testing.T) {
	store := &mockStore{upsertErr: errors.New("constraint violation")}
	svc := NewPlatformSettingsService(store)

	_, err := svc.UpdatePlatformSettings(ctxWithTenant(1), &model.PlatformSettings{})

	require.Error(t, err)
}
