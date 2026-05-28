package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.PlatformSettingsServiceServer = (*platformSettingsControllerAuthorization)(nil)

type platformSettingsControllerAuthorization struct {
	Authorization
	next chorus.PlatformSettingsServiceServer
}

func PlatformSettingsAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer) func(chorus.PlatformSettingsServiceServer) chorus.PlatformSettingsServiceServer {
	return func(next chorus.PlatformSettingsServiceServer) chorus.PlatformSettingsServiceServer {
		return &platformSettingsControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c platformSettingsControllerAuthorization) GetPlatformSettings(ctx context.Context, req *chorus.GetPlatformSettingsRequest) (*chorus.GetPlatformSettingsReply, error) {
	return c.next.GetPlatformSettings(ctx, req)
}

func (c platformSettingsControllerAuthorization) UpdatePlatformSettings(ctx context.Context, req *chorus.UpdatePlatformSettingsRequest) (*chorus.UpdatePlatformSettingsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionSetPlatformSettings)
	if err != nil {
		return nil, err
	}

	return c.next.UpdatePlatformSettings(ctx, req)
}
