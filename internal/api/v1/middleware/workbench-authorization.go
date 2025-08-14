package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.WorkbenchServiceServer = (*workbenchControllerAuthorization)(nil)

type workbenchControllerAuthorization struct {
	Authorization
	next chorus.WorkbenchServiceServer
}

func WorkbenchAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
	return func(next chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
		return &workbenchControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c workbenchControllerAuthorization) ListWorkbenchs(ctx context.Context, req *chorus.ListWorkbenchsRequest) (*chorus.ListWorkbenchsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListWorkbenchs)
	if err != nil {
		return nil, err
	}

	return c.next.ListWorkbenchs(ctx, req)
}

func (c workbenchControllerAuthorization) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateWorkbench)
	if err != nil {
		return nil, err
	}

	return c.next.CreateWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkbench(ctx, req)
}
