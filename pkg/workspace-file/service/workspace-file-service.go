package service

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type WorkspaceFiler interface {
	ListWorkspaceFileStores(ctx context.Context, workspaceID uint64) ([]*model.WorkspaceFileStoreInfo, error)
	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error)
	GetWorkspaceFileWithContent(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*filestore.File, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *filestore.File) (*filestore.File, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *filestore.File, isCopy bool) (*filestore.File, error)
	DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error
	InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *filestore.File) (*filestore.FileUploadInfo, error)
	UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *filestore.FilePart) (*filestore.FilePart, error)
	CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*filestore.FilePart) (*filestore.File, error)
	AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error
}

type workspaceFileStore struct {
	workspacePrefix string
	description     string
	storeType       string
	enabled         bool
	order           uint
	store           filestore.FileStore
}

type WorkspaceFileService struct {
	stores map[string]workspaceFileStore
}

func NewWorkspaceFileService(cfg config.Config, fileStores map[string]filestore.FileStore) (*WorkspaceFileService, error) {
	storeConfigs := cfg.Services.WorkspaceFileService.Stores

	stores := make(map[string]workspaceFileStore, len(storeConfigs))
	for storeName, storeCfg := range storeConfigs {
		if !strings.Contains(storeCfg.WorkspacePrefix, "%s") {
			return nil, fmt.Errorf("workspace file store %q: workspace_prefix must contain %%s for workspace name substitution", storeName)
		}
		rawCfg, ok := cfg.Storage.FileStores[storeCfg.FileStoreName]
		if !ok {
			return nil, fmt.Errorf("workspace file store %q references unknown file store %q", storeName, storeCfg.FileStoreName)
		}
		if isFileStoreEnabled(rawCfg) && fileStores[storeCfg.FileStoreName] == nil {
			return nil, fmt.Errorf("workspace file store %q: file store %q is enabled but was not initialized", storeName, storeCfg.FileStoreName)
		}
		stores[storeName] = workspaceFileStore{
			workspacePrefix: storeCfg.WorkspacePrefix,
			description:     storeCfg.Description,
			storeType:       rawCfg.Type,
			enabled:         isFileStoreEnabled(rawCfg),
			order:           storeCfg.Order,
			store:           fileStores[storeCfg.FileStoreName],
		}
	}

	return &WorkspaceFileService{stores: stores}, nil
}

func isFileStoreEnabled(cfg config.FileStore) bool {
	switch cfg.Type {
	case "minio":
		return cfg.MinioConfig.Enabled
	case "disk":
		return cfg.DiskConfig.Enabled
	default:
		return false
	}
}

// toStorePath converts a user path (/{storeName}/relative/path) to the internal storage path.
func (s *WorkspaceFileService) toStorePath(storeName string, workspaceID uint64, filePath string) string {
	store := s.stores[storeName]
	normalizedPath := "/" + strings.TrimPrefix(filePath, "/")
	relPath := strings.TrimPrefix(normalizedPath, "/"+storeName)
	relPath = strings.TrimLeft(relPath, "/")
	workspaceDir := fmt.Sprintf(store.workspacePrefix, workspace_model.GetWorkspaceClusterName(workspaceID))
	return fmt.Sprintf("%s/%s", workspaceDir, relPath)
}

// fromStorePath converts an internal storage path back to a user path (/{storeName}/relative/path).
func (s *WorkspaceFileService) fromStorePath(storeName string, workspaceID uint64, storePath string) string {
	store := s.stores[storeName]
	workspaceDir := fmt.Sprintf(store.workspacePrefix, workspace_model.GetWorkspaceClusterName(workspaceID))
	relPath := strings.TrimPrefix(storePath, workspaceDir+"/")
	return "/" + storeName + "/" + strings.TrimPrefix(relPath, "/")
}

func (s *WorkspaceFileService) selectFileStore(filePath string) (string, error) {
	parts := strings.SplitN(strings.TrimPrefix(filePath, "/"), "/", 2)
	if parts[0] == "" {
		return "", cerr.ErrInvalidRequest.WithMessage("path must include a store name as the first segment")
	}
	storeName := parts[0]
	store, ok := s.stores[storeName]
	if !ok {
		return "", cerr.ErrInvalidRequest.WithMessage(fmt.Sprintf("Unknown file store: %s", storeName))
	}
	if !store.enabled {
		return "", cerr.ErrInvalidRequest.WithMessage(fmt.Sprintf("File store %s is disabled", storeName))
	}
	return storeName, nil
}

