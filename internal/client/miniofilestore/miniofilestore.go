package miniofilestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/miniofilestore/model"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/miniofilestore/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ filestore.FileStore = &minioFileStorage{}

type minioFileStorage struct {
	minioClient miniorawclient.MinioClienter
}

func NewMinioFileStorage(client miniorawclient.MinioClienter) (filestore.FileStore, error) {
	return &minioFileStorage{
		minioClient: client,
	}, nil
}

func (s *minioFileStorage) computeFilePartSize(fileSize uint64) (uint64, uint64, error) {
	cfg := s.minioClient.GetClientConfig()
	minPartSize := cfg.MultipartMinPartSize
	maxPartSize := cfg.MultipartMaxPartSize
	maxTotalParts := cfg.MultipartMaxTotalParts

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

	partSize := minPartSize
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

func (s *minioFileStorage) StatFile(ctx context.Context, path string) (*filestore.File, error) {
	objectInfo, err := s.minioClient.StatObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to stat file at %s: %w", path, err)
	}

	file := model.MinioObjectInfoToFile(objectInfo)

	logger.TechLog.Info(ctx, fmt.Sprintf("Fetched metadata for %s", path))
	return file, nil
}

func (s *minioFileStorage) GetFile(ctx context.Context, path string) (*filestore.File, error) {
	object, err := s.minioClient.GetObject(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get file at %s: %w", path, err)
	}

	file := model.MinioObjectToFile(object)

	logger.TechLog.Info(ctx, fmt.Sprintf("Downloaded %s", path))
	return file, nil
}

func (s *minioFileStorage) ListFiles(ctx context.Context, path string) ([]*filestore.File, error) {
	objects, err := s.minioClient.ListObjects(path, false)
	if err != nil {
		return nil, fmt.Errorf("unable to list files at path %s: %w", path, err)
	}

	var files []*filestore.File
	for _, objectInfo := range objects {
		file := model.MinioObjectInfoToFile(objectInfo)
		files = append(files, file)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Listed %d files from path %s", len(files), path))
	return files, nil
}

func (s *minioFileStorage) CreateFile(ctx context.Context, file *filestore.File) (*filestore.File, error) {
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

func (s *minioFileStorage) CreateDirectory(ctx context.Context, file *filestore.File) (*filestore.File, error) {
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

func (s *minioFileStorage) MoveFile(ctx context.Context, oldPath string, newPath string) (*filestore.File, error) {
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

func (s *minioFileStorage) DeleteFile(ctx context.Context, path string) error {
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

func (s *minioFileStorage) DeleteDirectory(ctx context.Context, path string) error {
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

func (s *minioFileStorage) InitiateMultipartUpload(ctx context.Context, file *filestore.File) (*filestore.FileUploadInfo, error) {
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

	partSize, totalParts, err := s.computeFilePartSize(file.Size)
	if err != nil {
		return nil, fmt.Errorf("unable to compute part size for file at %s: %w", path, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Initiated multipart upload for %s with upload ID %s (%d parts of size %d)", path, uploadID, totalParts, partSize))
	return &filestore.FileUploadInfo{
		UploadID:   uploadID,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

func (s *minioFileStorage) UploadPart(ctx context.Context, filePath string, uploadId string, part *filestore.FilePart) (*filestore.FilePart, error) {
	minioPart, err := s.minioClient.PutObjectPart(filePath, uploadId, int(part.PartNumber), part.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to upload part %d for upload ID %s: %w", part.PartNumber, uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Uploaded part %d for upload ID %s", part.PartNumber, uploadId))
	return &filestore.FilePart{
		PartNumber: uint64(minioPart.PartNumber),
		ETag:       minioPart.ETag,
	}, nil
}

func (s *minioFileStorage) CompleteMultipartUpload(ctx context.Context, filePath string, uploadId string, parts []*filestore.FilePart) (*filestore.File, error) {
	var completeParts []*miniorawclient.MinioObjectPartInfo
	for _, part := range parts {
		completeParts = append(completeParts, model.FilePartToMinioObjectPartInfo(part))
	}

	uploadInfo, err := s.minioClient.CompleteMultipartUpload(filePath, uploadId, completeParts)
	if err != nil {
		return nil, fmt.Errorf("unable to complete multipart upload %s: %w", uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Completed multipart upload %s", uploadId))

	createdFile := model.MinioObjectInfoToFile(&uploadInfo.MinioObjectInfo)

	return createdFile, nil
}

func (s *minioFileStorage) AbortMultipartUpload(ctx context.Context, filePath string, uploadId string) error {
	err := s.minioClient.AbortMultipartUpload(filePath, uploadId)
	if err != nil {
		return fmt.Errorf("unable to abort multipart upload %s: %w", uploadId, err)
	}

	logger.TechLog.Info(ctx, fmt.Sprintf("Aborted multipart upload %s", uploadId))
	return nil
}
