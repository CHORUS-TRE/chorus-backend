package diskfilestore

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ filestore.FileStore = &diskFileStorage{}

const (
	multipartDir  = ".multipart"
	minPartSize   = 5 * 1024 * 1024        // 5 MB
	maxPartSize   = 5 * 1024 * 1024 * 1024 // 5 GB
	maxTotalParts = 10000
	dirMarkerFile = ".directory"
)

type multipartUpload struct {
	FilePath   string
	FileSize   uint64
	Parts      map[uint64]*filestore.FilePart
	PartSize   uint64
	TotalParts uint64
	CreatedAt  time.Time
}

type diskFileStorage struct {
	basePath         string
	multipartUploads map[string]*multipartUpload
	mu               sync.RWMutex
}

// NewDiskFileStorage creates a new disk-based file storage at the given path.
// It recursively creates all needed folders if they don't exist.
func NewDiskFileStorage(basePath string) (filestore.FileStore, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create base directory %s: %w", basePath, err)
	}

	// Create multipart directory
	multipartPath := filepath.Join(basePath, multipartDir)
	if err := os.MkdirAll(multipartPath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create multipart directory %s: %w", multipartPath, err)
	}

	return &diskFileStorage{
		basePath:         basePath,
		multipartUploads: make(map[string]*multipartUpload),
	}, nil
}

func (s *diskFileStorage) resolvePath(path string) string {
	cleanPath := filepath.Clean(strings.TrimPrefix(path, "/"))
	return filepath.Join(s.basePath, cleanPath)
}

func (s *diskFileStorage) computeFilePartSize(fileSize uint64) (uint64, uint64, error) {
	if fileSize == 0 {
		return 0, 0, fmt.Errorf("unable to compute part size for empty files")
	}

	if fileSize > maxPartSize*maxTotalParts {
		return 0, 0, fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", fileSize, maxPartSize*maxTotalParts)
	}

	// Single part upload for small files
	if fileSize <= minPartSize {
		return fileSize, 1, nil
	}

	partSize := uint64(minPartSize)
	// Ensure we do not exceed MaxTotalParts (use ceiling division)
	if (fileSize+partSize-1)/partSize > maxTotalParts {
		partSize = (fileSize + maxTotalParts - 1) / maxTotalParts

		// Ensure partSize respects bounds
		if partSize < minPartSize {
			partSize = minPartSize
		}
		if partSize > maxPartSize {
			return 0, 0, fmt.Errorf("file size %d exceeds maximum uploadable size", fileSize)
		}
	}

	totalParts := (fileSize + partSize - 1) / partSize

	return partSize, totalParts, nil
}

func (s *diskFileStorage) getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func (s *diskFileStorage) fileInfoToFile(path string, info os.FileInfo) *filestore.File {
	isDir := info.IsDir()
	var size uint64
	if !isDir {
		size = uint64(info.Size())
	}

	// Normalize path to use forward slashes
	normalizedPath := filepath.ToSlash(strings.TrimPrefix(path, s.basePath))
	normalizedPath = strings.TrimPrefix(normalizedPath, "/")

	if isDir && !strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "/"
	}

	return &filestore.File{
		Path:        normalizedPath,
		Name:        info.Name(),
		IsDirectory: isDir,
		Size:        size,
		MimeType:    s.getMimeType(path),
		UpdatedAt:   info.ModTime(),
	}
}

func (s *diskFileStorage) StatFile(ctx context.Context, path string) (*filestore.File, error) {
	fullPath := s.resolvePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found at %s", path)
		}
		return nil, fmt.Errorf("unable to stat file at %s: %w", path, err)
	}

	file := s.fileInfoToFile(fullPath, info)

	logger.TechLog.Info(ctx, fmt.Sprintf("Fetched metadata for %s", path))
	return file, nil
}

func (s *diskFileStorage) GetFile(ctx context.Context, path string) (*filestore.File, error) {
	fullPath := s.resolvePath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found at %s", path)
		}
		return nil, fmt.Errorf("unable to stat file at %s: %w", path, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path %s is a directory, use StatFile for directories", path)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file at %s: %w", path, err)
	}

	file := s.fileInfoToFile(fullPath, info)
	file.Content = content

	logger.TechLog.Info(ctx, fmt.Sprintf("Downloaded %s", path))
	return file, nil
}

