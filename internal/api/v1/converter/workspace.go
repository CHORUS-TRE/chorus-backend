package converter

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/pkg/errors"
)

func WorkspaceToBusiness(workspace *chorus.Workspace) (*model.Workspace, error) {
	ca, err := FromProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := FromProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}
	status, err := model.ToWorkspaceStatus(workspace.Status)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert workspace status")
	}

	return &model.Workspace{
		ID: workspace.Id,

		TenantID: workspace.TenantId,
		UserID:   workspace.UserId,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		Status: status,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkspaceFromBusiness(workspace *model.Workspace) (*chorus.Workspace, error) {
	ca, err := ToProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := ToProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}

	return &chorus.Workspace{
		Id: workspace.ID,

		TenantId: workspace.TenantID,
		UserId:   workspace.UserID,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		Status: workspace.Status.String(),

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
