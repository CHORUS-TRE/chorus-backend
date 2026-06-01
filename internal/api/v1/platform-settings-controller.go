package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

var _ chorus.PlatformSettingsServiceServer = (*PlatformSettingsController)(nil)

type PlatformSettingsController struct {
	platformSettings service.PlatformSettingser
}

func NewPlatformSettingsController(platformSettings service.PlatformSettingser) PlatformSettingsController {
	return PlatformSettingsController{platformSettings: platformSettings}
}

func (c PlatformSettingsController) GetPlatformSettings(ctx context.Context, _ *chorus.GetPlatformSettingsRequest) (*chorus.GetPlatformSettingsReply, error) {
	settings, err := c.platformSettings.GetPlatformSettings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'GetPlatformSettings': %v", err.Error())
	}

	proto, err := converter.PlatformSettingsFromBusiness(settings)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert platform settings")
	}

	return &chorus.GetPlatformSettingsReply{Result: &chorus.GetPlatformSettingsResult{PlatformSettings: proto}}, nil
}

func (c PlatformSettingsController) UpdatePlatformSettings(ctx context.Context, req *chorus.UpdatePlatformSettingsRequest) (*chorus.UpdatePlatformSettingsReply, error) {
	if req.PlatformSettings == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("platform settings payload is required")
	}

	settings, err := c.platformSettings.UpdatePlatformSettings(ctx, converter.PlatformSettingsToBusiness(req.PlatformSettings))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'UpdatePlatformSettings': %v", err.Error())
	}

	proto, err := converter.PlatformSettingsFromBusiness(settings)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert platform settings")
	}

	return &chorus.UpdatePlatformSettingsReply{Result: &chorus.UpdatePlatformSettingsResult{PlatformSettings: proto}}, nil
}
