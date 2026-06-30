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

		Status: WorkspaceStatusToBusiness(workspace.Status),
		IsMain: workspace.IsMain,

		NetworkPolicy:        model.NetworkPolicyMode(workspace.NetworkPolicy),
		NetworkPolicyStatus:  workspace.NetworkPolicyStatus,
		NetworkPolicyMessage: workspace.NetworkPolicyMessage,
		AllowedFQDNs:         model.StringSlice(workspace.AllowedFqdns),
		Clipboard:            model.ClipboardMode(workspace.Clipboard),

		Visibility:    WorkspaceVisibilityToBusiness(workspace.Visibility),
		ContactUserID: nonZeroUint64(workspace.ContactUserId),

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

		Status: WorkspaceStatusFromBusiness(workspace.Status),

		IsMain: workspace.IsMain,

		Namespace: workspace.GetClusterName(),

		NetworkPolicy:        string(workspace.NetworkPolicy),
		AllowedFqdns:         []string(workspace.AllowedFQDNs),
		NetworkPolicyStatus:  workspace.NetworkPolicyStatus,
		NetworkPolicyMessage: workspace.NetworkPolicyMessage,
		Clipboard:            string(workspace.Clipboard),

		CreatedAt: ca,
		UpdatedAt: ua,

		Gid:           workspace.ID + gidOffset,
		Visibility:    WorkspaceVisibilityFromBusiness(workspace.Visibility),
		ContactUserId: workspace.GetContactUserID(),
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
		ShortName:   workspace.ShortName,
		Description: workspace.Description,
		Status:      WorkspaceStatusFromBusiness(workspace.Status),

		ContactUsername:  workspace.ContactUsername,
		ContactFirstName: workspace.ContactFirstName,
		ContactLastName:  workspace.ContactLastName,
		ContactEmail:     workspace.ContactEmail,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkspaceStatusToBusiness(s chorus.WorkspaceStatus) model.WorkspaceStatus {
	switch s {
	case chorus.WorkspaceStatus_WORKSPACE_STATUS_ACTIVE:
		return model.WorkspaceStatusActive
	case chorus.WorkspaceStatus_WORKSPACE_STATUS_INACTIVE:
		return model.WorkspaceStatusInactive
	case chorus.WorkspaceStatus_WORKSPACE_STATUS_DELETED:
		return model.WorkspaceStatusDeleted
	default:
		return model.WorkspaceStatusActive
	}
}

func WorkspaceStatusFromBusiness(s model.WorkspaceStatus) chorus.WorkspaceStatus {
	switch s {
	case model.WorkspaceStatusActive:
		return chorus.WorkspaceStatus_WORKSPACE_STATUS_ACTIVE
	case model.WorkspaceStatusInactive:
		return chorus.WorkspaceStatus_WORKSPACE_STATUS_INACTIVE
	case model.WorkspaceStatusDeleted:
		return chorus.WorkspaceStatus_WORKSPACE_STATUS_DELETED
	default:
		return chorus.WorkspaceStatus_WORKSPACE_STATUS_ACTIVE
	}
}

func WorkspaceVisibilityToBusiness(v chorus.WorkspaceVisibility) model.WorkspaceVisibility {
	switch v {
	case chorus.WorkspaceVisibility_WORKSPACE_VISIBILITY_PUBLIC:
		return model.WorkspaceVisibilityPublic
	default:
		return model.WorkspaceVisibilityPrivate
	}
}

func WorkspaceVisibilityFromBusiness(v model.WorkspaceVisibility) chorus.WorkspaceVisibility {
	switch v {
	case model.WorkspaceVisibilityPublic:
		return chorus.WorkspaceVisibility_WORKSPACE_VISIBILITY_PUBLIC
	default:
		return chorus.WorkspaceVisibility_WORKSPACE_VISIBILITY_PRIVATE
	}
}
