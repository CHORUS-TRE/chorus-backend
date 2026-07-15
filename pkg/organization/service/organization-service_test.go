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
	logoFn   func(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error)
	createFn func(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error)
	updateFn func(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error)
	deleteFn func(ctx context.Context, tenantID, id uint64) error
}

func (m *mockOrganizationStore) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error) {
	return m.listFn(ctx, tenantID, pagination)
}

func (m *mockOrganizationStore) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	return m.getFn(ctx, tenantID, id)
}

func (m *mockOrganizationStore) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	return m.logoFn(ctx, tenantID, id)
}

func (m *mockOrganizationStore) CreateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	return m.createFn(ctx, tenantID, organization)
}

func (m *mockOrganizationStore) UpdateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	return m.updateFn(ctx, tenantID, organization)
}

func (m *mockOrganizationStore) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	return m.deleteFn(ctx, tenantID, id)
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

	organizations, pagination, err := svc.ListOrganizations(context.Background(), testTenantID, nil)
	require.NoError(t, err)
	assert.Equal(t, testTenantID, gotTenantID)
	assert.Len(t, organizations, 1)
	assert.Equal(t, uint64(1), pagination.Total)
}

func TestOrganizationService_GetOrganization_WrapsNotFound(t *testing.T) {
	store := &mockOrganizationStore{
		getFn: func(_ context.Context, _, _ uint64) (*model.Organization, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := NewOrganizationService(store)

	_, err := svc.GetOrganization(context.Background(), testTenantID, 42)
	require.Error(t, err)
}

func TestOrganizationService_GetOrganizationLogo(t *testing.T) {
	store := &mockOrganizationStore{
		logoFn: func(_ context.Context, _, _ uint64) (*model.OrganizationLogo, error) {
			return &model.OrganizationLogo{Logo: []byte{0x89, 0x50}, LogoContentType: "image/png"}, nil
		},
	}
	svc := NewOrganizationService(store)

	logo, err := svc.GetOrganizationLogo(context.Background(), testTenantID, 1)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x89, 0x50}, logo.Logo)
	assert.Equal(t, "image/png", logo.LogoContentType)
}

func TestOrganizationService_CreateOrganization(t *testing.T) {
	var gotTenantID uint64
	store := &mockOrganizationStore{
		createFn: func(_ context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
			gotTenantID = tenantID
			created := *organization
			created.ID = 7
			return &created, nil
		},
	}
	svc := NewOrganizationService(store)

	created, err := svc.CreateOrganization(context.Background(), &model.Organization{TenantID: testTenantID, Name: "CHUV"})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), created.ID)
	assert.Equal(t, testTenantID, gotTenantID)
}

func TestOrganizationService_UpdateOrganization_WrapsError(t *testing.T) {
	store := &mockOrganizationStore{
		updateFn: func(_ context.Context, _ uint64, _ *model.Organization) (*model.Organization, error) {
			return nil, sql.ErrNoRows
		},
	}
	svc := NewOrganizationService(store)

	_, err := svc.UpdateOrganization(context.Background(), &model.Organization{ID: 1, TenantID: testTenantID, Name: "CHUV"})
	require.Error(t, err)
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

	err := svc.DeleteOrganization(context.Background(), testTenantID, 9)
	require.NoError(t, err)
	assert.Equal(t, uint64(9), gotID)
}