func (s *diskFileStorage) ListFiles(ctx context.Context, path string) ([]*filestore.File, error) {
	fullPath := s.resolvePath(path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found at %s", path)
		}
		return nil, fmt.Errorf("unable to list files at path %s: %w", path, err)
	}

	var files []*filestore.File
	for _, entry := range entries {
		// Skip hidden multipart directory and directory marker files
		if entry.Name() == multipartDir || entry.Name() == dirMarkerFile {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		entryPath := filepath.Join(fullPath, entry.Name())
		file := s.fileInfoToFile(entryPath, info)
		files = append(files, file)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Listed %d files from path %s", len(files), path))
	return files, nil
}

func (s *diskFileStorage) CreateFile(ctx context.Context, file *filestore.File) (*filestore.File, error) {
	if file.IsDirectory {
		return nil, fmt.Errorf("use CreateDirectory to create directories")
	}

	path := strings.TrimSuffix(file.Path, "/")
	fullPath := s.resolvePath(path)

	// Check for existing file
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("a file already exists at %s", path)
	}

	// Check for conflicting directory
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("a directory with conflicting name exists at %s", path)
	}

	// Create parent directories if needed
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create parent directory for %s: %w", path, err)
	}

	// Write file content
	if err := os.WriteFile(fullPath, file.Content, 0644); err != nil {
		return nil, fmt.Errorf("unable to write file at %s: %w", path, err)
	}

	// Get file info to return
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created file at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created %s", path))

	createdFile := s.fileInfoToFile(fullPath, info)
	return createdFile, nil
}

func (s *diskFileStorage) CreateDirectory(ctx context.Context, file *filestore.File) (*filestore.File, error) {
	if !file.IsDirectory {
		return nil, fmt.Errorf("use CreateFile to create files")
	}

	dirPath := strings.TrimSuffix(file.Path, "/") + "/"
	fullPath := s.resolvePath(dirPath)

	// Check for existing directory
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return nil, fmt.Errorf("a directory already exists at %s", dirPath)
	}

	// Check for conflicting file
	filePath := strings.TrimSuffix(fullPath, "/")
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		return nil, fmt.Errorf("a file with conflicting name exists at %s", filePath)
	}

	// Create directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory at %s: %w", dirPath, err)
	}

	// Create a marker file to ensure directory persistence (similar to S3)
	markerPath := filepath.Join(fullPath, dirMarkerFile)
	if err := os.WriteFile(markerPath, []byte{}, 0644); err != nil {
		return nil, fmt.Errorf("unable to create directory marker at %s: %w", dirPath, err)
	}

	// Get directory info to return
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created directory at %s: %w", dirPath, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created directory %s", dirPath))

	createdDir := s.fileInfoToFile(fullPath, info)
	return createdDir, nil
}

func (s *diskFileStorage) MoveFile(ctx context.Context, oldPath string, newPath string) (*filestore.File, error) {
	fullOldPath := s.resolvePath(oldPath)
	fullNewPath := s.resolvePath(newPath)

	// Check if source exists
	if _, err := os.Stat(fullOldPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file not found at %s", oldPath)
	}

	// Create parent directory for destination if needed
	parentDir := filepath.Dir(fullNewPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create parent directory for %s: %w", newPath, err)
	}

	// Move/rename file
	if err := os.Rename(fullOldPath, fullNewPath); err != nil {
		return nil, fmt.Errorf("unable to move file from %s to %s: %w", oldPath, newPath, err)
	}

	// Get moved file info
	info, err := os.Stat(fullNewPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify moved file at %s: %w", newPath, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Moved file from %s to %s", oldPath, newPath))

	movedFile := s.fileInfoToFile(fullNewPath, info)
	return movedFile, nil
}

func (s *diskFileStorage) DeleteFile(ctx context.Context, path string) error {
	if strings.HasSuffix(path, "/") {
		return fmt.Errorf("use DeleteDirectory to delete directories")
	}

	fullPath := s.resolvePath(path)

	// Check if it's actually a file
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found at %s", path)
		}
		return fmt.Errorf("unable to stat file at %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("use DeleteDirectory to delete directories")
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("unable to delete file at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted file at %s", path))
	return nil
}

func (s *diskFileStorage) DeleteDirectory(ctx context.Context, path string) error {
	if !strings.HasSuffix(path, "/") {
		return fmt.Errorf("use DeleteFile to delete files")
	}

	fullPath := s.resolvePath(path)

	// Check if directory exists and is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory not found at %s", path)
		}
		return fmt.Errorf("unable to stat directory at %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("use DeleteFile to delete files")
	}

	// Remove directory and all contents recursively
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("unable to delete directory at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted directory %s", path))
	return nil
}

