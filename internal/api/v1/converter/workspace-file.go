package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
)

func WorkspaceFileToBusiness(file *chorus.WorkspaceFile) (*filestore.File, error) {
	ua, err := FromProtoTimestamp(file.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &filestore.File{
		Name:        file.Name,
		Path:        file.Path,
		IsDirectory: file.IsDirectory,
		Size:        file.Size,
		MimeType:    file.MimeType,

		UpdatedAt: ua,

		Content: file.Content,
	}, nil
}

func WorkspaceFileFromBusiness(file *filestore.File) (*chorus.WorkspaceFile, error) {
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

func WorkspaceFilePartToBusiness(part *chorus.WorkspaceFilePart) (*filestore.FilePart, error) {
	if part == nil {
		return nil, fmt.Errorf("unable to convert nil workspace file part")
	}

	return &filestore.FilePart{
		PartNumber: part.PartNumber,
		Data:       part.Data,
		ETag:       part.Etag,
	}, nil
}

func WorkspaceFilePartFromBusiness(part *filestore.FilePart) (*chorus.WorkspaceFilePart, error) {
	if part == nil {
		return nil, fmt.Errorf("unable to convert nil workspace file part")
	}

	return &chorus.WorkspaceFilePart{
		PartNumber: part.PartNumber,
		Data:       part.Data,
		Etag:       part.ETag,
	}, nil
}
