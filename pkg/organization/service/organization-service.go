package service

import (
	"context"
	"fmt"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
)

// Organizationer defines the organization business logic.
type Organizationer interface {
	ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error)
	GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error)
	GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error)
	CreateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error)
	DeleteOrganization(ctx context.Context, tenantID, id uint64) error
}

// OrganizationStore defines the organization persistence layer.
type OrganizationStore interface {
	ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error)
	GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error)
	GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error)
	CreateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error)
	DeleteOrganization(ctx context.Context, tenantID, id uint64) error
}

var _ Organizationer = (*OrganizationService)(nil)

type OrganizationService struct {
	store OrganizationStore
}

func NewOrganizationService(store OrganizationStore) *OrganizationService {
	return &OrganizationService{store: store}
}

func (s *OrganizationService) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error) {
	organizations, paginationRes, err := s.store.ListOrganizations(ctx, tenantID, pagination)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, "Unable to list organizations")
	}

	return organizations, paginationRes, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	organization, err := s.store.GetOrganization(ctx, tenantID, id)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get organization %v", id))
	}

	return organization, nil
}

func (s *OrganizationService) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	logo, err := s.store.GetOrganizationLogo(ctx, tenantID, id)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get logo for organization %v", id))
	}

	return logo, nil
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	createdOrganization, err := s.store.CreateOrganization(ctx, organization.TenantID, organization)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to create organization")
	}

	return createdOrganization, nil
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	updatedOrganization, err := s.store.UpdateOrganization(ctx, organization.TenantID, organization)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to update organization %v", organization.ID))
	}

	return updatedOrganization, nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	if err := s.store.DeleteOrganization(ctx, tenantID, id); err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to delete organization %v", id))
	}

	return nil
}
