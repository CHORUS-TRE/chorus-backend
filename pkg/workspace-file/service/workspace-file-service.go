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

type WorkspaceFileStorePathManager interface {
	// GetStoreName returns a unique identifier for the file store.
	GetStoreName() string
	// GetStorePrefix returns the path prefix associated with the file store.
	GetStorePrefix() string

	// Path normalization
	NormalizePath(path string) string
	ToStorePath(workspaceID uint64, path string) string
	FromStorePath(workspaceID uint64, storePath string) string
}

type WorkspaceFileStore interface {
	// Workspace file operations
	GetFileMetadata(ctx context.Context, filePath string) (*model.WorkspaceFile, error)
	GetFile(ctx context.Context, filePath string) (*model.WorkspaceFile, error)
	ListFiles(ctx context.Context, filePath string) ([]*model.WorkspaceFile, error)
	CreateFile(ctx context.Context, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	UpdateFile(ctx context.Context, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	DeleteFile(ctx context.Context, filePath string) error
}

type WorkspaceFileStoreCombiner interface {
	WorkspaceFileStorePathManager
	WorkspaceFileStore
}

type fileStore struct {
	WorkspaceFileStorePathManager
	WorkspaceFileStore
}

type WorkspaceFileService struct {
	fileStores map[string]fileStore
}

func NewWorkspaceFileService(fileStores map[string]WorkspaceFileStoreCombiner) *WorkspaceFileService {
	fs := make(map[string]fileStore)
	for name, store := range fileStores {
		fs[name] = fileStore{
			WorkspaceFileStorePathManager: store,
			WorkspaceFileStore:            store,
		}
	}
	ws := &WorkspaceFileService{
		fileStores: fs,
	}

	return ws
}

func (s *WorkspaceFileService) findStore(filePath string) (fileStore, error) {
	var selectedStore fileStore
	maxPrefixLen := 0

	for _, store := range s.fileStores {
		normalizedPath := store.NormalizePath(filePath)
		prefix := store.GetStorePrefix()

		if strings.HasPrefix(normalizedPath, prefix) && len(prefix) > maxPrefixLen {
			maxPrefixLen = len(prefix)
			selectedStore = store
		}
	}

	if selectedStore == (fileStore{}) {
		return fileStore{}, fmt.Errorf("no suitable file store found for path %s", filePath)
	}

	return selectedStore, nil
}
func (s *WorkspaceFileService) findFileStore(filePath string) (WorkspaceFileStore, error) {
	return s.findStore(filePath)
}

func (s *WorkspaceFileService) findFileStorePathManager(filePath string) (WorkspaceFileStorePathManager, error) {
	return s.findStore(filePath)

}

func (s *WorkspaceFileService) listStores() []*model.WorkspaceFile {
	var stores []*model.WorkspaceFile
	for _, store := range s.fileStores {
		stores = append(stores, &model.WorkspaceFile{
			Path:        store.GetStorePrefix(),
			Name:        store.GetStoreName(),
			IsDirectory: true,
		})
	}
	return stores
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	store, err := s.findFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePathManager, err := s.findFileStorePathManager(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := storePathManager.ToStorePath(workspaceID, filePath)

	// For now, only return object Metadata, not the content
	file, err := store.GetFileMetadata(ctx, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace file at path %s: %w", filePath, err)
	}

	return file, nil
}

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	if filePath == "" || filePath == "/" {
		return s.listStores(), nil
	}

	store, err := s.findFileStore(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePathManager, err := s.findFileStorePathManager(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := storePathManager.ToStorePath(workspaceID, filePath)
	storeFiles, err := store.ListFiles(ctx, storePath)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace files at path %s: %w", filePath, err)
	}

	var files []*model.WorkspaceFile
	for _, f := range storeFiles {
		files = append(files, &model.WorkspaceFile{
			Path:        storePathManager.FromStorePath(workspaceID, f.Path),
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
	store, err := s.findFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePathManager, err := s.findFileStorePathManager(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := storePathManager.ToStorePath(workspaceID, file.Path)
	createdFile, err := store.CreateFile(ctx, &model.WorkspaceFile{
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
		Path:        storePathManager.FromStorePath(workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	oldStore, err := s.findFileStore(oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for old path: %w", err)
	}

	oldStorePathManager, err := s.findFileStorePathManager(oldPath)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for old path: %w", err)
	}

	newStore, err := s.findFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for new path: %w", err)
	}

	newStorePathManager, err := s.findFileStorePathManager(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store for new path: %w", err)
	}

	oldStorePath := oldStorePathManager.ToStorePath(workspaceID, oldPath)
	oldFile, err := oldStore.GetFile(ctx, oldStorePath)
	if err != nil {
		return nil, fmt.Errorf("workspace file at path %s does not exist: %w", oldPath, err)
	}

	if oldStorePathManager.GetStoreName() != newStorePathManager.GetStoreName() {
		newStorePath := newStorePathManager.ToStorePath(workspaceID, file.Path)

		// Cross-store move
		createdFile, err := newStore.CreateFile(ctx, &model.WorkspaceFile{
			Path:        newStorePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			MimeType:    file.MimeType,
			Content:     oldFile.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create new workspace file at path %s: %w", file.Path, err)
		}

		err = oldStore.DeleteFile(ctx, oldStorePath)
		if err != nil {
			_ = newStore.DeleteFile(ctx, newStorePath)
			return nil, fmt.Errorf("unable to delete old workspace file at path %s: %w", oldPath, err)
		}

		return &model.WorkspaceFile{
			Path:        newStorePathManager.FromStorePath(workspaceID, createdFile.Path),
			Name:        createdFile.Name,
			IsDirectory: createdFile.IsDirectory,
			Size:        createdFile.Size,
			MimeType:    createdFile.MimeType,
			UpdatedAt:   createdFile.UpdatedAt,
		}, nil
	}

	// Same store move
	updatedFile, err := oldStore.UpdateFile(ctx, oldStorePath, &model.WorkspaceFile{
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
	store, err := s.findFileStore(filePath)
	if err != nil {
		return fmt.Errorf("unable to select file store: %w", err)
	}

	storePathManager, err := s.findFileStorePathManager(filePath)
	if err != nil {
		return fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := storePathManager.ToStorePath(workspaceID, filePath)
	_, stat := store.GetFileMetadata(ctx, storePath)
	if stat != nil {
		return fmt.Errorf("workspace file at path %s does not exist: %w", filePath, stat)
	}

	err = store.DeleteFile(ctx, storePath)
	if err != nil {
		return fmt.Errorf("unable to delete workspace file at path %s: %w", filePath, err)
	}

	return nil
}
