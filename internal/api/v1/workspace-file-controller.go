package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
)

var _ chorus.WorkspaceFileServiceServer = (*WorkspaceFileController)(nil)

// WorkspaceFileController is the workspace file service controller handler.
type WorkspaceFileController struct {
	workspaceFile service.WorkspaceFiler
}

// NewWorkspaceFileController returns a fresh admin service controller instance.
func NewWorkspaceFileController(workspaceFile service.WorkspaceFiler) WorkspaceFileController {
	return WorkspaceFileController{workspaceFile: workspaceFile}
}

func (c WorkspaceFileController) ListWorkspaceFileStores(ctx context.Context, req *chorus.ListWorkspaceFileStoresRequest) (*chorus.ListWorkspaceFileStoresReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	stores, err := c.workspaceFile.ListWorkspaceFileStores(ctx, req.WorkspaceId)
	if err != nil {
		return nil, err
	}

	tgStores := make([]*chorus.WorkspaceFileStoreInfo, 0, len(stores))
	for _, store := range stores {
		tgStores = append(tgStores, converter.WorkspaceFileStoreInfoFromBusiness(store))
	}

	return &chorus.ListWorkspaceFileStoresReply{Result: &chorus.ListWorkspaceFileStoresResult{Stores: tgStores}}, nil
}

func (c WorkspaceFileController) GetWorkspaceFile(ctx context.Context, req *chorus.GetWorkspaceFileRequest) (*chorus.GetWorkspaceFileReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	file, err := c.workspaceFile.GetWorkspaceFile(ctx, req.WorkspaceId, req.Path)
	if err != nil {
		return nil, err
	}

	tgFile, err := converter.WorkspaceFileFromBusiness(file)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	return &chorus.GetWorkspaceFileReply{Result: &chorus.GetWorkspaceFileResult{File: tgFile}}, nil
}

func (c WorkspaceFileController) ListWorkspaceFiles(ctx context.Context, req *chorus.ListWorkspaceFilesRequest) (*chorus.ListWorkspaceFilesReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	files, err := c.workspaceFile.ListWorkspaceFiles(ctx, req.WorkspaceId, req.Path)
	if err != nil {
		return nil, err
	}

	tgFiles := make([]*chorus.WorkspaceFile, 0, len(files))
	for _, file := range files {
		tgFile, err := converter.WorkspaceFileFromBusiness(file)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
		}
		tgFiles = append(tgFiles, tgFile)
	}

	return &chorus.ListWorkspaceFilesReply{Result: &chorus.ListWorkspaceFilesResult{Files: tgFiles}}, nil
}

func (c WorkspaceFileController) CreateWorkspaceFile(ctx context.Context, req *chorus.CreateWorkspaceFileRequest) (*chorus.CreateWorkspaceFileReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	file, err := converter.WorkspaceFileToBusiness(req.File)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	newFile, err := c.workspaceFile.CreateWorkspaceFile(ctx, req.WorkspaceId, file)
	if err != nil {
		return nil, err
	}

	tgFile, err := converter.WorkspaceFileFromBusiness(newFile)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	return &chorus.CreateWorkspaceFileReply{Result: &chorus.CreateWorkspaceFileResult{File: tgFile}}, nil
}

func (c WorkspaceFileController) UpdateWorkspaceFile(ctx context.Context, req *chorus.UpdateWorkspaceFileRequest) (*chorus.UpdateWorkspaceFileReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	file, err := converter.WorkspaceFileToBusiness(req.File)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	updatedFile, err := c.workspaceFile.UpdateWorkspaceFile(ctx, req.WorkspaceId, req.OldPath, file)
	if err != nil {
		return nil, err
	}

	tgFile, err := converter.WorkspaceFileFromBusiness(updatedFile)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	return &chorus.UpdateWorkspaceFileReply{Result: &chorus.UpdateWorkspaceFileResult{File: tgFile}}, nil
}

func (c WorkspaceFileController) DeleteWorkspaceFile(ctx context.Context, req *chorus.DeleteWorkspaceFileRequest) (*chorus.DeleteWorkspaceFileReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	err := c.workspaceFile.DeleteWorkspaceFile(ctx, req.WorkspaceId, req.Path)
	if err != nil {
		return nil, err
	}

	return &chorus.DeleteWorkspaceFileReply{Result: &chorus.DeleteWorkspaceFileResult{}}, nil
}

func (c WorkspaceFileController) InitiateWorkspaceFileUpload(ctx context.Context, req *chorus.InitiateWorkspaceFileUploadRequest) (*chorus.InitiateWorkspaceFileUploadReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	file, err := converter.WorkspaceFileToBusiness(req.File)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	uploadInfo, err := c.workspaceFile.InitiateWorkspaceFileUpload(ctx, req.WorkspaceId, req.Path, file)
	if err != nil {
		return nil, err
	}

	return &chorus.InitiateWorkspaceFileUploadReply{Result: &chorus.InitiateWorkspaceFileUploadResult{
		UploadId:   uploadInfo.UploadID,
		PartSize:   uploadInfo.PartSize,
		TotalParts: uploadInfo.TotalParts,
	}}, nil
}

func (c WorkspaceFileController) UploadWorkspaceFilePart(ctx context.Context, req *chorus.UploadWorkspaceFilePartRequest) (*chorus.UploadWorkspaceFilePartReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	filePart, err := converter.WorkspaceFilePartToBusiness(req.Part)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file part")
	}

	uploadedPart, err := c.workspaceFile.UploadWorkspaceFilePart(ctx, req.WorkspaceId, req.Path, req.UploadId, filePart)
	if err != nil {
		return nil, err
	}

	part, err := converter.WorkspaceFilePartFromBusiness(uploadedPart)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file part")
	}

	return &chorus.UploadWorkspaceFilePartReply{Result: &chorus.UploadWorkspaceFilePartResult{Part: part}}, nil
}

func (c WorkspaceFileController) CompleteWorkspaceFileUpload(ctx context.Context, req *chorus.CompleteWorkspaceFileUploadRequest) (*chorus.CompleteWorkspaceFileUploadReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	var parts []*filestore.FilePart
	for _, tgPart := range req.Parts {
		part, err := converter.WorkspaceFilePartToBusiness(tgPart)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file part")
		}
		parts = append(parts, part)
	}

	uploadedFile, err := c.workspaceFile.CompleteWorkspaceFileUpload(ctx, req.WorkspaceId, req.Path, req.UploadId, parts)
	if err != nil {
		return nil, err
	}

	tgFile, err := converter.WorkspaceFileFromBusiness(uploadedFile)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workspace file")
	}

	return &chorus.CompleteWorkspaceFileUploadReply{Result: &chorus.CompleteWorkspaceFileUploadResult{File: tgFile}}, nil
}

func (c WorkspaceFileController) AbortWorkspaceFileUpload(ctx context.Context, req *chorus.AbortWorkspaceFileUploadRequest) (*chorus.AbortWorkspaceFileUploadReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	err := c.workspaceFile.AbortWorkspaceFileUpload(ctx, req.WorkspaceId, req.Path, req.UploadId)
	if err != nil {
		return nil, err
	}

	return &chorus.AbortWorkspaceFileUploadReply{Result: &chorus.AbortWorkspaceFileUploadResult{}}, nil
}
