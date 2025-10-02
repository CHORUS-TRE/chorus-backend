package minio

import (
	"path"
	"regexp"
	"strings"

	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/minio/minio-go/v7"
)

const workspacePrefixPattern = `^workspaces/workspace\d+/`

func (c *client) ObjectToWorkspaceFile(objectInfo minio.ObjectInfo) (workspace_model.WorkspaceFile, error) {
	isDir := strings.HasSuffix(objectInfo.Key, "/")
	name := path.Base(strings.TrimRight(objectInfo.Key, "/"))

	// Trim prefix to get workspace-relative path
	workspacePattern := regexp.MustCompile(workspacePrefixPattern)
	relativePath := workspacePattern.ReplaceAllString(objectInfo.Key, "")

	return workspace_model.WorkspaceFile{
		Path:        relativePath,
		Name:        name,
		IsDirectory: isDir,
		Size:        objectInfo.Size,
		MimeType:    objectInfo.ContentType,
		UpdatedAt:   objectInfo.LastModified,
	}, nil
}
