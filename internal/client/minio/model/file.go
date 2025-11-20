package model

import (
	"path"
	"strings"
	"time"

	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
)

type File struct {
	Path string

	Name        string
	IsDirectory bool
	Size        int64
	MimeType    string

	UpdatedAt time.Time

	Content []byte
}

func MinioObjectInfoToFile(info *miniorawclient.MinioObjectInfo) *File {
	isDir := strings.HasSuffix(info.Key, "/")
	name := path.Base(strings.TrimRight(info.Key, "/"))

	return &File{
		Path:        info.Key,
		Name:        name,
		IsDirectory: isDir,
		Size:        info.Size,
		MimeType:    info.MimeType,
		UpdatedAt:   info.LastModified,
	}
}

func MinioObjectToFile(object *miniorawclient.MinioObject) *File {
	file := MinioObjectInfoToFile(&object.MinioObjectInfo)
	file.Content = object.Content

	return file
}

func FileToMinioObjectInfo(file *File) *miniorawclient.MinioObjectInfo {
	return &miniorawclient.MinioObjectInfo{
		Key:          file.Path,
		Size:         file.Size,
		LastModified: file.UpdatedAt,
		MimeType:     file.MimeType,
	}
}

func FileToMinioObject(file *File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
		Content:         file.Content,
	}
}

func FileToMinioObjectWithoutContent(file *File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
	}
}