func (s *WorkspaceFileService) ListWorkspaceFileStores(ctx context.Context, workspaceID uint64) ([]*model.WorkspaceFileStoreInfo, error) {
	var storeInfos []*model.WorkspaceFileStoreInfo
	for storeName, store := range s.stores {
		// Determine store status
		var status model.WorkspaceFileStoreStatus
		switch {
		case !store.enabled:
			status = model.WorkspaceFileStoreStatusDisabled
		case store.store == nil:
			status = model.WorkspaceFileStoreStatusDisconnected
		default:
			if err := store.store.Ping(ctx); err != nil {
				logger.TechLog.Warn(ctx, fmt.Sprintf("file store %s is unreachable: %v", storeName, err))
				status = model.WorkspaceFileStoreStatusDisconnected
			} else {
				status = model.WorkspaceFileStoreStatusReady
			}
		}

		storeInfos = append(storeInfos, &model.WorkspaceFileStoreInfo{
			Name:        storeName,
			Type:        store.storeType,
			Description: store.description,
			Status:      status,
		})
	}

	// Sort stores by order, then by name
	slices.SortFunc(storeInfos, func(a, b *model.WorkspaceFileStoreInfo) int {
		orderA, orderB := s.stores[a.Name].order, s.stores[b.Name].order
		if orderA != orderB {
			return int(orderA) - int(orderB)
		}
		return strings.Compare(a.Name, b.Name)
	})

	return storeInfos, nil
}

func (s *WorkspaceFileService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	storePath := s.toStorePath(storeName, workspaceID, filePath)

	// Returns only file metadata without content
	file, err := s.stores[storeName].store.StatFile(ctx, storePath)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workspace file at path %s", filePath))
	}

	return file, nil
}

func (s *WorkspaceFileService) GetWorkspaceFileWithContent(ctx context.Context, workspaceID uint64, filePath string) (*filestore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	storePath := s.toStorePath(storeName, workspaceID, filePath)

	file, err := s.stores[storeName].store.GetFile(ctx, storePath)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workspace file with content at path %s", filePath))
	}

	return file, nil
}

func (s *WorkspaceFileService) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*filestore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	storePath := s.toStorePath(storeName, workspaceID, filePath)
	storeFiles, err := s.stores[storeName].store.ListFiles(ctx, storePath)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to list workspace files at path %s", filePath))
	}

	var files []*filestore.File
	for _, f := range storeFiles {
		files = append(files, &filestore.File{
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

func (s *WorkspaceFileService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *filestore.File) (*filestore.File, error) {
	storeName, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, err
	}

	storePath := s.toStorePath(storeName, workspaceID, file.Path)
	store := s.stores[storeName].store

	var createdFile *filestore.File
	if file.IsDirectory {
		createdFile, err = store.CreateDirectory(ctx, &filestore.File{
			Path:        storePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
		})
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create workspace directory at path %s", file.Path))
		}
	} else {
		createdFile, err = store.CreateFile(ctx, &filestore.File{
			Path:        storePath,
			Name:        file.Name,
			IsDirectory: file.IsDirectory,
			MimeType:    file.MimeType,
			Content:     file.Content,
		})
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create workspace file at path %s", file.Path))
		}
	}

	return &filestore.File{
		Path:        s.fromStorePath(storeName, workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, sourcePath string, file *filestore.File, isCopy bool) (*filestore.File, error) {
	sourceStoreName, err := s.selectFileStore(sourcePath)
	if err != nil {
		return nil, err
	}

	destinationStoreName, err := s.selectFileStore(file.Path)
	if err != nil {
		return nil, err
	}

	sourceStorePath := s.toStorePath(sourceStoreName, workspaceID, sourcePath)
	destinationStorePath := s.toStorePath(destinationStoreName, workspaceID, file.Path)

	sourceStore := s.stores[sourceStoreName].store
	destinationStore := s.stores[destinationStoreName].store

	if !isCopy && sourceStoreName == destinationStoreName {
		// Same-store move
		_, err = sourceStore.StatFile(ctx, sourceStorePath)
		if err != nil {
			return nil, cerr.ErrNotFound.Wrap(err, fmt.Sprintf("Workspace file at path %s does not exist", sourcePath))
		}

		if _, statErr := sourceStore.StatFile(ctx, destinationStorePath); statErr == nil {
			return nil, cerr.ErrAlreadyExists.WithMessage(fmt.Sprintf("Workspace file already exists at path %s", file.Path))
		}

		updatedFile, err := sourceStore.MoveFile(ctx, sourceStorePath, destinationStorePath)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to move workspace file from path %s", sourcePath))
		}

		return &filestore.File{
			Path:        s.fromStorePath(destinationStoreName, workspaceID, updatedFile.Path),
			Name:        updatedFile.Name,
			IsDirectory: updatedFile.IsDirectory,
			Size:        updatedFile.Size,
			MimeType:    updatedFile.MimeType,
			UpdatedAt:   updatedFile.UpdatedAt,
		}, nil
	}

	// Copy (same-store or cross-store) and cross-store move
	sourceFile, err := sourceStore.StatFile(ctx, sourceStorePath)
	if err != nil {
		return nil, cerr.ErrNotFound.Wrap(err, fmt.Sprintf("Workspace file at path %s does not exist", sourcePath))
	}

	// Check if destination file already exists
	if _, statErr := destinationStore.StatFile(ctx, destinationStorePath); statErr == nil {
		return nil, cerr.ErrAlreadyExists.WithMessage(fmt.Sprintf("Workspace file already exists at path %s", file.Path))
	}

	uploadInfo, err := destinationStore.InitiateMultipartUpload(ctx, &filestore.File{
		Path:        destinationStorePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		MimeType:    file.MimeType,
		Size:        sourceFile.Size,
	})
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to initiate multipart upload for workspace file at path %s", file.Path))
	}

	// Ensure multipart upload is aborted if anything goes wrong during the copy process
	var completed bool
	defer func() {
		if !completed {
			if abortErr := destinationStore.AbortMultipartUpload(ctx, destinationStorePath, uploadInfo.UploadID); abortErr != nil {
				logger.TechLog.Warn(ctx, fmt.Sprintf("Failed to abort multipart upload %s at path %s: %v", uploadInfo.UploadID, destinationStorePath, abortErr))
			}
		}
	}()

	// Stream file from source to destination in parts
	reader, _, err := sourceStore.GetFileStream(ctx, sourceStorePath)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to open stream for workspace file at path %s", sourceStorePath))
	}
	defer reader.Close()

	parts, err := uploadFromReader(ctx, destinationStore, destinationStorePath, uploadInfo.UploadID, uploadInfo.PartSize, reader)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to upload workspace file at path %s", file.Path))
	}

	// Complete the multipart upload to finalize the copy
	createdFile, err := destinationStore.CompleteMultipartUpload(ctx, destinationStorePath, uploadInfo.UploadID, parts)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to complete multipart upload for workspace file at path %s", file.Path))
	}
	completed = true

	// Delete source file if this is a move operation (not copy)
	if !isCopy {
		if err := sourceStore.DeleteFile(ctx, sourceStorePath); err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to delete source workspace file at path %s after move (file was copied to destination but source was not removed)", sourceStorePath))
		}
	}

	return &filestore.File{
		Path:        s.fromStorePath(destinationStoreName, workspaceID, createdFile.Path),
		Name:        createdFile.Name,
		IsDirectory: createdFile.IsDirectory,
		Size:        createdFile.Size,
		MimeType:    createdFile.MimeType,
		UpdatedAt:   createdFile.UpdatedAt,
	}, nil
}

