package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.UserServiceServer = (*userControllerAuthorization)(nil)

type userControllerAuthorization struct {
	Authorization
	next chorus.UserServiceServer
}

func UserAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.UserServiceServer) chorus.UserServiceServer {
	return func(next chorus.UserServiceServer) chorus.UserServiceServer {
		return &userControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c userControllerAuthorization) ListUsers(ctx context.Context, req *chorus.ListUsersRequest) (*chorus.ListUsersReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListUsers)
	if err != nil {
		return nil, err
	}

	return c.next.ListUsers(ctx, req)
}

func (c userControllerAuthorization) GetUser(ctx context.Context, req *chorus.GetUserRequest) (*chorus.GetUserReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetUser, authorization.WithUser(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetUser(ctx, req)
}

func (c userControllerAuthorization) CreateUser(ctx context.Context, req *chorus.User) (*chorus.CreateUserReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateUser)
	if err != nil {
		return nil, err
	}

	return c.next.CreateUser(ctx, req)
}

func (c userControllerAuthorization) GetUserMe(ctx context.Context, req *chorus.GetUserMeRequest) (*chorus.GetUserMeReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetMyOwnUser, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	return c.next.GetUserMe(ctx, req)
}

func (c userControllerAuthorization) UpdatePassword(ctx context.Context, req *chorus.UpdatePasswordRequest) (*chorus.UpdatePasswordReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdatePassword, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	return c.next.UpdatePassword(ctx, req)
}

func (c userControllerAuthorization) UpdateUser(ctx context.Context, req *chorus.User) (*chorus.UpdateUserReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateUser, authorization.WithUser(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateUser(ctx, req)
}

func (c userControllerAuthorization) DeleteUser(ctx context.Context, req *chorus.DeleteUserRequest) (*chorus.DeleteUserReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteUser, authorization.WithUser(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteUser(ctx, req)
}

func (c userControllerAuthorization) EnableTotp(ctx context.Context, req *chorus.EnableTotpRequest) (*chorus.EnableTotpReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionEnableTotp, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	return c.next.EnableTotp(ctx, req)
}

func (c userControllerAuthorization) ResetTotp(ctx context.Context, req *chorus.ResetTotpRequest) (*chorus.ResetTotpReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionResetTotp, authorization.WithUserFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	return c.next.ResetTotp(ctx, req)
}

func (c userControllerAuthorization) ResetPassword(ctx context.Context, req *chorus.ResetPasswordRequest) (*chorus.ResetPasswordReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionResetPassword, authorization.WithUser(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ResetPassword(ctx, req)
}
