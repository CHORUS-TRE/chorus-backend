package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.RequestServiceServer = (*requestControllerAuthorization)(nil)

type requestControllerAuthorization struct {
	Authorization
	next chorus.RequestServiceServer
}

func RequestAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer, cfg config.Config, refresher Refresher) func(chorus.RequestServiceServer) chorus.RequestServiceServer {
	return func(next chorus.RequestServiceServer) chorus.RequestServiceServer {
		return &requestControllerAuthorization{
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

func (c requestControllerAuthorization) GetRequest(ctx context.Context, req *chorus.GetRequestRequest) (*chorus.GetRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetRequest(ctx, req)
}

func (c requestControllerAuthorization) ListRequests(ctx context.Context, req *chorus.ListRequestsRequest) (*chorus.ListRequestsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListRequests)
	if err != nil {
		return nil, err
	}

	return c.next.ListRequests(ctx, req)
}

func (c requestControllerAuthorization) CreateRequest(ctx context.Context, req *chorus.CreateRequestRequest) (*chorus.CreateRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateRequest, authorization.WithWorkspace(req.SourceWorkspaceId))
	if err != nil {
		return nil, err
	}

	if req.DestinationWorkspaceId != nil {
		err = c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(*req.DestinationWorkspaceId))
		if err != nil {
			return nil, err
		}
	}

	return c.next.CreateRequest(ctx, req)
}

func (c requestControllerAuthorization) ApproveRequest(ctx context.Context, req *chorus.ApproveRequestRequest) (*chorus.ApproveRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionApproveRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ApproveRequest(ctx, req)
}

func (c requestControllerAuthorization) DeleteRequest(ctx context.Context, req *chorus.DeleteRequestRequest) (*chorus.DeleteRequestReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteRequest, authorization.WithRequest(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteRequest(ctx, req)
}
