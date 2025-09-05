package middleware

import (
	"context"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
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

func (v validation) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, nil, err
	}
	return v.next.ListWorkspaces(ctx, tenantID, pagination)
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

func (v validation) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	if err := v.validate.Var(filePath, "required"); err != nil {
		return nil, err
	}
	return v.next.GetWorkspaceFile(ctx, workspaceID, filePath)
}

func (v validation) GetWorkspaceFileChildren(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	return v.next.GetWorkspaceFileChildren(ctx, workspaceID, filePath)
}

func (v validation) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.CreateWorkspaceFile(ctx, workspaceID, file)
}

func (v validation) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.UpdateWorkspaceFile(ctx, workspaceID, file)
}

func (v validation) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	if err := v.validate.Var(filePath, "required"); err != nil {
		return err
	}
	return v.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
}
