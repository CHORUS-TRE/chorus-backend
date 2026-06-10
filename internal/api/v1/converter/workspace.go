package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

func WorkspaceToBusiness(workspace *chorus.Workspace) (*model.Workspace, error) {
	ca, err := FromProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &model.Workspace{
		ID: workspace.Id,

		TenantID: workspace.TenantId,
		UserID:   workspace.UserId,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		IsMain: workspace.IsMain,

		NetworkPolicy: model.NetworkPolicyMode(workspace.NetworkPolicy),
		AllowedFQDNs:  model.StringSlice(workspace.AllowedFqdns),
		Clipboard:     model.ClipboardMode(workspace.Clipboard),

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkspaceFromBusiness(workspace *model.Workspace, gidOffset uint64) (*chorus.Workspace, error) {
	ca, err := ToProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.Workspace{
		Id: workspace.ID,

		TenantId: workspace.TenantID,
		UserId:   workspace.UserID,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		Status: workspace.Status.String(),

		IsMain: workspace.IsMain,

		Namespace: workspace.GetClusterName(),

		NetworkPolicy:        string(workspace.NetworkPolicy),
		AllowedFqdns:         []string(workspace.AllowedFQDNs),
		NetworkPolicyStatus:  workspace.NetworkPolicyStatus,
		NetworkPolicyMessage: workspace.NetworkPolicyMessage,
		Clipboard:            string(workspace.Clipboard),

		CreatedAt: ca,
		UpdatedAt: ua,

		Gid: workspace.ID + gidOffset,
	}, nil
}

func PublicWorkspaceFromBusiness(workspace *model.PublicWorkspace, gidOffset uint64) (*chorus.PublicWorkspace, error) {
	ca, err := ToProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.PublicWorkspace{
		Id:       workspace.ID,
		TenantId: workspace.TenantID,

		Name:        workspace.Name,
		Description: workspace.Description,
		Status:      workspace.Status.String(),

		ContactUsername:  workspace.ContactUsername,
		ContactFirstName: workspace.ContactFirstName,
		ContactLastName:  workspace.ContactLastName,
		ContactEmail:     workspace.ContactEmail,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
