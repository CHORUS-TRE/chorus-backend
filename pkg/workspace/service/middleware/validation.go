package middleware

import (
	"context"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Workspaceer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Workspaceer) service.Workspaceer {
	return func(next service.Workspaceer) service.Workspaceer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, nil, err
	}
	return v.next.ListWorkspaces(ctx, tenantID, pagination, filter)
}

func (v validation) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	return v.next.GetWorkspace(ctx, tenantID, workspaceID)
}

func (v validation) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	return v.next.DeleteWorkspace(ctx, tenantID, workspaceID)
}

func (v validation) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	if err := v.validate.Struct(workspace); err != nil {
		return nil, err
	}
	return v.next.UpdateWorkspace(ctx, workspace)
}

func (v validation) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	if err := v.validate.Struct(workspace); err != nil {
		return nil, err
	}
	return v.next.CreateWorkspace(ctx, workspace)
}

func (v validation) ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	return v.next.ManageUserRoleInWorkspace(ctx, tenantID, userID, role)
}

func (v validation) RemoveUserRoleInWorkspace(ctx context.Context, tenantID, userID, workspaceID uint64, roleName authorization_model.RoleName) error {
	return v.next.RemoveUserRoleInWorkspace(ctx, tenantID, userID, workspaceID, roleName)
}

func (v validation) RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error {
	return v.next.RemoveUserFromWorkspace(ctx, tenantID, userID, workspaceID)
}

func (v validation) GetWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) (*model.WorkspaceSvc, error) {
	return v.next.GetWorkspaceSvc(ctx, tenantID, workspaceSvcID)
}

func (v validation) ListWorkspaceSvcs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter service.WorkspaceSvcFilter) ([]*model.WorkspaceSvc, *common_model.PaginationResult, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, nil, err
	}
	return v.next.ListWorkspaceSvcs(ctx, tenantID, pagination, filter)
}

func (v validation) CreateWorkspaceSvc(ctx context.Context, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	return v.next.CreateWorkspaceSvc(ctx, svc)
}

func (v validation) UpdateWorkspaceSvc(ctx context.Context, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	return v.next.UpdateWorkspaceSvc(ctx, svc)
}

func (v validation) DeleteWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) error {
	return v.next.DeleteWorkspaceSvc(ctx, tenantID, workspaceSvcID)
}
