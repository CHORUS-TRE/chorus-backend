package service

import (
	"fmt"
	"regexp"
	"strings"

	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
)

var _ WorkspaceFileStorePathManager = &MinioFileStoragePathManager{}

type MinioFileStoragePathManager struct {
	storeName   string
	storePrefix string
	minioClient miniorawclient.MinioClienter
}

const workspacePrefix = "workspaces/workspace"
const workspacePrefixPattern = `^` + workspacePrefix + `\d+/`

func NewMinioFileStoragePathManager(clientName string, client miniorawclient.MinioClienter, clientPrefix string) (*MinioFileStoragePathManager, error) {
	return &MinioFileStoragePathManager{
		storeName:   clientName,
		storePrefix: clientPrefix,
		minioClient: client,
	}, nil
}

func (s *MinioFileStoragePathManager) GetStoreName() string {
	return s.storeName
}

func (s *MinioFileStoragePathManager) GetStorePrefix() string {
	return s.storePrefix
}

func (s *MinioFileStoragePathManager) NormalizePath(path string) string {
	return "/" + strings.TrimPrefix(path, "/")
}

func (s *MinioFileStoragePathManager) ToStorePath(workspaceID uint64, path string) string {
	normalized := s.NormalizePath(path)
	storePath := strings.TrimPrefix(normalized, s.storePrefix)
	objectKey := fmt.Sprintf("%s%d/%s", workspacePrefix, workspaceID, strings.TrimPrefix(storePath, "/"))
	return objectKey
}

func (s *MinioFileStoragePathManager) FromStorePath(workspaceID uint64, storePath string) string {
	pattern := regexp.MustCompile(workspacePrefixPattern)
	objectKey := pattern.ReplaceAllString(storePath, "")
	return s.storePrefix + strings.TrimPrefix(objectKey, "/")
}
