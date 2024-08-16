package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
)

func WorkbenchToBusiness(workbench *chorus.Workbench) (*model.Workbench, error) {
	ca, err := FromProtoTimestamp(workbench.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(workbench.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}
	status, err := model.ToWorkbenchStatus(workbench.Status)
	if err != nil {
		return nil, fmt.Errorf("unable to convert workbench status: %w", err)
	}

	return &model.Workbench{
		ID: workbench.Id,

		TenantID:    workbench.TenantId,
		UserID:      workbench.UserId,
		WorkspaceID: workbench.WorkspaceId,

		Name:        workbench.Name,
		ShortName:   workbench.ShortName,
		Description: workbench.Description,

		Status: status,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkbenchFromBusiness(workbench *model.Workbench) (*chorus.Workbench, error) {
	ca, err := ToProtoTimestamp(workbench.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(workbench.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.Workbench{
		Id: workbench.ID,

		TenantId:    workbench.TenantID,
		UserId:      workbench.UserID,
		WorkspaceId: workbench.WorkspaceID,

		Name:        workbench.Name,
		ShortName:   workbench.ShortName,
		Description: workbench.Description,

		Status: workbench.Status.String(),

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
