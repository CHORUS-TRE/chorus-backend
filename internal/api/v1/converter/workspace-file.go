package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
)

func WorkspaceFileToBusiness(file *chorus.WorkspaceFile) (*model.File, error) {
	ua, err := FromProtoTimestamp(file.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &model.File{
		Name:        file.Name,
		Path:        file.Path,
		IsDirectory: file.IsDirectory,
		Size:        file.Size,
		MimeType:    file.MimeType,

		UpdatedAt: ua,

		Content: file.Content,
	}, nil
}

func WorkspaceFileFromBusiness(file *model.File) (*chorus.WorkspaceFile, error) {
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

func WorkspaceFilePartToBusiness(part *chorus.WorkspaceFilePart) (*model.FilePart, error) {
	if part == nil {
		return nil, fmt.Errorf("unable to convert nil workspace file part")
	}

	return &model.FilePart{
		PartNumber: part.PartNumber,
		Data:       part.Data,
		ETag:       part.Etag,
	}, nil
}

func WorkspaceFilePartFromBusiness(part *model.FilePart) (*chorus.WorkspaceFilePart, error) {
	if part == nil {
		return nil, fmt.Errorf("unable to convert nil workspace file part")
	}

	return &chorus.WorkspaceFilePart{
		PartNumber: part.PartNumber,
		Data:       part.Data,
		Etag:       part.ETag,
	}, nil
}
