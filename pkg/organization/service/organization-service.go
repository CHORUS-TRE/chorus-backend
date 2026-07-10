package service

import (
	"context"
	"fmt"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
)

// maxLogoSizeBytes documents the 512 KB limit enforced by the `max=524288` validator
// tags below; struct tags can't reference Go constants, so keep the two in sync by hand.
const maxLogoSizeBytes = 512 * 1024

type ListOrganizationsReq struct {
	TenantID   uint64
	Pagination *common_model.Pagination
}

type GetOrganizationReq struct {
	TenantID uint64
	ID       uint64 `validate:"required,min=1"`
}

type GetOrganizationLogoReq struct {
	TenantID uint64
	ID       uint64 `validate:"required,min=1"`
}

type CreateOrganizationReq struct {
	TenantID uint64

	Name        string  `validate:"required,generalstring,max=255"`
	Description *string `validate:"omitempty,max=250,generalstring"`

	Logo            []byte  `validate:"omitempty,max=524288"`
	LogoContentType *string `validate:"omitempty,oneof=image/png image/jpeg image/webp"`

	Country *string `validate:"omitempty,iso3166_1_alpha2"`
	City    *string `validate:"omitempty,max=100,generalstring"`

	ContactUserID *uint64
	WebsiteURL    *string `validate:"omitempty,max=2048,url"`
}

type UpdateOrganizationReq struct {
	TenantID uint64
	ID       uint64 `validate:"required,min=1"`

	Name        string  `validate:"required,generalstring,max=255"`
	Description *string `validate:"omitempty,max=250,generalstring"`

	// An empty Logo means "not provided, leave the existing logo untouched" - there is
	// no way to clear an existing logo, only add one or replace it with another.
	Logo            []byte  `validate:"omitempty,max=524288"`
	LogoContentType *string `validate:"omitempty,oneof=image/png image/jpeg image/webp"`

	Country *string `validate:"omitempty,iso3166_1_alpha2"`
	City    *string `validate:"omitempty,max=100,generalstring"`

	ContactUserID *uint64
	WebsiteURL    *string `validate:"omitempty,max=2048,url"`
}

type DeleteOrganizationReq struct {
	TenantID uint64
	ID       uint64 `validate:"required,min=1"`
}

// Organizationer defines the organization business logic.
type Organizationer interface {
	ListOrganizations(ctx context.Context, req ListOrganizationsReq) ([]*model.Organization, *common_model.PaginationResult, error)
	GetOrganization(ctx context.Context, req GetOrganizationReq) (*model.Organization, error)
	GetOrganizationLogo(ctx context.Context, req GetOrganizationLogoReq) ([]byte, *string, error)
	CreateOrganization(ctx context.Context, req CreateOrganizationReq) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, req UpdateOrganizationReq) (*model.Organization, error)
	DeleteOrganization(ctx context.Context, req DeleteOrganizationReq) error
}

// OrganizationStore defines the organization persistence layer.
type OrganizationStore interface {
	ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error)
	GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error)
	GetOrganizationLogo(ctx context.Context, tenantID, id uint64) ([]byte, *string, error)
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

func (s *OrganizationService) ListOrganizations(ctx context.Context, req ListOrganizationsReq) ([]*model.Organization, *common_model.PaginationResult, error) {
	organizations, paginationRes, err := s.store.ListOrganizations(ctx, req.TenantID, req.Pagination)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, "Unable to list organizations")
	}

	return organizations, paginationRes, nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, req GetOrganizationReq) (*model.Organization, error) {
	organization, err := s.store.GetOrganization(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get organization %v", req.ID))
	}

	return organization, nil
}

func (s *OrganizationService) GetOrganizationLogo(ctx context.Context, req GetOrganizationLogoReq) ([]byte, *string, error) {
	logo, contentType, err := s.store.GetOrganizationLogo(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get logo for organization %v", req.ID))
	}

	return logo, contentType, nil
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req CreateOrganizationReq) (*model.Organization, error) {
	organization := &model.Organization{
		TenantID:        req.TenantID,
		Name:            req.Name,
		Description:     req.Description,
		Logo:            req.Logo,
		LogoContentType: req.LogoContentType,
		Country:         req.Country,
		City:            req.City,
		ContactUserID:   req.ContactUserID,
		WebsiteURL:      req.WebsiteURL,
	}

	createdOrganization, err := s.store.CreateOrganization(ctx, req.TenantID, organization)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to create organization")
	}

	return createdOrganization, nil
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, req UpdateOrganizationReq) (*model.Organization, error) {
	organization := &model.Organization{
		ID:              req.ID,
		TenantID:        req.TenantID,
		Name:            req.Name,
		Description:     req.Description,
		Logo:            req.Logo,
		LogoContentType: req.LogoContentType,
		Country:         req.Country,
		City:            req.City,
		ContactUserID:   req.ContactUserID,
		WebsiteURL:      req.WebsiteURL,
	}

	updatedOrganization, err := s.store.UpdateOrganization(ctx, req.TenantID, organization)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to update organization %v", req.ID))
	}

	return updatedOrganization, nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, req DeleteOrganizationReq) error {
	if err := s.store.DeleteOrganization(ctx, req.TenantID, req.ID); err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to delete organization %v", req.ID))
	}

	return nil
}