func (s *diskFileStorage) InitiateMultipartUpload(ctx context.Context, file *filestore.File) (*filestore.FileUploadInfo, error) {
	if file.IsDirectory {
		return nil, fmt.Errorf("use CreateDirectory to create directories")
	}

	path := strings.TrimSuffix(file.Path, "/")
	fullPath := s.resolvePath(path)

	// Check for existing file
	if _, err := os.Stat(fullPath); err == nil {
		return nil, fmt.Errorf("a file already exists at %s", path)
	}

	// Check for conflicting directory
	dirPath := fullPath + "/"
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		return nil, fmt.Errorf("a directory with conflicting name exists at %s", dirPath)
	}

	partSize, totalParts, err := s.computeFilePartSize(file.Size)
	if err != nil {
		return nil, fmt.Errorf("unable to compute part size for file at %s: %w", path, err)
	}

	// Generate upload ID
	uploadID := s.generateUploadID(path)

	// Create upload directory
	uploadDir := filepath.Join(s.basePath, multipartDir, uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create upload directory: %w", err)
	}

	// Store upload metadata
	s.mu.Lock()
	s.multipartUploads[uploadID] = &multipartUpload{
		FilePath:   path,
		FileSize:   file.Size,
		Parts:      make(map[uint64]*filestore.FilePart),
		PartSize:   partSize,
		TotalParts: totalParts,
		CreatedAt:  time.Now(),
	}
	s.mu.Unlock()

	logger.TechLog.Info(ctx, fmt.Sprintf("Initiated multipart upload for %s with upload ID %s (%d parts of size %d)", path, uploadID, totalParts, partSize))
	return &filestore.FileUploadInfo{
		UploadID:   uploadID,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

func (s *diskFileStorage) generateUploadID(path string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s-%d", path, timestamp)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *diskFileStorage) UploadPart(ctx context.Context, filePath string, uploadId string, part *filestore.FilePart) (*filestore.FilePart, error) {
	s.mu.RLock()
	upload, exists := s.multipartUploads[uploadId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("upload ID %s not found", uploadId)
	}

	// Write part to disk
	partPath := filepath.Join(s.basePath, multipartDir, uploadId, fmt.Sprintf("part-%d", part.PartNumber))
	if err := os.WriteFile(partPath, part.Data, 0644); err != nil {
		return nil, fmt.Errorf("unable to write part %d: %w", part.PartNumber, err)
	}

	// Calculate ETag (MD5 hash)
	hash := md5.Sum(part.Data)
	etag := hex.EncodeToString(hash[:])

	// Store part metadata
	s.mu.Lock()
	upload.Parts[part.PartNumber] = &filestore.FilePart{
		PartNumber: part.PartNumber,
		ETag:       etag,
	}
	s.mu.Unlock()

	logger.TechLog.Info(ctx, fmt.Sprintf("Uploaded part %d for upload ID %s", part.PartNumber, uploadId))
	return &filestore.FilePart{
		PartNumber: part.PartNumber,
		ETag:       etag,
	}, nil
}

func (s *diskFileStorage) CompleteMultipartUpload(ctx context.Context, filePath string, uploadId string, parts []*filestore.FilePart) (*filestore.File, error) {
	s.mu.RLock()
	upload, exists := s.multipartUploads[uploadId]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("upload ID %s not found", uploadId)
	}

	// Validate all parts are present
	if uint64(len(parts)) != upload.TotalParts {
		return nil, fmt.Errorf("expected %d parts, got %d", upload.TotalParts, len(parts))
	}

	fullPath := s.resolvePath(upload.FilePath)

	// Create parent directory if needed
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create parent directory: %w", err)
	}

	// Create final file
	finalFile, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create final file: %w", err)
	}
	defer finalFile.Close()

	// Concatenate all parts in order
	uploadDir := filepath.Join(s.basePath, multipartDir, uploadId)
	for _, part := range parts {
		partPath := filepath.Join(uploadDir, fmt.Sprintf("part-%d", part.PartNumber))
		partData, err := os.ReadFile(partPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read part %d: %w", part.PartNumber, err)
		}

		if _, err := finalFile.Write(partData); err != nil {
			return nil, fmt.Errorf("unable to write part %d to final file: %w", part.PartNumber, err)
		}
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify completed file: %w", err)
	}

	// Clean up upload directory
	if err := os.RemoveAll(uploadDir); err != nil {
		logger.TechLog.Warn(ctx, fmt.Sprintf("Unable to clean up upload directory %s: %v", uploadDir, err))
	}

	// Remove upload from memory
	s.mu.Lock()
	delete(s.multipartUploads, uploadId)
	s.mu.Unlock()

	logger.TechLog.Info(ctx, fmt.Sprintf("Completed multipart upload %s", uploadId))

	createdFile := s.fileInfoToFile(fullPath, info)
	return createdFile, nil
}

func (s *diskFileStorage) AbortMultipartUpload(ctx context.Context, filePath string, uploadId string) error {
	s.mu.RLock()
	_, exists := s.multipartUploads[uploadId]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("upload ID %s not found", uploadId)
	}

	// Remove upload directory
	uploadDir := filepath.Join(s.basePath, multipartDir, uploadId)
	if err := os.RemoveAll(uploadDir); err != nil {
		return fmt.Errorf("unable to remove upload directory: %w", err)
	}

	// Remove upload from memory
	s.mu.Lock()
	delete(s.multipartUploads, uploadId)
	s.mu.Unlock()

	logger.TechLog.Info(ctx, fmt.Sprintf("Aborted multipart upload %s", uploadId))
	return nil
}
