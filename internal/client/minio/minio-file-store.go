package minio

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

// The MinioFileStore interface abstracts UNIX-like file operations which can be performed on a MinIO object storage.
// The directories are represented as objects with keys ending in a '/' character.
type MinioFileStore interface {
	// Get file metadata at the specified path without downloading the content.
	StatFile(ctx context.Context, path string) (*model.File, error)

	// Get the file at the specified path, including its content.
	GetFile(ctx context.Context, path string) (*model.File, error)

	// List files and directories at the specified path.
	ListFiles(ctx context.Context, path string) ([]*model.File, error)

	// Create a new file at the specified path.
	CreateFile(ctx context.Context, file *model.File) (*model.File, error)

	// Create a new directory at the specified path.
	CreateDirectory(ctx context.Context, file *model.File) (*model.File, error)

	// Move a file from oldPath to newPath.
	MoveFile(ctx context.Context, oldPath string, newPath string) (*model.File, error)

	// Delete a file at the specified path.
	DeleteFile(ctx context.Context, path string) error

	// Delete a directory and all its contents recursively.
	DeleteDirectory(ctx context.Context, path string) error

	// Initiate a new multipart upload for a file.
	InitiateMultipartUpload(ctx context.Context, file *model.File) (*model.FileUploadInfo, error)

	// Upload a single part of a multipart upload.
	UploadPart(ctx context.Context, uploadId string, part *model.FilePart) (*model.FilePart, error)

	// Complete a multipart upload after all parts of a file have been uploaded.
	CompleteMultipartUpload(ctx context.Context, file *model.File, uploadId string, parts []*model.FilePart) (*model.File, error)

	// Abort a multipart upload, discarding all uploaded parts.
	AbortMultipartUpload(ctx context.Context, uploadId string) error
}

var _ MinioFileStore = &MinioFileStorage{}

type MinioFileStorage struct {
	minioClient miniorawclient.MinioClienter
}

func NewMinioFileStorage(client miniorawclient.MinioClienter) (*MinioFileStorage, error) {
	return &MinioFileStorage{
		minioClient: client,
	}, nil
}

func (s *MinioFileStorage) computeFilePartSize(fileSize uint64) (uint64, uint64) {
	const (
		MinPartSize    = 5 * 1024 * 1024 // 5 MB
		MaxTotalParts  = 10000           // Max parts allowed by MinIO/S3
		PreferredParts = 1000            // Preferred number of parts
	)

	if fileSize == 0 {
		return MinPartSize, 1
	}

	if fileSize <= MinPartSize {
		return fileSize, 1
	}

	partSize := fileSize / PreferredParts
	if partSize < MinPartSize {
		partSize = MinPartSize
	}

	// Ensure we do not exceed MaxTotalParts
	if (fileSize+partSize-1)/partSize > MaxTotalParts {
		partSize = (fileSize + MaxTotalParts - 1) / MaxTotalParts
	}

	totalParts := (fileSize + partSize - 1) / partSize

	return partSize, totalParts
}

func (s *MinioFileStorage) StatFile(ctx context.Context, path string) (*model.File, error) {
	objectInfo, err := s.minioClient.StatObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to stat file at %s: %w", path, err)
	}

	file := model.MinioObjectInfoToFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("Fetched metadata for %s", path))
	return file, nil
}

func (s *MinioFileStorage) GetFile(ctx context.Context, path string) (*model.File, error) {
	object, err := s.minioClient.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get file at %s: %w", path, err)
	}

	file := model.MinioObjectToFile(object)

	logger.TechLog.Info(ctx, fmt.Sprintf("Downloaded %s", path))
	return file, nil
}

