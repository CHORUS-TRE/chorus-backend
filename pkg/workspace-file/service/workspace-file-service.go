package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
)

type WorkspaceFiler interface {
	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error
}

type WorkspaceFileStore interface {
	// GetStoreName returns a unique identifier for the file store.
	GetStoreName() string
	// GetStorePrefix returns the path prefix associated with the file store.
	GetStorePrefix() string

	// Path normalization
	NormalizePath(path string) string
	ToStorePath(workspaceID uint64, path string) string
	FromStorePath(workspaceID uint64, storePath string) string

	// Workspace file operations
	GetFileMetadata(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error)
	GetFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error)
	ListFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error)
	CreateFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	UpdateFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	DeleteFile(ctx context.Context, workspaceID uint64, filePath string) error
}

type WorkspaceFileService struct {
	fileStores map[string]WorkspaceFileStore
}

func NewWorkspaceFileService(fileStores map[string]WorkspaceFileStore) *WorkspaceFileService {
	ws := &WorkspaceFileService{
		fileStores: fileStores,
	}

	return ws
}

func (s *WorkspaceFileService) selectFileStore(filePath string) (WorkspaceFileStore, error) {
	var selectedStore WorkspaceFileStore
	maxPrefixLen := 0

	for _, store := range s.fileStores {
		normalizedPath := store.NormalizePath(filePath)
		prefix := store.GetStorePrefix()

		if strings.HasPrefix(normalizedPath, prefix) && len(prefix) > maxPrefixLen {
			maxPrefixLen = len(prefix)
			selectedStore = store
		}
	}

	if selectedStore == nil {
		return nil, fmt.Errorf("no suitable file store found for path %s", filePath)
	}

	return selectedStore, nil
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	store, err := s.selectFileStore(filePath)
	storePath := store.ToStorePath(workspaceID, filePath)

	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	// For now, only return object Metadata, not the content
	file, err := store.GetFileMetadata(ctx, workspaceID, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace file at path %s: %w", filePath, err)
	}

	return file, nil
}

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	if filePath == "" || filePath == "/" {
		var stores []*model.WorkspaceFile
		for _, store := range s.fileStores {
			stores = append(stores, &model.WorkspaceFile{
				Path:        store.GetStorePrefix(),
				Name:        store.GetStoreName(),
				IsDirectory: true,
			})
		}
		return stores, nil
	}

	store, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := store.ToStorePath(workspaceID, filePath)
	storeFiles, err := store.ListFiles(ctx, workspaceID, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace files at path %s: %w", filePath, err)
	}

	var files []*model.WorkspaceFile
	for _, f := range storeFiles {
		files = append(files, &model.WorkspaceFile{
			Path:        store.FromStorePath(workspaceID, f.Path),
			Name:        f.Name,
			IsDirectory: f.IsDirectory,
			Size:        f.Size,
			MimeType:    f.MimeType,
			UpdatedAt:   f.UpdatedAt,
		})
	}

	return files, nil
}

func (s *WorkspaceFileService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	store, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := store.ToStorePath(workspaceID, file.Path)
	createdFile, err := store.CreateFile(ctx, workspaceID, &model.WorkspaceFile{
		Path:        storePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		MimeType:    file.MimeType,
		Content:     file.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace file at path %s: %w", file.Path, err)
	}

	return &model.WorkspaceFile{
		Path:        store.FromStorePath(workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	oldStore, err := s.selectFileStore(oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for old path: %w", err)
	}

	newStore, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for new path: %w", err)
	}

	oldStorePath := oldStore.ToStorePath(workspaceID, oldPath)
	oldFile, err := oldStore.GetFile(ctx, workspaceID, oldStorePath)
	if err != nil {
		return nil, fmt.Errorf("workspace file at path %s does not exist: %w", oldPath, err)
	}

	if oldStore.GetStoreName() != newStore.GetStoreName() {
		newStorePath := newStore.ToStorePath(workspaceID, file.Path)

		// Cross-store move
		createdFile, err := newStore.CreateFile(ctx, workspaceID, &model.WorkspaceFile{
			Path:        newStorePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			MimeType:    file.MimeType,
			Content:     oldFile.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create new workspace file at path %s: %w", file.Path, err)
		}

		err = oldStore.DeleteFile(ctx, workspaceID, oldStorePath)
		if err != nil {
			_ = newStore.DeleteFile(ctx, workspaceID, newStorePath)
			return nil, fmt.Errorf("unable to delete old workspace file at path %s: %w", oldPath, err)
		}

		return &model.WorkspaceFile{
			Path:        newStore.FromStorePath(workspaceID, createdFile.Path),
			Name:        createdFile.Name,
			IsDirectory: createdFile.IsDirectory,
			Size:        createdFile.Size,
			MimeType:    createdFile.MimeType,
			UpdatedAt:   createdFile.UpdatedAt,
		}, nil
	}

	// Same store move
	updatedFile, err := oldStore.UpdateFile(ctx, workspaceID, oldStorePath, &model.WorkspaceFile{
		Path:        oldStorePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		Size:        file.Size,
		MimeType:    file.MimeType,
		UpdatedAt:   file.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update old workspace file at path %s: %w", oldPath, err)
	}

	return updatedFile, nil
}

func (s *WorkspaceFileService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	store, err := s.selectFileStore(filePath)
	if err != nil {
		return fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := store.ToStorePath(workspaceID, filePath)
	_, stat := store.GetFileMetadata(ctx, workspaceID, storePath)
	if stat != nil {
		return fmt.Errorf("workspace file at path %s does not exist: %w", filePath, stat)
	}

	err = store.DeleteFile(ctx, workspaceID, storePath)
	if err != nil {
		return fmt.Errorf("unable to delete workspace file at path %s: %w", filePath, err)
	}

	return nil
}
