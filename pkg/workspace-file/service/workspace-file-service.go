package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/blockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type WorkspaceFiler interface {
	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*blockstore.File, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*blockstore.File, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *blockstore.File) (*blockstore.File, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *blockstore.File) (*blockstore.File, error)
	DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error
	InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *blockstore.File) (*blockstore.FileUploadInfo, error)
	UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *blockstore.FilePart) (*blockstore.FilePart, error)
	CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*blockstore.FilePart) (*blockstore.File, error)
	AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error
}

type WorkspaceFileStore interface {
	StatFile(ctx context.Context, path string) (*blockstore.File, error)
	GetFile(ctx context.Context, path string) (*blockstore.File, error)
	ListFiles(ctx context.Context, path string) ([]*blockstore.File, error)
	CreateFile(ctx context.Context, file *blockstore.File) (*blockstore.File, error)
	CreateDirectory(ctx context.Context, file *blockstore.File) (*blockstore.File, error)
	MoveFile(ctx context.Context, oldPath string, newPath string) (*blockstore.File, error)
	DeleteFile(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, dirPath string) error
	InitiateMultipartUpload(ctx context.Context, file *blockstore.File) (*blockstore.FileUploadInfo, error)
	UploadPart(ctx context.Context, path string, uploadID string, part *blockstore.FilePart) (*blockstore.FilePart, error)
	CompleteMultipartUpload(ctx context.Context, path string, uploadID string, parts []*blockstore.FilePart) (*blockstore.File, error)
	AbortMultipartUpload(ctx context.Context, path string, uploadID string) error
}

type WorkspaceFileService struct {
	fileStores   map[string]WorkspaceFileStore
	storeConfigs map[string]config.WorkspaceFileStore
}

func NewWorkspaceFileService(fileStores map[string]WorkspaceFileStore, fileStoreConfigs map[string]config.WorkspaceFileStore) (*WorkspaceFileService, error) {
	// Validate store prefixes uniqueness
	for storeName, storeCfg := range fileStoreConfigs {
		for otherStoreName, otherStoreCfg := range fileStoreConfigs {
			trimmedPrefix := strings.Trim(storeCfg.StorePrefix, "/")
			otherTrimmedPrefix := strings.Trim(otherStoreCfg.StorePrefix, "/")
			if storeName != otherStoreName && strings.HasPrefix(trimmedPrefix, otherTrimmedPrefix) {
				return nil, fmt.Errorf("workspace file store prefix conflict: store %s prefix %s overlaps with store %s prefix %s", storeName, storeCfg.StorePrefix, otherStoreName, otherStoreCfg.StorePrefix)
			}
		}
	}

	// Normalize store prefixes
	for storeName, storeCfg := range fileStoreConfigs {
		storeCfg.StorePrefix = "/" + strings.Trim(storeCfg.StorePrefix, "/") + "/"
		fileStoreConfigs[storeName] = storeCfg
	}

	ws := &WorkspaceFileService{
		fileStores:   fileStores,
		storeConfigs: fileStoreConfigs,
	}

	return ws, nil
}

func (s *WorkspaceFileService) toStorePath(storeName string, workspaceID uint64, filePath string) string {
	storeCfg := s.storeConfigs[storeName]
	normalizedPath := "/" + strings.TrimPrefix(filePath, "/")                                                   // Normalize user path
	relPath := strings.TrimPrefix(normalizedPath, storeCfg.StorePrefix)                                         // Strip store prefix to get relative path
	workspaceDir := fmt.Sprintf(storeCfg.WorkspacePrefix, workspace_model.GetWorkspaceClusterName(workspaceID)) // Format workspace prefix with workspace ID
	objectKey := fmt.Sprintf("%s/%s", workspaceDir, strings.TrimPrefix(relPath, "/"))                           // Combine workspace_prefix and relative path
	return objectKey
}

func (s *WorkspaceFileService) fromStorePath(storeName string, workspaceID uint64, storePath string) string {
	storeCfg := s.storeConfigs[storeName]
	workspaceDir := fmt.Sprintf(storeCfg.WorkspacePrefix, workspace_model.GetWorkspaceClusterName(workspaceID)) // Format workspace prefix with workspace ID
	relPath := strings.TrimPrefix(storePath, workspaceDir+"/")                                                  // Strip workspace prefix to get relative path
	userPath := storeCfg.StorePrefix + strings.TrimPrefix(relPath, "/")                                         // Prepend store prefix to get user path
	return userPath
}

func (s *WorkspaceFileService) selectFileStore(filePath string) (string, error) {
	var selectedStoreName string
	for storeName, storeCfg := range s.storeConfigs {
		normalizedPath := "/" + strings.TrimPrefix(filePath, "/")
		if strings.HasPrefix(normalizedPath, storeCfg.StorePrefix) {
			selectedStoreName = storeName
			break
		}
	}

	if selectedStoreName == "" {
		return "", fmt.Errorf("no suitable file store found for path %s", filePath)
	}

	return selectedStoreName, nil
}

func (s *WorkspaceFileService) listStores() []*blockstore.File {
	var stores []*blockstore.File
	for storeName, storeCfg := range s.storeConfigs {
		stores = append(stores, &blockstore.File{
			Path:        storeCfg.StorePrefix,
			Name:        storeName,
			IsDirectory: true,
		})
	}
	return stores
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*blockstore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := s.toStorePath(storeName, workspaceID, filePath)

	// For now, only return object Metadata, not the content
	file, err := s.fileStores[storeName].StatFile(ctx, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace file at path %s: %w", filePath, err)
	}

	return file, nil
}

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*blockstore.File, error) {
	if filePath == "" || filePath == "/" {
		return s.listStores(), nil
	}

	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := s.toStorePath(storeName, workspaceID, filePath)
	storeFiles, err := s.fileStores[storeName].ListFiles(ctx, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace files at path %s: %w", filePath, err)
	}

	var files []*blockstore.File
	for _, f := range storeFiles {
		files = append(files, &blockstore.File{
			Path:        s.fromStorePath(storeName, workspaceID, f.Path),
			Name:        f.Name,
			IsDirectory: f.IsDirectory,
			Size:        f.Size,
			MimeType:    f.MimeType,
			UpdatedAt:   f.UpdatedAt,
		})
	}

	return files, nil
}