func (s *MinioFileStorage) ListFiles(ctx context.Context, path string) ([]*model.File, error) {
	objects, err := s.minioClient.ListObjects(path, false)
	if err != nil {
		return nil, fmt.Errorf("unable to list files at path %s: %w", path, err)
	}

	var files []*model.File
	for _, objectInfo := range objects {
		file := model.MinioObjectInfoToFile(objectInfo)
		files = append(files, file)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Listed %d files from path %s", len(files), path))
	return files, nil
}

func (s *MinioFileStorage) CreateFile(ctx context.Context, file *model.File) (*model.File, error) {
	if file.IsDirectory {
		return nil, fmt.Errorf("use CreateDirectory to create directories")
	}

	path := strings.TrimSuffix(file.Path, "/")

	// Check for existing file
	existingObject, err := s.minioClient.StatObject(path)
	if existingObject != nil {
		return nil, fmt.Errorf("a file already exists at %s", path)
	}

	// Check for conflicting directory
	dirPath := path + "/"
	object, err := s.minioClient.StatObject(dirPath)
	if object != nil {
		return nil, fmt.Errorf("a directory with conflicting name exists at %s", dirPath)
	}

	// Upload file
	_, err = s.minioClient.PutObject(path, model.FileToMinioObject(file))
	if err != nil {
		return nil, fmt.Errorf("unable to upload file at %s: %w", path, err)
	}

	// Verify file upload
	objectInfo, err := s.minioClient.StatObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created file at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created %s", path))

	createdFile := model.MinioObjectInfoToFile(objectInfo)

	return createdFile, nil
}

func (s *MinioFileStorage) CreateDirectory(ctx context.Context, file *model.File) (*model.File, error) {
	if file.IsDirectory == false {
		return nil, fmt.Errorf("use CreateFile to create files")
	}

	dirPath := strings.TrimSuffix(file.Path, "/") + "/"

	// Check for existing directory
	existingObject, err := s.minioClient.StatObject(dirPath)
	if existingObject != nil {
		return nil, fmt.Errorf("a directory already exists at %s", dirPath)
	}

	// Check for conflicting file
	object, err := s.minioClient.StatObject(strings.TrimSuffix(dirPath, "/"))
	if object != nil {
		return nil, fmt.Errorf("a file with conflicting name exists at %s", strings.TrimSuffix(dirPath, "/"))
	}

	// Create directory as an empty object with trailing '/'
	_, err = s.minioClient.PutObject(dirPath, &miniorawclient.MinioObject{
		MinioObjectInfo: miniorawclient.MinioObjectInfo{
			Key: dirPath,
		},
		Content: []byte{},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create directory at %s: %w", dirPath, err)
	}

	// Verify directory creation
	objectInfo, err := s.minioClient.StatObject(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify created directory at %s: %w", dirPath, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Created directory %s", dirPath))

	createdDir := model.MinioObjectInfoToFile(objectInfo)

	return createdDir, nil
}

func (s *MinioFileStorage) MoveFile(ctx context.Context, oldPath string, newPath string) (*model.File, error) {
	err := s.minioClient.MoveObject(oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("unable to move file from %s to %s: %w", oldPath, newPath, err)
	}

	// Verify moved file
	objectInfo, err := s.minioClient.StatObject(newPath)
	if err != nil {
		return nil, fmt.Errorf("unable to verify moved file at %s: %w", newPath, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Moved file from %s to %s", oldPath, newPath))

	movedFile := model.MinioObjectInfoToFile(objectInfo)

	return movedFile, nil
}

func (s *MinioFileStorage) DeleteFile(ctx context.Context, path string) error {
	if strings.HasSuffix(path, "/") {
		return fmt.Errorf("use DeleteDirectory to delete directories")
	}

	err := s.minioClient.DeleteObject(path)
	if err != nil {
		return fmt.Errorf("unable to delete file at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Deleted file at %s", path))
	return nil
}

func (s *MinioFileStorage) DeleteDirectory(ctx context.Context, path string) error {
	if !strings.HasSuffix(path, "/") {
		return fmt.Errorf("use DeleteFile to delete files")
	}

	// List all objects with the given prefix, recursively
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

func (s *MinioFileStorage) InitiateMultipartUpload(ctx context.Context, file *model.File) (*model.FileUploadInfo, error) {
	if file.IsDirectory {
		return nil, fmt.Errorf("use CreateDirectory to create directories")
	}

	path := strings.TrimSuffix(file.Path, "/")

	// Check for existing file
	existingObject, err := s.minioClient.StatObject(path)
	if existingObject != nil {
		return nil, fmt.Errorf("a file already exists at %s", path)
	}

	// Check for conflicting directory
	dirPath := path + "/"
	object, err := s.minioClient.StatObject(dirPath)
	if object != nil {
		return nil, fmt.Errorf("a directory with conflicting name exists at %s", dirPath)
	}

	uploadID, err := s.minioClient.NewMultipartUpload(path, file.Size)
	if err != nil {
		return nil, fmt.Errorf("unable to initiate multipart upload for file at %s: %w", path, err)
	}

	partSize, totalParts := s.computeFilePartSize(file.Size)

	logger.TechLog.Info(ctx, fmt.Sprintf("Initiated multipart upload for %s with upload ID %s (%d parts of size %d)", path, uploadID, totalParts, partSize))
	return &model.FileUploadInfo{
		UploadID:   uploadID,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

func (s *MinioFileStorage) UploadPart(ctx context.Context, uploadId string, part *model.FilePart) (*model.FilePart, error) {
	minioPart, err := s.minioClient.PutObjectPart(uploadId, int(part.PartNumber), part.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to upload part %d for upload ID %s: %w", part.PartNumber, uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Uploaded part %d for upload ID %s", part.PartNumber, uploadId))
	return &model.FilePart{
		PartNumber: uint64(minioPart.PartNumber),
		ETag:       minioPart.ETag,
	}, nil
}

func (s *MinioFileStorage) CompleteMultipartUpload(ctx context.Context, file *model.File, uploadId string, parts []*model.FilePart) (*model.File, error) {
	var completeParts []*miniorawclient.MinioObjectPartInfo
	for _, part := range parts {
		completeParts = append(completeParts, model.FilePartToMinioObjectPartInfo(part))
	}

	uploadInfo, err := s.minioClient.CompleteMultipartUpload(file.Path, uploadId, completeParts)
	if err != nil {
		return nil, fmt.Errorf("unable to complete multipart upload %s: %w", uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Completed multipart upload %s", uploadId))

	createdFile := model.MinioObjectInfoToFile(&uploadInfo.MinioObjectInfo)

	return createdFile, nil
}

func (s *MinioFileStorage) AbortMultipartUpload(ctx context.Context, uploadId string) error {
	err := s.minioClient.AbortMultipartUpload(uploadId)
	if err != nil {
		return fmt.Errorf("unable to abort multipart upload %s: %w", uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Aborted multipart upload %s", uploadId))
	return nil
}
