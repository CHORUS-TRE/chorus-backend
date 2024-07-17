package converter

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/pkg/errors"
)

func AppInstanceToBusiness(appInstance *chorus.AppInstance) (*model.AppInstance, error) {
	ca, err := FromProtoTimestamp(appInstance.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := FromProtoTimestamp(appInstance.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}
	status, err := model.ToAppInstanceStatus(appInstance.Status)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert appInstance status")
	}

	return &model.AppInstance{
		ID: appInstance.Id,

		TenantID:    appInstance.TenantId,
		UserID:      appInstance.UserId,
		AppID:       appInstance.AppId,
		WorkspaceID: appInstance.WorkspaceId,
		WorkbenchID: appInstance.WorkbenchId,

		Status: status,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func AppInstanceFromBusiness(appInstance *model.AppInstance) (*chorus.AppInstance, error) {
	ca, err := ToProtoTimestamp(appInstance.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := ToProtoTimestamp(appInstance.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}

	return &chorus.AppInstance{
		Id: appInstance.ID,

		TenantId:    appInstance.TenantID,
		UserId:      appInstance.UserID,
		AppId:       appInstance.AppID,
		WorkspaceId: appInstance.WorkspaceID,
		WorkbenchId: appInstance.WorkbenchID,

		Status: appInstance.Status.String(),

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
