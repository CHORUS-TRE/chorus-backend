package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
)

func WorkspaceFileToBusiness(file *chorus.WorkspaceFile) (*model.WorkspaceFile, error) {
	ua, err := FromProtoTimestamp(file.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &model.WorkspaceFile{
		Name:        file.Name,
		Path:        file.Path,
		IsDirectory: file.IsDirectory,
		Size:        file.Size,
		MimeType:    file.MimeType,

		UpdatedAt: ua,

		Content: file.Content,
	}, nil
}

func WorkspaceFileFromBusiness(file *model.WorkspaceFile) (*chorus.WorkspaceFile, error) {
	if file == nil {
		return nil, fmt.Errorf("unable to convert nil workspace file")
	}

	ua, err := ToProtoTimestamp(file.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.WorkspaceFile{
		Name:        file.Name,
		Path:        file.Path,
		IsDirectory: file.IsDirectory,
		Size:        file.Size,
		MimeType:    file.MimeType,

		UpdatedAt: ua,

		Content: file.Content,
	}, nil
}
