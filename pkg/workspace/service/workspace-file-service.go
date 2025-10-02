package service

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

func (s *WorkspaceService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	// For now, only return object Metadata, not the content
	file, err := s.minioClient.StatWorkspaceObject(workspaceID, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace file at path %s: %w", filePath, err)
	}

	return file, nil
}

func (s *WorkspaceService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	files, err := s.minioClient.ListWorkspaceObjects(workspaceID, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace files at path %s: %w", filePath, err)
	}

	return files, nil
}

func (s *WorkspaceService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	err := s.minioClient.PutWorkspaceObject(workspaceID, file.Path, file.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
	}

	result := &model.WorkspaceFile{
		Path:        file.Path,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		Size:        int64(len(file.Content)),
		MimeType:    file.MimeType,
		UpdatedAt:   time.Now(),
		Content:     file.Content,
	}

	return result, nil
}

func (s *WorkspaceService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *WorkspaceService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return fmt.Errorf("not implemented")
}
