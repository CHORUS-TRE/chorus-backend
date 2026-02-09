package middleware

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	approval_request_model "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.ApprovalRequestServiceServer = (*approvalRequestControllerAuthorization)(nil)

type ApprovalRequestResolver interface {
	GetApprovalRequest(ctx context.Context, tenantID uint64, requestID uint64) (*approval_request_model.ApprovalRequest, error)
}

type approvalRequestControllerAuthorization struct {
	Authorization
	resolver ApprovalRequestResolver
	next     chorus.ApprovalRequestServiceServer
}

func ApprovalRequestAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher, resolver ApprovalRequestResolver) func(chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
	return func(next chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
		return &approvalRequestControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
				cfg:        cfg,
				refresher:  refresher,
			},
			resolver: resolver,
			next:     next,
		}
	}
}

func (c approvalRequestControllerAuthorization) GetApprovalRequest(ctx context.Context, req *chorus.GetApprovalRequestRequest) (*chorus.GetApprovalRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetApprovalRequest(ctx, req)
}

func (c approvalRequestControllerAuthorization) ListApprovalRequests(ctx context.Context, req *chorus.ListApprovalRequestsRequest) (*chorus.ListApprovalRequestsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListRequests)
	if err != nil {
		return nil, err
	}

	return c.next.ListApprovalRequests(ctx, req)
}

func (c approvalRequestControllerAuthorization) CreateDataExtractionRequest(ctx context.Context, req *chorus.CreateDataExtractionRequestRequest) (*chorus.CreateDataExtractionRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateRequest, authorization.WithWorkspace(req.SourceWorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.CreateDataExtractionRequest(ctx, req)
}

func (c approvalRequestControllerAuthorization) CreateDataTransferRequest(ctx context.Context, req *chorus.CreateDataTransferRequestRequest) (*chorus.CreateDataTransferRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateRequest, authorization.WithWorkspace(req.SourceWorkspaceId))
	if err != nil {
		return nil, err
	}

	err = c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(req.DestinationWorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.CreateDataTransferRequest(ctx, req)
}

func (c approvalRequestControllerAuthorization) ApproveApprovalRequest(ctx context.Context, req *chorus.ApproveApprovalRequestRequest) (*chorus.ApproveApprovalRequestReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unable to extract tenant ID")
	}

	approvalRequest, err := c.resolver.GetApprovalRequest(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "approval request not found")
	}

	switch approvalRequest.Type {
	case approval_request_model.ApprovalRequestTypeDataExtraction:
		workspaceID := approvalRequest.GetSourceWorkspaceID()
		err = c.IsAuthorized(ctx, authorization.PermissionDownloadFilesFromWorkspace, authorization.WithWorkspace(workspaceID))
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "user does not have permission to approve requests for this workspace")
		}
	case approval_request_model.ApprovalRequestTypeDataTransfer:
		workspaceID := approvalRequest.GetSourceWorkspaceID()
		err = c.IsAuthorized(ctx, authorization.PermissionDownloadFilesFromWorkspace, authorization.WithWorkspace(workspaceID))
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "user does not have permission to approve requests for this workspace")
		}

		targetWorkspaceID := approvalRequest.Details.DataTransferDetails.DestinationWorkspaceID
		err = c.IsAuthorized(ctx, authorization.PermissionUploadFilesToWorkspace, authorization.WithWorkspace(targetWorkspaceID))
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "user does not have permission to approve requests for this workspace")
		}
	default:
		return nil, status.Error(codes.Internal, "unable to determine source workspace for approval request")
	}

	return c.next.ApproveApprovalRequest(ctx, req)
}

func (c approvalRequestControllerAuthorization) DeleteApprovalRequest(ctx context.Context, req *chorus.DeleteApprovalRequestRequest) (*chorus.DeleteApprovalRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteApprovalRequest(ctx, req)
}

func (c approvalRequestControllerAuthorization) DownloadApprovalRequestFile(ctx context.Context, req *chorus.DownloadApprovalRequestFileRequest) (*chorus.DownloadApprovalRequestFileReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unable to extract tenant ID")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unable to extract user ID")
	}

	approvalRequest, err := c.resolver.GetApprovalRequest(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "approval request not found")
	}

	if approvalRequest.Type != approval_request_model.ApprovalRequestTypeDataExtraction {
		return nil, status.Error(codes.PermissionDenied, "only data extraction requests have downloadable files")
	}

	if approvalRequest.RequesterID != userID {
		workspaceID := approvalRequest.GetSourceWorkspaceID()
		err = c.IsAuthorized(ctx, authorization.PermissionDownloadFilesFromWorkspace, authorization.WithWorkspace(workspaceID))
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "user does not have permission to approve requests for this workspace")
		}
	}

	return c.next.DownloadApprovalRequestFile(ctx, req)
}
