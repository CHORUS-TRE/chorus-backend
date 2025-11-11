package model

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
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

const workspacePrefix = "workspaces/workspace"
const workspacePrefixPattern = `^` + workspacePrefix + `\d+/`

func WorkspacePathToObjectKey(workspaceID uint64, filePath string) string {
	objectKey := fmt.Sprintf("%s%d/%s", workspacePrefix, workspaceID, strings.TrimPrefix(filePath, "/"))
	return objectKey
}

func objectKeyToWorkspacePath(objectKey string) string {
	pattern := regexp.MustCompile(workspacePrefixPattern)
	return pattern.ReplaceAllString(objectKey, "")
}

func ObjectToWorkspaceFile(objectInfo minio.ObjectInfo) (*WorkspaceFile, error) {
	isDir := strings.HasSuffix(objectInfo.Key, "/")
	name := path.Base(strings.TrimRight(objectInfo.Key, "/"))

	return &WorkspaceFile{
		Path:        objectKeyToWorkspacePath(objectInfo.Key),
		Name:        name,
		IsDirectory: isDir,
		Size:        objectInfo.Size,
		MimeType:    objectInfo.ContentType,
		UpdatedAt:   objectInfo.LastModified,
	}, nil
}

func FormatPrefix(prefix string) string {
	normalized := "/" + strings.TrimPrefix(prefix, "/")
	normalized = strings.TrimSuffix(normalized, "/") + "/"
	return normalized
}
