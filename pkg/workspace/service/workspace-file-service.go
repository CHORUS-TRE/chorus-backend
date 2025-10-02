package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

func (s *WorkspaceService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	// For now, only return object Metadata, not the content
	file, err := s.minioClient.StatObject(toAbsolutePath(workspaceID, filePath))
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace file at path %s: %w", filePath, err)
	}

	return file, nil
}

func (s *WorkspaceService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	files, err := s.minioClient.ListObjects(toAbsolutePath(workspaceID, filePath))
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace files at path %s: %w", filePath, err)
	}

	return files, nil
}

func (s *WorkspaceService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	err := s.minioClient.PutObject(toAbsolutePath(workspaceID, file.Path), file.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
	}

	createdFile, err := s.minioClient.StatObject(toAbsolutePath(workspaceID, file.Path))
	if err != nil {
		return nil, fmt.Errorf("unable to get created workspace file at path %s: %w", file.Path, err)
	}

	return createdFile, nil
}

func (s *WorkspaceService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *WorkspaceService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return fmt.Errorf("not implemented")
}

func toAbsolutePath(workspaceID uint64, path string) string {
	return fmt.Sprintf("workspaces/workspace%d/%s", workspaceID, strings.TrimPrefix(path, "/"))
}
