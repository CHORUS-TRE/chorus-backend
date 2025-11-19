package service

import (
	"context"
	"fmt"
	"strings"

	minio "github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
)

type WorkspaceFiler interface {
	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.File, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.File, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.File) (*model.File, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.File) (*model.File, error)
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

type WorkspaceFileService struct {
	fileStores            map[string]minio.MinioFileStore
	fileStorePathManagers map[string]WorkspaceFileStorePathManager
}

func NewWorkspaceFileService(fileStores map[string]minio.MinioFileStore, fileStporePathManagers map[string]WorkspaceFileStorePathManager) (*WorkspaceFileService, error) {
	if len(fileStores) != len(fileStporePathManagers) {
		return nil, fmt.Errorf("file stores and path managers count mismatch")
	}

	fs := make(map[string]minio.MinioFileStore)
	fsMgr := make(map[string]WorkspaceFileStorePathManager)

	for name, store := range fileStores {
		storeMgr, ok := fileStporePathManagers[name]
		if !ok {
			return nil, fmt.Errorf("missing path manager for file store %s", name)
		}

		fs[name] = store
		fsMgr[name] = storeMgr
	}

	ws := &WorkspaceFileService{
		fileStores:            fs,
		fileStorePathManagers: fsMgr,
	}

	return ws, nil
}

func (s *WorkspaceFileService) findStore(filePath string) (string, error) {
	var selectedStoreName string
	maxPrefixLen := 0

	for storeName, storeMgr := range s.fileStorePathManagers {
		normalizedPath := storeMgr.NormalizePath(filePath)
		prefix := storeMgr.GetStorePrefix()

		if strings.HasPrefix(normalizedPath, prefix) && len(prefix) > maxPrefixLen {
			maxPrefixLen = len(prefix)
			selectedStoreName = storeName
		}
	}

	if selectedStoreName == "" {
		return "", fmt.Errorf("no suitable file store found for path %s", filePath)
	}

	return selectedStoreName, nil
}
func (s *WorkspaceFileService) findFileStore(filePath string) (minio.MinioFileStore, error) {
	storeName, err := s.findStore(filePath)
	if err != nil {
		return nil, err
	}

	return s.fileStores[storeName], nil
}

func (s *WorkspaceFileService) findFileStorePathManager(filePath string) (WorkspaceFileStorePathManager, error) {
	storeName, err := s.findStore(filePath)
	if err != nil {
		return nil, err
	}

	return s.fileStorePathManagers[storeName], nil
}

func (s *WorkspaceFileService) listStores() []*model.File {
	var stores []*model.File
	for _, storeMgr := range s.fileStorePathManagers {
		stores = append(stores, &model.File{
			Path:        storeMgr.GetStorePrefix(),
			Name:        storeMgr.GetStoreName(),
			IsDirectory: true,
		})
	}
	return stores
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.File, error) {
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

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.File, error) {
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

	var files []*model.File
	for _, f := range storeFiles {
		files = append(files, &model.File{
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

func (s *WorkspaceFileService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.File) (*model.File, error) {
	store, err := s.findFileStore(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePathManager, err := s.findFileStorePathManager(file.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to select file store: %w", err)
	}

	storePath := storePathManager.ToStorePath(workspaceID, file.Path)
	createdFile, err := store.CreateFile(ctx, &model.File{
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
		Path:        storePathManager.FromStorePath(workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.File) (*model.File, error) {
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
		createdFile, err := newStore.CreateFile(ctx, &model.File{
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

		return &model.File{
			Path:        newStorePathManager.FromStorePath(workspaceID, createdFile.Path),
			Name:        createdFile.Name,
			IsDirectory: createdFile.IsDirectory,
			Size:        createdFile.Size,
			MimeType:    createdFile.MimeType,
			UpdatedAt:   createdFile.UpdatedAt,
		}, nil
	}

	// Same store move
	updatedFile, err := oldStore.UpdateFile(ctx, oldStorePath, &model.File{
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
