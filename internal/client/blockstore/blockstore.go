package blockstore

import (
	"context"
	"time"
)

type File struct {
	Path string

	Name        string
	IsDirectory bool
	Size        uint64
	MimeType    string

	UpdatedAt time.Time

	Content []byte
}

type FilePart struct {
	PartNumber uint64
	Data       []byte
	ETag       string
}

type FileUploadInfo struct {
	UploadID   string
	PartSize   uint64
	TotalParts uint64
}

// The MinioFileStore interface abstracts UNIX-like file operations which can be performed on a MinIO object storage.
// The directories are represented as objects with keys ending in a '/' character.
type MinioFileStore interface {
	// Get file metadata at the specified path without downloading the content.
	StatFile(ctx context.Context, path string) (*File, error)

	// Get the file at the specified path, including its content.
	GetFile(ctx context.Context, path string) (*File, error)

	// List files and directories at the specified path.
	ListFiles(ctx context.Context, path string) ([]*File, error)

	// Create a new file at the specified path.
	CreateFile(ctx context.Context, file *File) (*File, error)

	// Create a new directory at the specified path.
	CreateDirectory(ctx context.Context, file *File) (*File, error)

	// Move a file from oldPath to newPath.
	MoveFile(ctx context.Context, oldPath string, newPath string) (*File, error)

	// Delete a file at the specified path.
	DeleteFile(ctx context.Context, path string) error

	// Delete a directory and all its contents recursively.
	DeleteDirectory(ctx context.Context, path string) error

	// Initiate a new multipart upload for a file.
	InitiateMultipartUpload(ctx context.Context, file *File) (*FileUploadInfo, error)

	// Upload a single part of a multipart upload.
	UploadPart(ctx context.Context, path string, uploadId string, part *FilePart) (*FilePart, error)

	// Complete a multipart upload after all parts of a file have been uploaded.
	CompleteMultipartUpload(ctx context.Context, path string, uploadId string, parts []*FilePart) (*File, error)

	// Abort a multipart upload, discarding all uploaded parts.
	AbortMultipartUpload(ctx context.Context, path string, uploadId string) error
}
