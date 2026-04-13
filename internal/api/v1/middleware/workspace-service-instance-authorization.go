package middleware

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.WorkspaceServiceInstanceServiceServer = (*workspaceServiceInstanceControllerAuthorization)(nil)

type workspaceServiceInstanceControllerAuthorization struct {
	Authorization
	next chorus.WorkspaceServiceInstanceServiceServer
}

func WorkspaceServiceInstanceAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer) func(chorus.WorkspaceServiceInstanceServiceServer) chorus.WorkspaceServiceInstanceServiceServer {
	return func(next chorus.WorkspaceServiceInstanceServiceServer) chorus.WorkspaceServiceInstanceServiceServer {
		return &workspaceServiceInstanceControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c workspaceServiceInstanceControllerAuthorization) ListWorkspaceServiceInstances(ctx context.Context, req *chorus.ListWorkspaceServiceInstancesRequest) (*chorus.ListWorkspaceServiceInstancesReply, error) {
	if req.Filter != nil && len(req.Filter.WorkspaceIdsIn) > 0 {
		for _, id := range req.Filter.WorkspaceIdsIn {
			err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(id))
			if err != nil {
				return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("not authorized to access workspace %d", id))
			}
		}
	} else {
		attrs, err := c.GetContextListForPermission(ctx, authorization.PermissionGetWorkspace)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get context list for permission: %v", err.Error()))
		}

		if len(attrs) == 0 {
			return &chorus.ListWorkspaceServiceInstancesReply{Result: &chorus.ListWorkspaceServiceInstancesResult{WorkspaceServiceInstances: []*chorus.WorkspaceServiceInstance{}}}, nil
		}

		for _, attr := range attrs {
			if workspaceIDStr, ok := attr[authorization.RoleContextWorkspace]; ok {
				if workspaceIDStr == "" {
					continue
				}
				if workspaceIDStr == "*" {
					req.Filter = nil
					return c.next.ListWorkspaceServiceInstances(ctx, req)
				}
				if req.Filter == nil {
					req.Filter = &chorus.WorkspaceServiceInstanceFilter{}
				}
				workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
				if err != nil {
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to parse workspace ID from context: %v", err.Error()))
				}
				req.Filter.WorkspaceIdsIn = append(req.Filter.WorkspaceIdsIn, workspaceID)
			}
		}
	}

	return c.next.ListWorkspaceServiceInstances(ctx, req)
}

func (c workspaceServiceInstanceControllerAuthorization) GetWorkspaceServiceInstance(ctx context.Context, req *chorus.GetWorkspaceServiceInstanceRequest) (*chorus.GetWorkspaceServiceInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspaceServiceInstance)
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkspaceServiceInstance(ctx, req)
}

func (c workspaceServiceInstanceControllerAuthorization) CreateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.CreateWorkspaceServiceInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateWorkspaceServiceInstance, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.CreateWorkspaceServiceInstance(ctx, req)
}

func (c workspaceServiceInstanceControllerAuthorization) UpdateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.UpdateWorkspaceServiceInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkspaceServiceInstance, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkspaceServiceInstance(ctx, req)
}

func (c workspaceServiceInstanceControllerAuthorization) DeleteWorkspaceServiceInstance(ctx context.Context, req *chorus.DeleteWorkspaceServiceInstanceRequest) (*chorus.DeleteWorkspaceServiceInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteWorkspaceServiceInstance)
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkspaceServiceInstance(ctx, req)
}
