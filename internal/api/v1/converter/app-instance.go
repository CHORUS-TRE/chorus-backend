package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
)

func AppInstanceToBusiness(appInstance *chorus.AppInstance) (*model.AppInstance, error) {
	ca, err := FromProtoTimestamp(appInstance.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(appInstance.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}
	status, err := model.ToAppInstanceStatus(appInstance.Status)
	if err != nil {
		return nil, fmt.Errorf("unable to convert appInstance status: %w", err)
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
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(appInstance.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
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
