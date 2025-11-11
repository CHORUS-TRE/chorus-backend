package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.WorkspaceFiler
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.WorkspaceFiler) service.WorkspaceFiler {
	return func(next service.WorkspaceFiler) service.WorkspaceFiler {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	return v.next.GetWorkspaceFile(ctx, workspaceID, filePath)
}

func (v validation) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	return v.next.ListWorkspaceFiles(ctx, workspaceID, filePath)
}

func (v validation) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.CreateWorkspaceFile(ctx, workspaceID, file)
}

func (v validation) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.UpdateWorkspaceFile(ctx, workspaceID, oldPath, file)
}

func (v validation) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return v.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
}
