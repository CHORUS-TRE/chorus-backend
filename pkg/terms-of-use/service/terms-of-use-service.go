package service

import (
	"context"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
)

var _ TermsOfUseer = (*TermsOfUseService)(nil)

type TermsOfUseer interface {
	CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error)
	UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error)
	PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error)
	GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error)
	ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error)
	GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error)
	ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error)
	GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error)
	AcceptTermsOfUse(ctx context.Context, tenantID, userID uint64) (*model.TermsOfUseAcceptance, error)
}

type TermsOfUseStore interface {
	CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error)
	UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error)
	PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error)
	GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error)
	ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error)
	GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error)
	ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error)
	GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error)
	AcceptTermsOfUse(ctx context.Context, tenantID, userID, versionID uint64) (*model.TermsOfUseAcceptance, error)
}

type TermsOfUseService struct {
	store TermsOfUseStore
}

func NewTermsOfUseService(store TermsOfUseStore) *TermsOfUseService {
	return &TermsOfUseService{store: store}
}

func (s *TermsOfUseService) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	version, err := s.store.CreateTermsOfUseVersion(ctx, tenantID, content)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to create terms of use version")
	}
	return version, nil
}

func (s *TermsOfUseService) UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error) {
	version, err := s.store.UpdateTermsOfUseVersion(ctx, tenantID, versionID, content)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to update terms of use version")
	}
	return version, nil
}

func (s *TermsOfUseService) PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	version, err := s.store.PublishTermsOfUseVersion(ctx, tenantID, versionID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to publish terms of use version")
	}
	return version, nil
}

func (s *TermsOfUseService) GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	version, err := s.store.GetTermsOfUseVersion(ctx, tenantID, versionID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to get terms of use version")
	}
	return version, nil
}

func (s *TermsOfUseService) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	versions, paginationResult, err := s.store.ListTermsOfUseVersions(ctx, tenantID, pagination)
	if err != nil {
		return nil, nil, cerr.ErrInternal.Wrap(err, "Unable to list terms of use versions")
	}
	return versions, paginationResult, nil
}

func (s *TermsOfUseService) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error) {
	version, err := s.store.GetCurrentTermsOfUseVersion(ctx, tenantID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to get current terms of use version")
	}
	return version, nil
}

func (s *TermsOfUseService) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	acceptances, err := s.store.ListTermsOfUseAcceptances(ctx, tenantID, userID)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to list terms of use acceptances")
	}
	return acceptances, nil
}

func (s *TermsOfUseService) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error) {
	accepted, err := s.store.GetMyTermsOfUseStatus(ctx, tenantID, userID)
	if err != nil {
		return false, cerr.ErrInternal.Wrap(err, "Unable to get terms of use status")
	}
	return accepted, nil
}

func (s *TermsOfUseService) AcceptTermsOfUse(ctx context.Context, tenantID, userID uint64) (*model.TermsOfUseAcceptance, error) {
	current, err := s.store.GetCurrentTermsOfUseVersion(ctx, tenantID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to get current terms of use version")
	}

	acceptance, err := s.store.AcceptTermsOfUse(ctx, tenantID, userID, current.ID)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to accept terms of use")
	}

	return acceptance, nil
}
