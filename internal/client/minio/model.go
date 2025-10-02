package minio

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/minio/minio-go/v7"
)

const WORKSPACE_PREFIX = "workspaces/workspace"
const workspacePrefixPattern = `^` + WORKSPACE_PREFIX + `\d+/`

func (c *client) ObjectToWorkspaceFile(objectInfo minio.ObjectInfo) (workspace_model.WorkspaceFile, error) {
	isDir := strings.HasSuffix(objectInfo.Key, "/")
	name := path.Base(strings.TrimRight(objectInfo.Key, "/"))

	return workspace_model.WorkspaceFile{
		Path:        ObjectKeyToWorkspacePath(objectInfo.Key),
		Name:        name,
		IsDirectory: isDir,
		Size:        objectInfo.Size,
		MimeType:    objectInfo.ContentType,
		UpdatedAt:   objectInfo.LastModified,
	}, nil
}

func WorkspacePathToObjectKey(workspaceID uint64, filePath string) string {
	return fmt.Sprintf("%s%d/%s", WORKSPACE_PREFIX, workspaceID, strings.TrimPrefix(filePath, "/"))
}

func ObjectKeyToWorkspacePath(objectKey string) string {
	workspacePattern := regexp.MustCompile(workspacePrefixPattern)
	return workspacePattern.ReplaceAllString(objectKey, "")
}
