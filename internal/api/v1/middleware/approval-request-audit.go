package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.ApprovalRequestServiceServer = (*approvalRequestControllerAudit)(nil)

type approvalRequestControllerAudit struct {
	next        chorus.ApprovalRequestServiceServer
	auditWriter service.AuditWriter
}

func NewApprovalRequestAuditMiddleware(auditWriter service.AuditWriter) func(chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
	return func(next chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
		return &approvalRequestControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c approvalRequestControllerAudit) GetApprovalRequest(ctx context.Context, req *chorus.GetApprovalRequestRequest) (*chorus.GetApprovalRequestReply, error) {
	res, err := c.next.GetApprovalRequest(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestRead,
			audit.WithDetail("approval_request_id", req.Id),
			audit.WithDescription("Failed to get approval request."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestRead,
	// 			audit.WithDetail("approval_request_id", req.Id),
	// 			audit.WithDescription(fmt.Sprintf("Retrieved approval request with ID %d.", req.Id)),
	// 		)
	// }

	return res, err
}

func (c approvalRequestControllerAudit) ListApprovalRequests(ctx context.Context, req *chorus.ListApprovalRequestsRequest) (*chorus.ListApprovalRequestsReply, error) {
	res, err := c.next.ListApprovalRequests(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestList,
			audit.WithDetail("filter", req.Filter),
			audit.WithDescription("Failed to list approval requests."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestList,
	// 			audit.WithDetail("filter", req.Filter),
	// 			audit.WithDescription("Listed approval requests."),
	// 			audit.WithDetail("result_count", len(res.Result.ApprovalRequests)),
	// 		)
	// }

	return res, err
}

func (c approvalRequestControllerAudit) CreateDataExtractionRequest(ctx context.Context, req *chorus.CreateDataExtractionRequestRequest) (*chorus.CreateDataExtractionRequestReply, error) {
	res, err := c.next.CreateDataExtractionRequest(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
		audit.WithDetail("file_paths", req.FilePaths),
		audit.WithDetail("title", req.Title),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to create data extraction request."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created data extraction request with ID %d.", res.Result.ApprovalRequest.Id)),
			audit.WithWorkspaceID(req.SourceWorkspaceId),
			audit.WithDetail("approval_request_id", res.Result.ApprovalRequest.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionDataExtractionRequestCreate, opts...)

	return res, err
}

func (c approvalRequestControllerAudit) CreateDataTransferRequest(ctx context.Context, req *chorus.CreateDataTransferRequestRequest) (*chorus.CreateDataTransferRequestReply, error) {
	res, err := c.next.CreateDataTransferRequest(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
		audit.WithDetail("destination_workspace_id", req.DestinationWorkspaceId),
		audit.WithDetail("file_paths", req.FilePaths),
		audit.WithDetail("title", req.Title),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to create data transfer request."),
			audit.WithError(err),
		)
		audit.Record(ctx, c.auditWriter, model.AuditActionDataTransferRequestCreate, opts...)
	} else {
		successOpts := append(opts,
			audit.WithDescription(fmt.Sprintf("Created data transfer request with ID %d from workspace %d to workspace %d.", res.Result.ApprovalRequest.Id, req.SourceWorkspaceId, req.DestinationWorkspaceId)),
			audit.WithDetail("approval_request_id", res.Result.ApprovalRequest.Id),
		)

		// Record for source workspace
		audit.Record(ctx, c.auditWriter, model.AuditActionDataTransferRequestCreate,
			append(successOpts, audit.WithWorkspaceID(req.SourceWorkspaceId))...)

		// Record for destination workspace
		audit.Record(ctx, c.auditWriter, model.AuditActionDataTransferRequestCreate,
			append(successOpts, audit.WithWorkspaceID(req.DestinationWorkspaceId))...)
	}

	return res, err
}

func (c approvalRequestControllerAudit) ApproveApprovalRequest(ctx context.Context, req *chorus.ApproveApprovalRequestRequest) (*chorus.ApproveApprovalRequestReply, error) {
	res, err := c.next.ApproveApprovalRequest(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("approval_request_id", req.Id),
		audit.WithDetail("approve", req.Approve),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to %s approval request %d.", map[bool]string{true: "approve", false: "reject"}[req.Approve], req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("%s approval request with ID %d.", map[bool]string{true: "Approved", false: "Rejected"}[req.Approve], req.Id)),
			audit.WithDetail("requester_id", res.Result.ApprovalRequest.RequesterId),
			audit.WithDetail("approval_request_type", res.Result.ApprovalRequest.Type),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestApprove, opts...)

	return res, err
}

func (c approvalRequestControllerAudit) DeleteApprovalRequest(ctx context.Context, req *chorus.DeleteApprovalRequestRequest) (*chorus.DeleteApprovalRequestReply, error) {
	res, err := c.next.DeleteApprovalRequest(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("approval_request_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to delete approval request."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted approval request with ID %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestDelete, opts...)

	return res, err
}

func (c approvalRequestControllerAudit) DownloadApprovalRequestFile(ctx context.Context, req *chorus.DownloadApprovalRequestFileRequest) (*chorus.DownloadApprovalRequestFileReply, error) {
	res, err := c.next.DownloadApprovalRequestFile(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("approval_request_id", req.Id),
		audit.WithDetail("file_path", req.Path),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to download file from approval request %d.", req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Downloaded file from approval request %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionApprovalRequestFileDownload, opts...)

	return res, err
}
