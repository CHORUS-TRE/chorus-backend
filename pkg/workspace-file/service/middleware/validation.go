package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
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

func (v validation) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error) {
	return v.next.GetWorkspaceFile(ctx, workspaceID, filePath)
}

func (v validation) GetWorkspaceFileWithContent(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error) {
	return v.next.GetWorkspaceFileWithContent(ctx, workspaceID, filePath)
}

func (v validation) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*filestore.File, error) {
	return v.next.ListWorkspaceFiles(ctx, workspaceID, filePath)
}

func (v validation) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *filestore.File) (*filestore.File, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.CreateWorkspaceFile(ctx, workspaceID, file)
}

func (v validation) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *filestore.File) (*filestore.File, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.UpdateWorkspaceFile(ctx, workspaceID, oldPath, file)
}

func (v validation) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return v.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
}

func (v validation) InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *filestore.File) (*filestore.FileUploadInfo, error) {
	if err := v.validate.Struct(file); err != nil {
		return nil, err
	}
	return v.next.InitiateWorkspaceFileUpload(ctx, workspaceID, filePath, file)
}

func (v validation) UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *filestore.FilePart) (*filestore.FilePart, error) {
	if err := v.validate.Struct(part); err != nil {
		return nil, err
	}
	return v.next.UploadWorkspaceFilePart(ctx, workspaceID, filePath, uploadID, part)
}

func (v validation) CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*filestore.FilePart) (*filestore.File, error) {
	for _, part := range parts {
		if err := v.validate.Struct(part); err != nil {
			return nil, err
		}
	}
	return v.next.CompleteWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID, parts)
}

func (v validation) AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error {
	return v.next.AbortWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID)
}
