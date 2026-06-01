package service

import (
	"context"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
)

var _ PlatformSettingser = (*PlatformSettingsService)(nil)

type PlatformSettingser interface {
	GetPlatformSettings(ctx context.Context) (*model.PlatformSettings, error)
	UpdatePlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error)
}

type PlatformSettingsStore interface {
	GetPlatformSettings(ctx context.Context, tenantID uint64) (*model.PlatformSettings, error)
	UpsertPlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error)
}

type PlatformSettingsService struct {
	store PlatformSettingsStore
}

func NewPlatformSettingsService(store PlatformSettingsStore) *PlatformSettingsService {
	return &PlatformSettingsService{store: store}
}

func (s *PlatformSettingsService) GetPlatformSettings(ctx context.Context) (*model.PlatformSettings, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		tenantID = 1 // public endpoint: fall back to the default tenant
	}
	settings, err := s.store.GetPlatformSettings(ctx, tenantID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to get platform settings")
	}
	return settings, nil
}

func (s *PlatformSettingsService) UpdatePlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrUnauthenticated.WithMessage("Unable to extract tenant ID")
	}
	settings.TenantID = tenantID
	result, err := s.store.UpsertPlatformSettings(ctx, settings)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to update platform settings")
	}
	return result, nil
}
