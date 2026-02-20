package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.UserServiceServer = (*userControllerAudit)(nil)

type userControllerAudit struct {
	next        chorus.UserServiceServer
	auditWriter service.AuditWriter
}

func NewUserAuditMiddleware(auditWriter service.AuditWriter) func(chorus.UserServiceServer) chorus.UserServiceServer {
	return func(next chorus.UserServiceServer) chorus.UserServiceServer {
		return &userControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c userControllerAudit) GetUserMe(ctx context.Context, req *chorus.GetUserMeRequest) (*chorus.GetUserMeReply, error) {
	res, err := c.next.GetUserMe(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionUserRead,
			audit.WithDescription("Failed to get current user."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionUserRead,
	// 			audit.WithDescription("Retrieved current user."),
	// 			audit.WithDetail("user_id", res.Result.Me.Id),
	// 			audit.WithDetail("username", res.Result.Me.Username),
	// 		)
	// }

	return res, err
}

func (c userControllerAudit) GetUser(ctx context.Context, req *chorus.GetUserRequest) (*chorus.GetUserReply, error) {
	res, err := c.next.GetUser(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionUserRead,
			audit.WithDetail("user_id", req.Id),
			audit.WithDescription("Failed to get user."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionUserRead,
	// 			audit.WithDetail("user_id", req.Id),
	// 			audit.WithDescription(fmt.Sprintf("Retrieved user with ID %d.", req.Id)),
	// 			audit.WithDetail("username", res.Result.User.Username),
	// 		)
	// }

	return res, err
}

func (c userControllerAudit) ListUsers(ctx context.Context, req *chorus.ListUsersRequest) (*chorus.ListUsersReply, error) {
	res, err := c.next.ListUsers(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionUserList,
			audit.WithDetail("filter", req.Filter),
			audit.WithDescription("Failed to list users."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionUserList,
	// 			audit.WithDetail("filter", req.Filter),
	// 			audit.WithDescription("Listed users."),
	// 			audit.WithDetail("result_count", len(res.Result.Users)),
	// 		)
	// }

	return res, err
}

func (c userControllerAudit) CreateUser(ctx context.Context, req *chorus.User) (*chorus.CreateUserReply, error) {
	res, err := c.next.CreateUser(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("username", req.Username),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to create user."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created user with ID %d.", res.Result.User.Id)),
			audit.WithDetail("user_id", res.Result.User.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserCreate, opts...)

	return res, err
}

func (c userControllerAudit) UpdateUser(ctx context.Context, req *chorus.User) (*chorus.UpdateUserReply, error) {
	res, err := c.next.UpdateUser(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("user_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to update user."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated user with ID %d.", req.Id)),
			audit.WithDetail("username", req.Username),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserUpdate, opts...)

	return res, err
}

func (c userControllerAudit) UpdatePassword(ctx context.Context, req *chorus.UpdatePasswordRequest) (*chorus.UpdatePasswordReply, error) {
	res, err := c.next.UpdatePassword(ctx, req)

	var opts []audit.Option

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to change password."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("Successfully changed password."),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserPasswordChange, opts...)

	return res, err
}

func (c userControllerAudit) DeleteUser(ctx context.Context, req *chorus.DeleteUserRequest) (*chorus.DeleteUserReply, error) {
	res, err := c.next.DeleteUser(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("user_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to delete user."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted user with ID %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserDelete, opts...)

	return res, err
}

func (c userControllerAudit) CreateUserRole(ctx context.Context, req *chorus.CreateUserRoleRequest) (*chorus.CreateUserRoleReply, error) {
	res, err := c.next.CreateUserRole(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("user_id", req.UserId),
		audit.WithDetail("role_name", req.Role.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to assign role %s to user %d.", req.Role.Name, req.UserId)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Assigned role %s to user %d.", req.Role.Name, req.UserId)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserRoleAssign, opts...)

	return res, err
}

func (c userControllerAudit) DeleteUserRole(ctx context.Context, req *chorus.DeleteUserRoleRequest) (*chorus.DeleteUserRoleReply, error) {
	res, err := c.next.DeleteUserRole(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("user_id", req.UserId),
		audit.WithDetail("role_id", req.RoleId),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to revoke role %d from user %d.", req.RoleId, req.UserId)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Revoked role %d from user %d.", req.RoleId, req.UserId)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserRoleRevoke, opts...)

	return res, err
}

func (c userControllerAudit) EnableTotp(ctx context.Context, req *chorus.EnableTotpRequest) (*chorus.EnableTotpReply, error) {
	res, err := c.next.EnableTotp(ctx, req)

	var opts []audit.Option

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to enable TOTP."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("Successfully enabled TOTP."),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserTotpEnable, opts...)

	return res, err
}

func (c userControllerAudit) ResetTotp(ctx context.Context, req *chorus.ResetTotpRequest) (*chorus.ResetTotpReply, error) {
	res, err := c.next.ResetTotp(ctx, req)

	var opts []audit.Option

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to reset TOTP."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("Successfully reset TOTP."),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserTotpReset, opts...)

	return res, err
}

func (c userControllerAudit) ResetPassword(ctx context.Context, req *chorus.ResetPasswordRequest) (*chorus.ResetPasswordReply, error) {
	res, err := c.next.ResetPassword(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("user_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to reset password for user %d.", req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Successfully reset password for user %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionUserPasswordReset, opts...)

	return res, err
}
