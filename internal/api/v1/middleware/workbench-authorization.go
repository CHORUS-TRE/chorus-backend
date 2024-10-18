package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

type workbenchControllerAuthorization struct {
	authorization
	next chorus.WorkbenchServiceServer
}

func WorkbenchAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
	return func(next chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
		return &workbenchControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c workbenchControllerAuthorization) ListWorkbenches(ctx context.Context, req *chorus.ListWorkbenchesRequest) (*chorus.ListWorkbenchesReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ListWorkbenches(ctx, req)
}

func (c workbenchControllerAuthorization) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	// nolint: staticcheck
	return c.next.CreateWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) UpdateWorkbench(ctx context.Context, req *chorus.UpdateWorkbenchRequest) (*chorus.UpdateWorkbenchReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteWorkbench(ctx, req)
}
