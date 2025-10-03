package service

import (
	"context"
	"fmt"

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
	createdFile, err := s.minioClient.PutWorkspaceObject(workspaceID, file.Path, file.Content, file.MimeType)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
	}

	return createdFile, nil
}

func (s *WorkspaceService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	_, stat := s.minioClient.StatWorkspaceObject(workspaceID, oldPath)
	if stat != nil {
		return nil, fmt.Errorf("workspace file at path %s does not exist: %w", oldPath, stat)
	}

	err := s.minioClient.DeleteWorkspaceObject(workspaceID, oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to update old workspace file at path %s: %w", oldPath, err)
	}

	file, err = s.minioClient.PutWorkspaceObject(workspaceID, file.Path, file.Content, file.MimeType)
	if err != nil {
		return nil, fmt.Errorf("unable to create new workspace file at path %s: %w", file.Path, err)
	}

	return file, nil
}

func (s *WorkspaceService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	_, stat := s.minioClient.StatWorkspaceObject(workspaceID, filePath)
	if stat != nil {
		return fmt.Errorf("workspace file at path %s does not exist: %w", filePath, stat)
	}

	err := s.minioClient.DeleteWorkspaceObject(workspaceID, filePath)
	if err != nil {
		return fmt.Errorf("unable to delete workspace file at path %s: %w", filePath, err)
	}

	return nil
}
