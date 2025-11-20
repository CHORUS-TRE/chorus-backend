package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type WorkspaceFiler interface {
	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.File, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.File, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.File) (*model.File, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.File) (*model.File, error)
	DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error
}

type WorkspaceFileStore interface {
	StatFile(ctx context.Context, path string) (*model.File, error)
	GetFile(ctx context.Context, path string) (*model.File, error)
	ListFiles(ctx context.Context, path string) ([]*model.File, error)
	CreateFile(ctx context.Context, file *model.File) (*model.File, error)
	CreateDirectory(ctx context.Context, file *model.File) (*model.File, error)
	MoveFile(ctx context.Context, oldPath string, newPath string) (*model.File, error)
	DeleteFile(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, dirPath string) error
}

type WorkspaceFileService struct {
	fileStores   map[string]WorkspaceFileStore
	storeConfigs map[string]config.WorkspaceFileStore
}

func NewWorkspaceFileService(fileStores map[string]WorkspaceFileStore, fileStoreConfigs map[string]config.WorkspaceFileStore) (*WorkspaceFileService, error) {
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

func (s *WorkspaceFileService) listStores() []*model.File {
	var stores []*model.File
	for storeName, storeCfg := range s.storeConfigs {
		stores = append(stores, &model.File{
			Path:        storeCfg.StorePrefix,
			Name:        storeName,
			IsDirectory: true,
		})
	}
	return stores
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.File, error) {
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

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.File, error) {
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

	var files []*model.File
	for _, f := range storeFiles {
		files = append(files, &model.File{
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

func (s *WorkspaceFileService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.File) (*model.File, error) {
	storeName, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := s.toStorePath(storeName, workspaceID, file.Path)
	createdFile, err := s.fileStores[storeName].CreateFile(ctx, &model.File{
		Path:        storePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		MimeType:    file.MimeType,
		Content:     file.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
	}

	return &model.File{
		Path:        s.fromStorePath(storeName, workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.File) (*model.File, error) {
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
		createdFile, err := s.fileStores[newStoreName].CreateFile(ctx, &model.File{
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

		return &model.File{
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
