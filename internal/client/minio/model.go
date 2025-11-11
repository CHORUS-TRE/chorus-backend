package minio

import (
	"path"
	"strings"
	"time"

	workspace_file_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
)

type MinioObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	MimeType     string
}

type MinioObject struct {
	MinioObjectInfo
	Content []byte
}

func MinioObjectInfoToWorkspaceFile(info *MinioObjectInfo) *workspace_file_model.WorkspaceFile {
	isDir := strings.HasSuffix(info.Key, "/")
	name := path.Base(strings.TrimRight(info.Key, "/"))

	return &workspace_file_model.WorkspaceFile{
		Path:        info.Key,
		Name:        name,
		IsDirectory: isDir,
		Size:        info.Size,
		MimeType:    info.MimeType,
		UpdatedAt:   info.LastModified,
	}
}

func MinioObjectToWorkspaceFile(object *MinioObject) *workspace_file_model.WorkspaceFile {
	file := MinioObjectInfoToWorkspaceFile(&object.MinioObjectInfo)
	file.Content = object.Content

	return file
}

func WorkspaceFileToMinioObjectInfo(file *workspace_file_model.WorkspaceFile) *MinioObjectInfo {
	return &MinioObjectInfo{
		Key:          file.Path,
		Size:         file.Size,
		LastModified: file.UpdatedAt,
		MimeType:     file.MimeType,
	}
}

func WorkspaceFileToMinioObject(file *workspace_file_model.WorkspaceFile) *MinioObject {
	return &MinioObject{
		MinioObjectInfo: *WorkspaceFileToMinioObjectInfo(file),
		Content:         file.Content,
	}
}

func WorkspaceFileToMinioObjectWithoutContent(file *workspace_file_model.WorkspaceFile) *MinioObject {
	return &MinioObject{
		MinioObjectInfo: *WorkspaceFileToMinioObjectInfo(file),
	}
}