func (s *WorkspaceFileService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *blockstore.File) (*blockstore.File, error) {
	storeName, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := s.toStorePath(storeName, workspaceID, file.Path)

	var createdFile *blockstore.File
	if file.IsDirectory {
		createdFile, err = s.fileStores[storeName].CreateDirectory(ctx, &blockstore.File{
			Path:        storePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create workspace directory at path %s: %w", file.Path, err)
		}
	} else {
		createdFile, err = s.fileStores[storeName].CreateFile(ctx, &blockstore.File{
			Path:        storePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			MimeType:    file.MimeType,
			Content:     file.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
		}
	}

	return &blockstore.File{
		Path:        s.fromStorePath(storeName, workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *blockstore.File) (*blockstore.File, error) {
	oldStoreName, err := s.selectFileStore(oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for old path: %w", err)
	}

	newStoreName, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for new path: %w", err)
	}

	oldStore := s.fileStores[oldStoreName]

	// Check if old file exists
	oldStorePath := s.toStorePath(oldStoreName, workspaceID, oldPath)
	oldFile, err := oldStore.GetFile(ctx, oldStorePath)
	if err != nil {
		return nil, fmt.Errorf("workspace file at path %s does not exist: %w", oldPath, err)
	}

	if oldStoreName != newStoreName {
		newStorePath := s.toStorePath(newStoreName, workspaceID, file.Path)

		// Cross-store move
		createdFile, err := s.fileStores[newStoreName].CreateFile(ctx, &blockstore.File{
			Path:        newStorePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			MimeType:    file.MimeType,
			Content:     oldFile.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create new workspace file at path %s: %w", file.Path, err)
		}

		err = s.fileStores[oldStoreName].DeleteFile(ctx, oldStorePath)
		if err != nil {
			_ = s.fileStores[newStoreName].DeleteFile(ctx, newStorePath)
			return nil, fmt.Errorf("unable to delete old workspace file at path %s: %w", oldPath, err)
		}

		return &blockstore.File{
			Path:        s.fromStorePath(newStoreName, workspaceID, createdFile.Path),
			Name:        createdFile.Name,
			IsDirectory: createdFile.IsDirectory,
			Size:        createdFile.Size,
			MimeType:    createdFile.MimeType,
			UpdatedAt:   createdFile.UpdatedAt,
		}, nil
	}

	// Same store move
	newStorePath := s.toStorePath(newStoreName, workspaceID, file.Path)
	updatedFile, err := s.fileStores[oldStoreName].MoveFile(ctx, oldStorePath, newStorePath)
	if err != nil {
		return nil, fmt.Errorf("unable to update old workspace file at path %s: %w", oldPath, err)
	}

	return updatedFile, nil
}

func (s *WorkspaceFileService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return fmt.Errorf("unable to select file store: %w", err)
	}

	store := s.fileStores[storeName]

	storePath := s.toStorePath(storeName, workspaceID, filePath)
	if strings.HasSuffix(storePath, "/") {
		err = store.DeleteDirectory(ctx, storePath)
		if err != nil {
			return fmt.Errorf("unable to delete workspace directory at path %s: %w", filePath, err)
		}
	} else {
		err = store.DeleteFile(ctx, storePath)
		if err != nil {
			return fmt.Errorf("unable to delete workspace file at path %s: %w", filePath, err)
		}
	}

	return nil
}

func (s *WorkspaceFileService) InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *blockstore.File) (*blockstore.FileUploadInfo, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	store := s.fileStores[storeName]
	storePath := s.toStorePath(storeName, workspaceID, file.Path)

	uploadInfo, err := store.InitiateMultipartUpload(ctx, &blockstore.File{
		Path:        storePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		MimeType:    file.MimeType,
		Size:        file.Size,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to initiate multipart upload for file at path %s: %w", file.Path, err)
	}

	return uploadInfo, nil
}

func (s *WorkspaceFileService) UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *blockstore.FilePart) (*blockstore.FilePart, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	store := s.fileStores[storeName]
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	uploadedPart, err := store.UploadPart(ctx, storePath, uploadID, part)
	if err != nil {
		return nil, fmt.Errorf("unable to upload part number %d for upload ID %s at path %s: %w", part.PartNumber, uploadID, filePath, err)
	}

	return uploadedPart, nil
}

func (s *WorkspaceFileService) CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*blockstore.FilePart) (*blockstore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	store := s.fileStores[storeName]
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	completedFile, err := store.CompleteMultipartUpload(ctx, storePath, uploadID, parts)
	if err != nil {
		return nil, fmt.Errorf("unable to complete multipart upload for upload ID %s at path %s: %w", uploadID, filePath, err)
	}

	return completedFile, nil
}

func (s *WorkspaceFileService) AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return fmt.Errorf("unable to select file store: %w", err)
	}

	store := s.fileStores[storeName]
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	err = store.AbortMultipartUpload(ctx, storePath, uploadID)
	if err != nil {
		return fmt.Errorf("unable to abort multipart upload for upload ID %s at path %s: %w", uploadID, filePath, err)
	}

	return nil
}