func (s *WorkspaceFileService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return err
	}

	store := s.stores[storeName].store
	storePath := s.toStorePath(storeName, workspaceID, filePath)
	if strings.HasSuffix(storePath, "/") {
		err = store.DeleteDirectory(ctx, storePath)
		if err != nil {
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to delete workspace directory at path %s", filePath))
		}
	} else {
		err = store.DeleteFile(ctx, storePath)
		if err != nil {
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to delete workspace file at path %s", filePath))
		}
	}

	return nil
}

func (s *WorkspaceFileService) InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *filestore.File) (*filestore.FileUploadInfo, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	store := s.stores[storeName].store
	storePath := s.toStorePath(storeName, workspaceID, file.Path)

	uploadInfo, err := store.InitiateMultipartUpload(ctx, &filestore.File{
		Path:        storePath,
		Name:        file.Name,
		IsDirectory: file.IsDirectory,
		MimeType:    file.MimeType,
		Size:        file.Size,
	})
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to initiate multipart upload for file at path %s", file.Path))
	}

	return uploadInfo, nil
}

func (s *WorkspaceFileService) UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *filestore.FilePart) (*filestore.FilePart, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	store := s.stores[storeName].store
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	uploadedPart, err := store.UploadPart(ctx, storePath, uploadID, part)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to upload part number %d for upload ID %s at path %s", part.PartNumber, uploadID, filePath))
	}

	return uploadedPart, nil
}

func (s *WorkspaceFileService) CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*filestore.FilePart) (*filestore.File, error) {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return nil, err
	}

	store := s.stores[storeName].store
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	completedFile, err := store.CompleteMultipartUpload(ctx, storePath, uploadID, parts)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to complete multipart upload for upload ID %s at path %s", uploadID, filePath))
	}

	return completedFile, nil
}

func (s *WorkspaceFileService) AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error {
	storeName, err := s.selectFileStore(filePath)
	if err != nil {
		return err
	}

	store := s.stores[storeName].store
	storePath := s.toStorePath(storeName, workspaceID, filePath)

	err = store.AbortMultipartUpload(ctx, storePath, uploadID)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to abort multipart upload for upload ID %s at path %s", uploadID, filePath))
	}

	return nil
}

func uploadFromReader(ctx context.Context, store filestore.FileStore, destPath, uploadID string, partSize uint64, reader io.Reader) ([]*filestore.FilePart, error) {
	var parts []*filestore.FilePart
	buf := make([]byte, partSize)
	var partNumber uint64 = 1
	for {
		n, readErr := io.ReadFull(reader, buf)
		if n > 0 {
			part, err := store.UploadPart(ctx, destPath, uploadID, &filestore.FilePart{
				PartNumber: partNumber,
				Data:       buf[:n],
			})
			if err != nil {
				return nil, fmt.Errorf("part %d: %w", partNumber, err)
			}
			parts = append(parts, part)
			partNumber++
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	return parts, nil
}
