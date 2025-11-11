package minio

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	workspace_file_service "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
)

var _ workspace_file_service.WorkspaceFileStore = &MinioFileStorage{}

type MinioFileStorage struct {
	storeName   string
	storePrefix string
	minioClient minio.MinioClienter
}

const workspacePrefix = "workspaces/workspace"
const workspacePrefixPattern = `^` + workspacePrefix + `\d+/`

func NewMinioFileStorage(clientName string, client minio.MinioClienter) (*MinioFileStorage, error) {
	return &MinioFileStorage{
		storeName:   clientName,
		storePrefix: client.GetClientPrefix(),
		minioClient: client,
	}, nil
}

func (s *MinioFileStorage) GetStoreName() string {
	return s.storeName
}

func (s *MinioFileStorage) GetStorePrefix() string {
	return s.storePrefix
}

func (s *MinioFileStorage) NormalizePath(path string) string {
	return "/" + strings.TrimPrefix(path, "/")
}

func (s *MinioFileStorage) ToStorePath(workspaceID uint64, path string) string {
	normalized := s.NormalizePath(path)
	storePath := strings.TrimPrefix(normalized, s.storePrefix)
	objectKey := fmt.Sprintf("%s%d/%s", workspacePrefix, workspaceID, strings.TrimPrefix(storePath, "/"))
	return strings.TrimPrefix(s.NormalizePath(objectKey), s.storePrefix)
}

func (s *MinioFileStorage) FromStorePath(workspaceID uint64, storePath string) string {
	pattern := regexp.MustCompile(workspacePrefixPattern)
	objectKey := pattern.ReplaceAllString(storePath, "")
	return s.storePrefix + strings.TrimPrefix(objectKey, "/")
}

func (s *MinioFileStorage) GetFileMetadata(ctx context.Context, workspaceID uint64, objectKey string) (*model.WorkspaceFile, error) {
	objectInfo, err := s.minioClient.StatObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	file := minio.MinioObjectInfoToWorkspaceFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("retrieved metadata for %s from workspace %d", objectKey, workspaceID))
	return file, nil
}

func (s *MinioFileStorage) StatFile(ctx context.Context, workspaceID uint64, path string) (*model.WorkspaceFile, error) {
	objectKey := model.WorkspacePathToObjectKey(workspaceID, path)

	objectInfo, err := s.minioClient.StatObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	file := minio.MinioObjectInfoToWorkspaceFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("Retrieved metadata for %s from workspace %d", objectKey, workspaceID))
	return file, nil
}

func (s *MinioFileStorage) GetFile(ctx context.Context, workspaceID uint64, path string) (*model.WorkspaceFile, error) {
	objectKey := model.WorkspacePathToObjectKey(workspaceID, path)

	object, err := s.minioClient.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get object %s: %w", objectKey, err)
	}

	file := minio.MinioObjectToWorkspaceFile(object)

	logger.TechLog.Info(ctx, fmt.Sprintf("Downloaded %s from workspace %d", objectKey, workspaceID))
	return file, nil
}

func (s *MinioFileStorage) ListFiles(ctx context.Context, workspaceID uint64, path string) ([]*model.WorkspaceFile, error) {
	objectKey := model.WorkspacePathToObjectKey(workspaceID, path)

	objects, err := s.minioClient.ListObjects(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects with prefix %s: %w", objectKey, err)
	}

	var files []*model.WorkspaceFile
	for _, objectInfo := range objects {
		file := minio.MinioObjectInfoToWorkspaceFile(objectInfo)
		files = append(files, file)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Listed %d files from workspace %d path %s", len(files), workspaceID, path))
	return files, nil
}

func (s *MinioFileStorage) CreateFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	objectKey := model.WorkspacePathToObjectKey(workspaceID, file.Path)

	// Check if exists
	_, err := s.minioClient.StatObject(objectKey)
	if err == nil {
		return nil, fmt.Errorf("object at %s already exists in workspace %d", objectKey, workspaceID)
	}

	// Upload
	_, err = s.minioClient.PutObject(objectKey, minio.WorkspaceFileToMinioObject(file))
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", objectKey, err)
	}

	// Verify upload
	objectInfo, err := s.minioClient.StatObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created object: %w", err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created %s in workspace %d", objectKey, workspaceID))

	createdFile := minio.MinioObjectInfoToWorkspaceFile(objectInfo)

	return createdFile, nil
}

func (s *MinioFileStorage) UpdateFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	oKey := model.WorkspacePathToObjectKey(workspaceID, oldPath)
	nKey := model.WorkspacePathToObjectKey(workspaceID, file.Path)

	// Check if old file exists
	_, err := s.minioClient.StatObject(oKey)
	if err != nil {
		return nil, fmt.Errorf("object at %s does not exist in workspace %d: %w", oKey, workspaceID, err)
	}

	// Upload new file
	_, err = s.minioClient.PutObject(nKey, minio.WorkspaceFileToMinioObjectWithoutContent(file))
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", nKey, err)
	}

	// Delete old file
	err = s.minioClient.DeleteObject(oKey)
	if err != nil {
		return nil, fmt.Errorf("unable to delete old object at %s: %w", oKey, err)
	}

	// Verify upload
	objectInfo, err := s.minioClient.StatObject(nKey)
	if err != nil {
		return nil, fmt.Errorf("unable to verify updated object: %w", err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Updated %s to %s in workspace %d", oKey, nKey, workspaceID))

	updatedFile := minio.MinioObjectInfoToWorkspaceFile(objectInfo)

	return updatedFile, nil
}

func (s *MinioFileStorage) DeleteFile(ctx context.Context, workspaceID uint64, path string) error {
	objectKey := model.WorkspacePathToObjectKey(workspaceID, path)

	err := s.minioClient.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", objectKey, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted %s from workspace %d", objectKey, workspaceID))
	return nil
}
