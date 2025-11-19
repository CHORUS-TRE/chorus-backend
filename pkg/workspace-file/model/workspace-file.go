package model

import (
	"path"
	"strings"
	"time"

	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
)

type WorkspaceFile struct {
	Path string

	Name        string
	IsDirectory bool
	Size        int64
	MimeType    string

	UpdatedAt time.Time

	Content []byte
}

func MinioObjectInfoToWorkspaceFile(info *miniorawclient.MinioObjectInfo) *WorkspaceFile {
	isDir := strings.HasSuffix(info.Key, "/")
	name := path.Base(strings.TrimRight(info.Key, "/"))

	return &WorkspaceFile{
		Path:        info.Key,
		Name:        name,
		IsDirectory: isDir,
		Size:        info.Size,
		MimeType:    info.MimeType,
		UpdatedAt:   info.LastModified,
	}
}

func MinioObjectToWorkspaceFile(object *miniorawclient.MinioObject) *WorkspaceFile {
	file := MinioObjectInfoToWorkspaceFile(&object.MinioObjectInfo)
	file.Content = object.Content

	return file
}

func WorkspaceFileToMinioObjectInfo(file *WorkspaceFile) *miniorawclient.MinioObjectInfo {
	return &miniorawclient.MinioObjectInfo{
		Key:          file.Path,
		Size:         file.Size,
		LastModified: file.UpdatedAt,
		MimeType:     file.MimeType,
	}
}

func WorkspaceFileToMinioObject(file *WorkspaceFile) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *WorkspaceFileToMinioObjectInfo(file),
		Content:         file.Content,
	}
}

func WorkspaceFileToMinioObjectWithoutContent(file *WorkspaceFile) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *WorkspaceFileToMinioObjectInfo(file),
	}
}
