package middleware

import (
	"context"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

type validation struct {
	next service.PlatformSettingser
}

func Validation() func(service.PlatformSettingser) service.PlatformSettingser {
	return func(next service.PlatformSettingser) service.PlatformSettingser {
		return &validation{next: next}
	}
}

func (v validation) GetPlatformSettings(ctx context.Context) (*model.PlatformSettings, error) {
	return v.next.GetPlatformSettings(ctx)
}

func (v validation) UpdatePlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	if settings == nil {
		return nil, cerr.ErrValidation.WithMessage("Platform settings are required")
	}
	return v.next.UpdatePlatformSettings(ctx, settings)
}
