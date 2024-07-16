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

func (v validation) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, err
	}
	return v.next.ListWorkspaces(ctx, tenantID, pagination)
}

func (v validation) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	return v.next.GetWorkspace(ctx, tenantID, workspaceID)
}

func (v validation) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	return v.next.DeleteWorkspace(ctx, tenantID, workspaceID)
}

func (v validation) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) error {
	if err := v.validate.Struct(workspace); err != nil {
		return v.next.UpdateWorkspace(ctx, workspace)
	}
	return v.next.UpdateWorkspace(ctx, workspace)
}

func (v validation) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (uint64, error) {
	if err := v.validate.Struct(workspace); err != nil {
		return 0, err
	}
	return v.next.CreateWorkspace(ctx, workspace)
}
