//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
)

const testTenantID = uint64(1)

type mockOrganizationStore struct {
	listFn   func(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error)
	getFn    func(ctx context.Context, tenantID, id uint64) (*model.Organization, error)
	logoFn   func(ctx context.Context, tenantID, id uint64) ([]byte, *string, error)
	createFn func(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error)
	updateFn func(ctx context.Context, tenantID uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error)
	deleteFn func(ctx context.Context, tenantID, id uint64) error
}

func (m *mockOrganizationStore) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error) {
	return m.listFn(ctx, tenantID, pagination)
}

func (m *mockOrganizationStore) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	return m.getFn(ctx, tenantID, id)
}

func (m *mockOrganizationStore) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) ([]byte, *string, error) {
	return m.logoFn(ctx, tenantID, id)
}

func (m *mockOrganizationStore) CreateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	return m.createFn(ctx, tenantID, organization)
}

func (m *mockOrganizationStore) UpdateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error) {
	return m.updateFn(ctx, tenantID, organization, updateLogo)
}

func (m *mockOrganizationStore) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	return m.deleteFn(ctx, tenantID, id)
}

func TestOrganizationService_GetOrganization_WrapsNotFound(t *testing.T) {
	store := &mockOrganizationStore{
		getFn: func(_ context.Context, _, _ uint64) (*model.Organization, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := NewOrganizationService(store)

	_, err := svc.GetOrganization(context.Background(), GetOrganizationReq{TenantID: testTenantID, ID: 42})
	require.Error(t, err)
}

func TestOrganizationService_ListOrganizations_ScopesToTenant(t *testing.T) {
	var gotTenantID uint64
	store := &mockOrganizationStore{
		listFn: func(_ context.Context, tenantID uint64, _ *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error) {
			gotTenantID = tenantID
			return []*model.Organization{{ID: 1, TenantID: tenantID}}, &common_model.PaginationResult{Total: 1}, nil
		},
	}
	svc := NewOrganizationService(store)

	organizations, pagination, err := svc.ListOrganizations(context.Background(), ListOrganizationsReq{TenantID: testTenantID})
	require.NoError(t, err)
	assert.Equal(t, testTenantID, gotTenantID)
	assert.Len(t, organizations, 1)
	assert.Equal(t, uint64(1), pagination.Total)
}

func TestOrganizationService_CreateOrganization_PassesLogoThrough(t *testing.T) {
	var gotOrganization *model.Organization
	store := &mockOrganizationStore{
		createFn: func(_ context.Context, _ uint64, organization *model.Organization) (*model.Organization, error) {
			gotOrganization = organization
			created := *organization
			created.ID = 7
			return &created, nil
		},
	}
	svc := NewOrganizationService(store)

	contentType := "image/png"
	created, err := svc.CreateOrganization(context.Background(), CreateOrganizationReq{
		TenantID:        testTenantID,
		Name:            "CHUV",
		Logo:            []byte{0x89, 0x50, 0x4E, 0x47},
		LogoContentType: &contentType,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), created.ID)
	assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, gotOrganization.Logo)
	assert.Equal(t, &contentType, gotOrganization.LogoContentType)
}

func TestOrganizationService_UpdateOrganization_OmittedLogoLeavesLogoUntouched(t *testing.T) {
	var gotUpdateLogo bool
	var gotOrganization *model.Organization
	store := &mockOrganizationStore{
		updateFn: func(_ context.Context, _ uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error) {
			gotOrganization = organization
			gotUpdateLogo = updateLogo
			return organization, nil
		},
	}
	svc := NewOrganizationService(store)

	_, err := svc.UpdateOrganization(context.Background(), UpdateOrganizationReq{
		TenantID: testTenantID,
		ID:       1,
		Name:     "CHUV renamed",
		// Logo intentionally omitted (nil).
	})
	require.NoError(t, err)
	assert.False(t, gotUpdateLogo, "updateLogo must be false when the request omits the logo, so the store preserves the existing one")
	assert.Nil(t, gotOrganization.Logo)
}

func TestOrganizationService_UpdateOrganization_ProvidedLogoReplacesExisting(t *testing.T) {
	var gotUpdateLogo bool
	var gotOrganization *model.Organization
	store := &mockOrganizationStore{
		updateFn: func(_ context.Context, _ uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error) {
			gotOrganization = organization
			gotUpdateLogo = updateLogo
			return organization, nil
		},
	}
	svc := NewOrganizationService(store)

	newLogo := []byte{0xFF, 0xD8, 0xFF}
	contentType := "image/jpeg"
	_, err := svc.UpdateOrganization(context.Background(), UpdateOrganizationReq{
		TenantID:        testTenantID,
		ID:              1,
		Name:            "CHUV renamed",
		Logo:            newLogo,
		LogoContentType: &contentType,
	})
	require.NoError(t, err)
	assert.True(t, gotUpdateLogo, "updateLogo must be true when the request provides non-empty logo bytes")
	assert.Equal(t, newLogo, gotOrganization.Logo)
}

func TestOrganizationService_UpdateOrganization_EmptyLogoLeavesLogoUntouched(t *testing.T) {
	var gotUpdateLogo bool
	store := &mockOrganizationStore{
		updateFn: func(_ context.Context, _ uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error) {
			gotUpdateLogo = updateLogo
			return organization, nil
		},
	}
	svc := NewOrganizationService(store)

	_, err := svc.UpdateOrganization(context.Background(), UpdateOrganizationReq{
		TenantID: testTenantID,
		ID:       1,
		Name:     "CHUV",
		Logo:     []byte{},
	})
	require.NoError(t, err)
	assert.False(t, gotUpdateLogo, "an empty logo is treated the same as an omitted one - there is no way to clear an existing logo")
}

func TestOrganizationService_DeleteOrganization(t *testing.T) {
	var gotID uint64
	store := &mockOrganizationStore{
		deleteFn: func(_ context.Context, _, id uint64) error {
			gotID = id
			return nil
		},
	}
	svc := NewOrganizationService(store)

	err := svc.DeleteOrganization(context.Background(), DeleteOrganizationReq{TenantID: testTenantID, ID: 9})
	require.NoError(t, err)
	assert.Equal(t, uint64(9), gotID)
}

func TestOrganizationService_GetOrganizationLogo(t *testing.T) {
	contentType := "image/png"
	store := &mockOrganizationStore{
		logoFn: func(_ context.Context, _, _ uint64) ([]byte, *string, error) {
			return []byte{0x89, 0x50}, &contentType, nil
		},
	}
	svc := NewOrganizationService(store)

	logo, gotContentType, err := svc.GetOrganizationLogo(context.Background(), GetOrganizationLogoReq{TenantID: testTenantID, ID: 1})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x89, 0x50}, logo)
	assert.Equal(t, &contentType, gotContentType)
}
