package model

import (
	"path"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/miniofilestore/raw-client"
)

func MinioObjectInfoToFile(info *miniorawclient.MinioObjectInfo) *filestore.File {
	isDir := strings.HasSuffix(info.Key, "/")
	name := path.Base(strings.TrimRight(info.Key, "/"))

	return &filestore.File{
		Path:        info.Key,
		Name:        name,
		IsDirectory: isDir,
		Size:        info.Size,
		MimeType:    info.MimeType,
		UpdatedAt:   info.LastModified,
	}
}

func MinioObjectToFile(object *miniorawclient.MinioObject) *filestore.File {
	file := MinioObjectInfoToFile(&object.MinioObjectInfo)
	file.Content = object.Content

	return file
}

func FileToMinioObjectInfo(file *filestore.File) *miniorawclient.MinioObjectInfo {
	return &miniorawclient.MinioObjectInfo{
		Key:          file.Path,
		Size:         file.Size,
		LastModified: file.UpdatedAt,
		MimeType:     file.MimeType,
	}
}

func FileToMinioObject(file *filestore.File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
		Content:         file.Content,
	}
}

func FileToMinioObjectWithoutContent(file *filestore.File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
	}
}

func FilePartToMinioObjectPartInfo(part *filestore.FilePart) *miniorawclient.MinioObjectPartInfo {
	return &miniorawclient.MinioObjectPartInfo{
		PartNumber: int(part.PartNumber),
		ETag:       part.ETag,
	}
}

func MinioObjectPartInfoToFilePart(partInfo *miniorawclient.MinioObjectPartInfo) *filestore.FilePart {
	return &filestore.FilePart{
		PartNumber: uint64(partInfo.PartNumber),
		ETag:       partInfo.ETag,
	}
}

func FilePartToMinioObjectPart(part *filestore.FilePart) *miniorawclient.MinioObjectPart {
	return &miniorawclient.MinioObjectPart{
		MinioObjectPartInfo: *FilePartToMinioObjectPartInfo(part),
		Data:                part.Data,
	}
}

func MinioObjectPartToFilePart(part *miniorawclient.MinioObjectPart) *filestore.FilePart {
	return &filestore.FilePart{
		PartNumber: uint64(part.PartNumber),
		ETag:       part.ETag,
		Data:       part.Data,
	}
}
