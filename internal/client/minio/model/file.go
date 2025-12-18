package model

import (
	"path"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/blockstore"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
)

func MinioObjectInfoToFile(info *miniorawclient.MinioObjectInfo) *blockstore.File {
	isDir := strings.HasSuffix(info.Key, "/")
	name := path.Base(strings.TrimRight(info.Key, "/"))

	return &blockstore.File{
		Path:        info.Key,
		Name:        name,
		IsDirectory: isDir,
		Size:        info.Size,
		MimeType:    info.MimeType,
		UpdatedAt:   info.LastModified,
	}
}

func MinioObjectToFile(object *miniorawclient.MinioObject) *blockstore.File {
	file := MinioObjectInfoToFile(&object.MinioObjectInfo)
	file.Content = object.Content

	return file
}

func FileToMinioObjectInfo(file *blockstore.File) *miniorawclient.MinioObjectInfo {
	return &miniorawclient.MinioObjectInfo{
		Key:          file.Path,
		Size:         file.Size,
		LastModified: file.UpdatedAt,
		MimeType:     file.MimeType,
	}
}

func FileToMinioObject(file *blockstore.File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
		Content:         file.Content,
	}
}

func FileToMinioObjectWithoutContent(file *blockstore.File) *miniorawclient.MinioObject {
	return &miniorawclient.MinioObject{
		MinioObjectInfo: *FileToMinioObjectInfo(file),
	}
}

func FilePartToMinioObjectPartInfo(part *blockstore.FilePart) *miniorawclient.MinioObjectPartInfo {
	return &miniorawclient.MinioObjectPartInfo{
		PartNumber: int(part.PartNumber),
		ETag:       part.ETag,
	}
}

func MinioObjectPartInfoToFilePart(partInfo *miniorawclient.MinioObjectPartInfo) *blockstore.FilePart {
	return &blockstore.FilePart{
		PartNumber: uint64(partInfo.PartNumber),
		ETag:       partInfo.ETag,
	}
}

func FilePartToMinioObjectPart(part *blockstore.FilePart) *miniorawclient.MinioObjectPart {
	return &miniorawclient.MinioObjectPart{
		MinioObjectPartInfo: *FilePartToMinioObjectPartInfo(part),
		Data:                part.Data,
	}
}

func MinioObjectPartToFilePart(part *miniorawclient.MinioObjectPart) *blockstore.FilePart {
	return &blockstore.FilePart{
		PartNumber: uint64(part.PartNumber),
		ETag:       part.ETag,
		Data:       part.Data,
	}
}
