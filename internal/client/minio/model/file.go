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

func FilePartToMinioObjectPartInfo(part *FilePart) *miniorawclient.MinioObjectPartInfo {
	return &miniorawclient.MinioObjectPartInfo{
		PartNumber: int(part.PartNumber),
		ETag:       part.ETag,
	}
}

func MinioObjectPartInfoToFilePart(partInfo *miniorawclient.MinioObjectPartInfo) *FilePart {
	return &FilePart{
		PartNumber: uint64(partInfo.PartNumber),
		ETag:       partInfo.ETag,
	}
}

func FilePartToMinioObjectPart(part *FilePart) *miniorawclient.MinioObjectPart {
	return &miniorawclient.MinioObjectPart{
		MinioObjectPartInfo: *FilePartToMinioObjectPartInfo(part),
		Data:                part.Data,
	}
}

func MinioObjectPartToFilePart(part *miniorawclient.MinioObjectPart) *FilePart {
	return &FilePart{
		PartNumber: uint64(part.PartNumber),
		ETag:       part.ETag,
		Data:       part.Data,
	}
}
