package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
)

var _ WorkspaceFileStore = &MinioFileStorage{}

type MinioFileStorage struct {
	storeName   string
	storePrefix string
	minioClient minio.MinioClienter
}

func NewMinioFileStorage(clientName string, client minio.MinioClienter, clientPrefix string) (*MinioFileStorage, error) {
	return &MinioFileStorage{
		storeName:   clientName,
		storePrefix: clientPrefix,
		minioClient: client,
	}, nil
}

func (s *MinioFileStorage) GetFileMetadata(ctx context.Context, objectKey string) (*model.WorkspaceFile, error) {
	objectInfo, err := s.minioClient.StatObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	file := model.MinioObjectInfoToWorkspaceFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("retrieved metadata for %s", objectKey))
	return file, nil
}

func (s *MinioFileStorage) StatFile(ctx context.Context, path string) (*model.WorkspaceFile, error) {
	objectInfo, err := s.minioClient.StatObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", path, err)
	}

	file := model.MinioObjectInfoToWorkspaceFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("Retrieved metadata for %s", path))
	return file, nil
}

func (s *MinioFileStorage) GetFile(ctx context.Context, path string) (*model.WorkspaceFile, error) {
	object, err := s.minioClient.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get object %s: %w", path, err)
	}

	file := model.MinioObjectToWorkspaceFile(object)

	logger.TechLog.Info(ctx, fmt.Sprintf("Downloaded %s", path))
	return file, nil
}

func (s *MinioFileStorage) ListFiles(ctx context.Context, path string) ([]*model.WorkspaceFile, error) {
	objects, err := s.minioClient.ListObjects(path, false)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects with prefix %s: %w", path, err)
	}

	var files []*model.WorkspaceFile
	for _, objectInfo := range objects {
		file := model.MinioObjectInfoToWorkspaceFile(objectInfo)
		files = append(files, file)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Listed %d files from path %s", len(files), path))
	return files, nil
}

func (s *MinioFileStorage) CreateFile(ctx context.Context, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	// Check if exists
	_, err := s.minioClient.StatObject(file.Path)
	if err == nil {
		return nil, fmt.Errorf("object at %s already exists", file.Path)
	}

	// Upload
	_, err = s.minioClient.PutObject(file.Path, model.WorkspaceFileToMinioObject(file))
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", file.Path, err)
	}

	// Verify upload
	objectInfo, err := s.minioClient.StatObject(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created object: %w", err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created %s", file.Path))

	createdFile := model.MinioObjectInfoToWorkspaceFile(objectInfo)

	return createdFile, nil
}

func (s *MinioFileStorage) UpdateFile(ctx context.Context, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	// Check if old file exists
	_, err := s.minioClient.StatObject(oldPath)
	if err != nil {
		return nil, fmt.Errorf("object at %s does not exist: %w", oldPath, err)
	}

	// Upload new file
	_, err = s.minioClient.PutObject(file.Path, model.WorkspaceFileToMinioObject(file))
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", file.Path, err)
	}

	// Delete old file
	err = s.minioClient.DeleteObject(oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to delete old object at %s: %w", oldPath, err)
	}

	// Verify upload
	objectInfo, err := s.minioClient.StatObject(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to verify updated object: %w", err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Updated %s to %s", oldPath, file.Path))

	updatedFile := model.MinioObjectInfoToWorkspaceFile(objectInfo)

	return updatedFile, nil
}

func (s *MinioFileStorage) DeleteFile(ctx context.Context, path string) error {
	if path[len(path)-1] == '/' {
		return s.deleteDirectory(ctx, path)
	}

	err := s.minioClient.DeleteObject(path)
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted %s", path))
	return nil
}

func (s *MinioFileStorage) deleteDirectory(ctx context.Context, path string) error {
	objects, err := s.minioClient.ListObjects(path, true)
	if err != nil {
		return fmt.Errorf("unable to list objects with prefix %s: %w", path, err)
	}

	for _, objectInfo := range objects {
		err := s.minioClient.DeleteObject(objectInfo.Key)
		if err != nil {
			return fmt.Errorf("unable to delete object at %s: %w", objectInfo.Key, err)
		}
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted directory %s", path))
	return nil
}
