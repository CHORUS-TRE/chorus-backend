package converter

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/pkg/errors"
)

func AppToBusiness(app *chorus.App) (*model.App, error) {
	ca, err := FromProtoTimestamp(app.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := FromProtoTimestamp(app.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}
	status, err := model.ToAppStatus(app.Status)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert app status")
	}

	return &model.App{
		ID: app.Id,

		TenantID: app.TenantId,
		UserID:   app.UserId,

		Name:        app.Name,
		Description: app.Description,

		Status: status,

		DockerImageName: app.DockerImageName,
		DockerImageTag:  app.DockerImageTag,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func AppFromBusiness(app *model.App) (*chorus.App, error) {
	ca, err := ToProtoTimestamp(app.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := ToProtoTimestamp(app.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}

	return &chorus.App{
		Id: app.ID,

		TenantId: app.TenantID,
		UserId:   app.UserID,

		Name:        app.Name,
		Description: app.Description,

		Status: app.Status.String(),

		DockerImageName: app.DockerImageName,
		DockerImageTag:  app.DockerImageTag,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
