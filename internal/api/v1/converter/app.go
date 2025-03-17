package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
)

func AppToBusiness(app *chorus.App) (*model.App, error) {
	ca, err := FromProtoTimestamp(app.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}
	status, err := model.ToAppStatus(app.Status)
	if err != nil {
		return nil, fmt.Errorf("unable to convert app status: %w", err)
	}

	return &model.App{
		ID: app.Id,

		TenantID: app.TenantId,
		UserID:   app.UserId,

		Name:        app.Name,
		Description: app.Description,

		Status: status,

		DockerImageName:     app.DockerImageName,
		DockerImageTag:      app.DockerImageTag,
		DockerImageRegistry: app.DockerImageRegistry,

		ShmSize:        app.ShmSize,
		KioskConfigURL: app.KioskConfigURL,
		MaxCPU:         app.MaxCPU,
		MinCPU:         app.MinCPU,
		MaxMemory:      app.MaxMemory,
		MinMemory:      app.MinMemory,
		IconURL:        app.IconURL,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func AppFromBusiness(app *model.App) (*chorus.App, error) {
	ca, err := ToProtoTimestamp(app.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.App{
		Id: app.ID,

		TenantId: app.TenantID,
		UserId:   app.UserID,

		Name:        app.Name,
		Description: app.Description,

		Status: app.Status.String(),

		DockerImageName:     app.DockerImageName,
		DockerImageTag:      app.DockerImageTag,
		DockerImageRegistry: app.DockerImageRegistry,

		ShmSize:        app.ShmSize,
		KioskConfigURL: app.KioskConfigURL,
		MaxCPU:         app.MaxCPU,
		MinCPU:         app.MinCPU,
		MaxMemory:      app.MaxMemory,
		MinMemory:      app.MinMemory,
		IconURL:        app.IconURL,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
