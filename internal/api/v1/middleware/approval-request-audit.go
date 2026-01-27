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
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestRead,
			audit.WithDescription("Failed to get approval request."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("approval_request_id", req.Id),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionApprovalRequestRead,
	// 		audit.WithDescription(fmt.Sprintf("Retrieved approval request with ID %d.", req.Id)),
	// 		audit.WithDetail("approval_request_id", req.Id),
	// 	)
	// }

	return res, err
}

func (c approvalRequestControllerAudit) ListApprovalRequests(ctx context.Context, req *chorus.ListApprovalRequestsRequest) (*chorus.ListApprovalRequestsReply, error) {
	res, err := c.next.ListApprovalRequests(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestList,
			audit.WithDescription("Failed to list approval requests."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("filter", req.Filter),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionApprovalRequestList,
	// 		audit.WithDescription("Listed approval requests."),
	// 		audit.WithDetail("result_count", len(res.Result.ApprovalRequests)),
	// 	)
	// }

	return res, err
}

func (c approvalRequestControllerAudit) CreateDataExtractionRequest(ctx context.Context, req *chorus.CreateDataExtractionRequestRequest) (*chorus.CreateDataExtractionRequestReply, error) {
	res, err := c.next.CreateDataExtractionRequest(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionDataExtractionRequestCreate,
			audit.WithDescription("Failed to create data extraction request."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
			audit.WithDetail("title", req.Title),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionDataExtractionRequestCreate,
			audit.WithDescription(fmt.Sprintf("Created data extraction request with ID %d.", res.Result.ApprovalRequest.Id)),
			audit.WithWorkspaceID(req.SourceWorkspaceId),
			audit.WithDetail("approval_request_id", res.Result.ApprovalRequest.Id),
			audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
			audit.WithDetail("title", req.Title),
		)
	}

	return res, err
}

func (c approvalRequestControllerAudit) CreateDataTransferRequest(ctx context.Context, req *chorus.CreateDataTransferRequestRequest) (*chorus.CreateDataTransferRequestReply, error) {
	res, err := c.next.CreateDataTransferRequest(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionDataTransferRequestCreate,
			audit.WithDescription("Failed to create data transfer request."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
			audit.WithDetail("destination_workspace_id", req.DestinationWorkspaceId),
			audit.WithDetail("title", req.Title),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionDataTransferRequestCreate,
			audit.WithDescription(fmt.Sprintf("Created data transfer request with ID %d from workspace %d to workspace %d.", res.Result.ApprovalRequest.Id, req.SourceWorkspaceId, req.DestinationWorkspaceId)),
			audit.WithWorkspaceID(req.SourceWorkspaceId),
			audit.WithDetail("approval_request_id", res.Result.ApprovalRequest.Id),
			audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
			audit.WithDetail("destination_workspace_id", req.DestinationWorkspaceId),
			audit.WithDetail("title", res.Result.ApprovalRequest.Title),
		)
		// Record again for destination workspace
		audit.Record(ctx, c.auditWriter,
			model.AuditActionDataTransferRequestCreate,
			audit.WithDescription(fmt.Sprintf("Created data transfer request with ID %d from workspace %d to workspace %d.", res.Result.ApprovalRequest.Id, req.SourceWorkspaceId, req.DestinationWorkspaceId)),
			audit.WithWorkspaceID(req.DestinationWorkspaceId),
			audit.WithDetail("approval_request_id", res.Result.ApprovalRequest.Id),
			audit.WithDetail("source_workspace_id", req.SourceWorkspaceId),
			audit.WithDetail("destination_workspace_id", req.DestinationWorkspaceId),
			audit.WithDetail("title", res.Result.ApprovalRequest.Title),
		)
	}

	return res, err
}

func (c approvalRequestControllerAudit) ApproveApprovalRequest(ctx context.Context, req *chorus.ApproveApprovalRequestRequest) (*chorus.ApproveApprovalRequestReply, error) {
	res, err := c.next.ApproveApprovalRequest(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestApprove,
			audit.WithDescription(fmt.Sprintf("Failed to %s approval request %d.", map[bool]string{true: "approve", false: "reject"}[req.Approve], req.Id)),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("approval_request_id", req.Id),
			audit.WithDetail("approve", req.Approve),
			audit.WithDetail("comment", req.Comment),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestApprove,
			audit.WithDescription(fmt.Sprintf("%s approval request with ID %d.", map[bool]string{true: "Approved", false: "Rejected"}[req.Approve], req.Id)),
			audit.WithDetail("approval_request_id", req.Id),
			audit.WithDetail("approve", req.Approve),
			audit.WithDetail("title", res.Result.ApprovalRequest.Title),
			audit.WithDetail("description", res.Result.ApprovalRequest.Description),
			audit.WithDetail("comment", req.Comment),
		)
	}

	return res, err
}

func (c approvalRequestControllerAudit) DeleteApprovalRequest(ctx context.Context, req *chorus.DeleteApprovalRequestRequest) (*chorus.DeleteApprovalRequestReply, error) {
	res, err := c.next.DeleteApprovalRequest(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestDelete,
			audit.WithDescription("Failed to delete approval request."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("approval_request_id", req.Id),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionApprovalRequestDelete,
			audit.WithDescription(fmt.Sprintf("Deleted approval request with ID %d.", req.Id)),
			audit.WithDetail("approval_request_id", req.Id),
		)
	}

	return res, err
}
