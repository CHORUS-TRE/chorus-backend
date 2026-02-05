package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.ApprovalRequestServiceServer = (*approvalRequestControllerAuthorization)(nil)

type approvalRequestControllerAuthorization struct {
	Authorization
	next chorus.ApprovalRequestServiceServer
}

func ApprovalRequestAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher) func(chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
	return func(next chorus.ApprovalRequestServiceServer) chorus.ApprovalRequestServiceServer {
		return &approvalRequestControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
				cfg:        cfg,
				refresher:  refresher,
			},
			next: next,
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
	err := c.IsAuthorized(ctx, authorization.PermissionApproveRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
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
