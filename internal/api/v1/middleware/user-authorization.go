package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/golang/protobuf/ptypes/empty"
)

type userControllerAuthorization struct {
	authorization
	next chorus.UserServiceServer
}

func UserAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.UserServiceServer) chorus.UserServiceServer {
	return func(next chorus.UserServiceServer) chorus.UserServiceServer {
		return &userControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c userControllerAuthorization) GetUsers(ctx context.Context, req *chorus.GetUsersRequest) (*chorus.GetUsersReply, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetUsers(ctx, req)
}

func (c userControllerAuthorization) GetUser(ctx context.Context, req *chorus.GetUserRequest) (*chorus.GetUserReply, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetUser(ctx, req)
}

func (c userControllerAuthorization) CreateUser(ctx context.Context, req *chorus.User) (*chorus.CreateUserReply, error) {
	// err := c.IsAuthenticatedAndAuthorized(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	//nolint: staticcheck
	return c.next.CreateUser(ctx, req)
}

func (c userControllerAuthorization) GetUserMe(ctx context.Context, empty *empty.Empty) (*chorus.GetUserMeReply, error) {
	err := c.IsAuthenticatedAndAuthorizedWithRoles(ctx, []string{model.RoleAuthenticated.String()})
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetUserMe(ctx, empty)
}

func (c userControllerAuthorization) UpdatePassword(ctx context.Context, req *chorus.UpdatePasswordRequest) (*chorus.UpdatePasswordReply, error) {
	err := c.IsAuthenticatedAndAuthorizedWithRoles(ctx, []string{model.RoleAuthenticated.String()})
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdatePassword(ctx, req)
}

func (c userControllerAuthorization) UpdateUser(ctx context.Context, req *chorus.UpdateUserRequest) (*chorus.UpdateUserReply, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateUser(ctx, req)
}

func (c userControllerAuthorization) DeleteUser(ctx context.Context, req *chorus.DeleteUserRequest) (*chorus.DeleteUserReply, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteUser(ctx, req)
}

func (c userControllerAuthorization) EnableTotp(ctx context.Context, req *chorus.EnableTotpRequest) (*chorus.EnableTotpReply, error) {
	err := c.IsAuthenticatedAndAuthorizedWithRoles(ctx, []string{model.RoleAuthenticated.String()})
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.EnableTotp(ctx, req)
}

func (c userControllerAuthorization) ResetTotp(ctx context.Context, req *chorus.ResetTotpRequest) (*chorus.ResetTotpReply, error) {
	err := c.IsAuthenticatedAndAuthorizedWithRoles(ctx, []string{model.RoleAuthenticated.String()})
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ResetTotp(ctx, req)
}

func (c userControllerAuthorization) ResetPassword(ctx context.Context, req *chorus.ResetPasswordRequest) (*chorus.ResetPasswordReply, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ResetPassword(ctx, req)
}
