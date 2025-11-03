package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.DevstoreServiceServer = (*devstoreControllerAuthorization)(nil)

type devstoreControllerAuthorization struct {
	Authorization
	next chorus.DevstoreServiceServer
}

func DevstoreAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer, cfg config.Config, refresher Refresher) func(chorus.DevstoreServiceServer) chorus.DevstoreServiceServer {
	return func(next chorus.DevstoreServiceServer) chorus.DevstoreServiceServer {
		return &devstoreControllerAuthorization{
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

func (c devstoreControllerAuthorization) ListGlobalEntries(ctx context.Context, req *chorus.ListEntriesRequest) (*chorus.ListEntriesReply, error) {
	return c.next.ListGlobalEntries(ctx, req)
}

func (c devstoreControllerAuthorization) GetGlobalEntry(ctx context.Context, req *chorus.GetEntryRequest) (*chorus.GetEntryReply, error) {
	return c.next.GetGlobalEntry(ctx, req)
}

func (c devstoreControllerAuthorization) PutGlobalEntry(ctx context.Context, req *chorus.PutEntryRequest) (*chorus.PutEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionSetPlatformSettings)
	if err != nil {
		return nil, err
	}
	return c.next.PutGlobalEntry(ctx, req)
}

func (c devstoreControllerAuthorization) DeleteGlobalEntry(ctx context.Context, req *chorus.DeleteEntryRequest) (*chorus.DeleteEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionSetPlatformSettings)
	if err != nil {
		return nil, err
	}
	return c.next.DeleteGlobalEntry(ctx, req)
}

func (c devstoreControllerAuthorization) ListUserEntries(ctx context.Context, req *chorus.ListEntriesRequest) (*chorus.ListEntriesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetMyOwnUser, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return c.next.ListUserEntries(ctx, req)
}

func (c devstoreControllerAuthorization) GetUserEntry(ctx context.Context, req *chorus.GetEntryRequest) (*chorus.GetEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetMyOwnUser, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return c.next.GetUserEntry(ctx, req)
}

func (c devstoreControllerAuthorization) PutUserEntry(ctx context.Context, req *chorus.PutEntryRequest) (*chorus.PutEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetMyOwnUser, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return c.next.PutUserEntry(ctx, req)
}

func (c devstoreControllerAuthorization) DeleteUserEntry(ctx context.Context, req *chorus.DeleteEntryRequest) (*chorus.DeleteEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetMyOwnUser, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	return c.next.DeleteUserEntry(ctx, req)
}

func (c devstoreControllerAuthorization) ListWorkspaceEntries(ctx context.Context, req *chorus.ListWorkspaceEntriesRequest) (*chorus.ListEntriesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}
	return c.next.ListWorkspaceEntries(ctx, req)
}

func (c devstoreControllerAuthorization) GetWorkspaceEntry(ctx context.Context, req *chorus.GetWorkspaceEntryRequest) (*chorus.GetEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}
	return c.next.GetWorkspaceEntry(ctx, req)
}

func (c devstoreControllerAuthorization) PutWorkspaceEntry(ctx context.Context, req *chorus.PutWorkspaceEntryRequest) (*chorus.PutEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}
	return c.next.PutWorkspaceEntry(ctx, req)
}

func (c devstoreControllerAuthorization) DeleteWorkspaceEntry(ctx context.Context, req *chorus.DeleteWorkspaceEntryRequest) (*chorus.DeleteEntryReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}
	return c.next.DeleteWorkspaceEntry(ctx, req)
}
