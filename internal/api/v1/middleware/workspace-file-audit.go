package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.WorkspaceFileServiceServer = (*workspaceFileControllerAudit)(nil)

type workspaceFileControllerAudit struct {
	next        chorus.WorkspaceFileServiceServer
	auditWriter service.AuditWriter
}

func NewWorkspaceFileAuditMiddleware(auditWriter service.AuditWriter) func(chorus.WorkspaceFileServiceServer) chorus.WorkspaceFileServiceServer {
	return func(next chorus.WorkspaceFileServiceServer) chorus.WorkspaceFileServiceServer {
		return &workspaceFileControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c workspaceFileControllerAudit) GetWorkspaceFile(ctx context.Context, req *chorus.GetWorkspaceFileRequest) (*chorus.GetWorkspaceFileReply, error) {
	res, err := c.next.GetWorkspaceFile(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileRead,
			audit.WithDescription(fmt.Sprintf("Failed to get file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileRead,
			audit.WithDescription(fmt.Sprintf("Retrieved file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) ListWorkspaceFiles(ctx context.Context, req *chorus.ListWorkspaceFilesRequest) (*chorus.ListWorkspaceFilesReply, error) {
	res, err := c.next.ListWorkspaceFiles(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileList,
			audit.WithDescription(fmt.Sprintf("Failed to list workspace %d files at %s.", req.WorkspaceId, req.Path)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileList,
			audit.WithDescription(fmt.Sprintf("Listed %d files in workspace %d at %s.", len(res.Result.Files), req.WorkspaceId, req.Path)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
			audit.WithDetail("result_count", len(res.Result.Files)),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) CreateWorkspaceFile(ctx context.Context, req *chorus.CreateWorkspaceFileRequest) (*chorus.CreateWorkspaceFileReply, error) {
	res, err := c.next.CreateWorkspaceFile(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileCreate,
			audit.WithDescription(fmt.Sprintf("Failed to upload file %s in workspace %d.", req.File.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.File.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileCreate,
			audit.WithDescription(fmt.Sprintf("Uploaded file %s in workspace %d.", req.File.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.File.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) UpdateWorkspaceFile(ctx context.Context, req *chorus.UpdateWorkspaceFileRequest) (*chorus.UpdateWorkspaceFileReply, error) {
	res, err := c.next.UpdateWorkspaceFile(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUpdate,
			audit.WithDescription(fmt.Sprintf("Failed to update file %s in workspace %d.", req.File.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("old_path", req.OldPath),
			audit.WithDetail("new_path", req.File.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUpdate,
			audit.WithDescription(fmt.Sprintf("Updated file %s in workspace %d.", req.File.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("old_path", req.OldPath),
			audit.WithDetail("new_path", req.File.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) DeleteWorkspaceFile(ctx context.Context, req *chorus.DeleteWorkspaceFileRequest) (*chorus.DeleteWorkspaceFileReply, error) {
	res, err := c.next.DeleteWorkspaceFile(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileDelete,
			audit.WithDescription(fmt.Sprintf("Failed to delete file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileDelete,
			audit.WithDescription(fmt.Sprintf("Deleted file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) InitiateWorkspaceFileUpload(ctx context.Context, req *chorus.InitiateWorkspaceFileUploadRequest) (*chorus.InitiateWorkspaceFileUploadReply, error) {
	res, err := c.next.InitiateWorkspaceFileUpload(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadInitiate,
			audit.WithDescription(fmt.Sprintf("Failed to initiate upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadInitiate,
			audit.WithDescription(fmt.Sprintf("Initiated upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) UploadWorkspaceFilePart(ctx context.Context, req *chorus.UploadWorkspaceFilePartRequest) (*chorus.UploadWorkspaceFilePartReply, error) {
	// No audit recording for individual file parts to avoid excessive log volume
	return c.next.UploadWorkspaceFilePart(ctx, req)
}

func (c workspaceFileControllerAudit) CompleteWorkspaceFileUpload(ctx context.Context, req *chorus.CompleteWorkspaceFileUploadRequest) (*chorus.CompleteWorkspaceFileUploadReply, error) {
	res, err := c.next.CompleteWorkspaceFileUpload(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadComplete,
			audit.WithDescription(fmt.Sprintf("Failed to complete upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadComplete,
			audit.WithDescription(fmt.Sprintf("Completed upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	}

	return res, err
}

func (c workspaceFileControllerAudit) AbortWorkspaceFileUpload(ctx context.Context, req *chorus.AbortWorkspaceFileUploadRequest) (*chorus.AbortWorkspaceFileUploadReply, error) {
	res, err := c.next.AbortWorkspaceFileUpload(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadAbort,
			audit.WithDescription(fmt.Sprintf("Failed to abort upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionFileUploadAbort,
			audit.WithDescription(fmt.Sprintf("Aborted upload for file %s in workspace %d.", req.Path, req.WorkspaceId)),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.WorkspaceId),
			audit.WithDetail("path", req.Path),
		)
	}

	return res, err
}
