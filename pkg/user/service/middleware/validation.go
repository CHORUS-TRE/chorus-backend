package middleware

import (
	"context"
	"fmt"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next         service.Userer
	validate     *val.Validate
	requireEmail bool
}

func Validation(validate *val.Validate, requireEmail bool) func(service.Userer) service.Userer {
	return func(next service.Userer) service.Userer {
		return &validation{
			next:         next,
			validate:     validate,
			requireEmail: requireEmail,
		}
	}
}

func (v validation) CreateRole(ctx context.Context, role string) error {
	allRoles := authorization_model.GetAllRoles()
	allRolesStr := make([]string, len(allRoles))
	for i, role := range allRoles {
		allRolesStr[i] = string(role)
	}
	if !contains(allRolesStr, role) {
		return cerr.ErrValidation.WithMessage(fmt.Sprintf("Invalid role '%v', should be one of: %v", role, allRolesStr))
	}
	return v.next.CreateRole(ctx, role)
}

func (v validation) CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []model.UserRole) error {
	return v.next.CreateUserRoles(ctx, tenantID, userID, roles)
}

func (v validation) RemoveUserRoles(ctx context.Context, tenantID, userID uint64, userRoleIDs []uint64) error {
	return v.next.RemoveUserRoles(ctx, tenantID, userID, userRoleIDs)
}

func (v validation) GetRoles(ctx context.Context) ([]*model.Role, error) {
	return v.next.GetRoles(ctx)
}

func (v validation) ListUsers(ctx context.Context, req service.ListUsersReq) ([]*model.User, *common.PaginationResult, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, nil, cerr.WrapValidationError(err)
	}
	return v.next.ListUsers(ctx, req)
}

func (v validation) GetUser(ctx context.Context, req service.GetUserReq) (*model.User, error) {
	return v.next.GetUser(ctx, req)
}

func (v validation) SoftDeleteUser(ctx context.Context, req service.DeleteUserReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.SoftDeleteUser(ctx, req)
}

func (v validation) UpdateUser(ctx context.Context, req service.UpdateUserReq) (*model.User, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	return v.next.UpdateUser(ctx, req)
}

func (v validation) CreateUser(ctx context.Context, req service.CreateUserReq) (*model.User, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	if v.requireEmail && req.User.Email == "" {
		return nil, cerr.ErrValidation.WithMessage("Email is required")
	}
	return v.next.CreateUser(ctx, req)
}

func (v validation) UpdateUserPassword(ctx context.Context, req service.UpdateUserPasswordReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.UpdateUserPassword(ctx, req)
}

func (v validation) EnableUserTotp(ctx context.Context, req service.EnableTotpReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.EnableUserTotp(ctx, req)
}

func (v validation) ResetUserTotp(ctx context.Context, req service.ResetTotpReq) (string, []string, error) {
	if err := v.validate.Struct(req); err != nil {
		return "", nil, cerr.WrapValidationError(err)
	}
	return v.next.ResetUserTotp(ctx, req)
}

func (v validation) ResetUserPassword(ctx context.Context, req service.ResetUserPasswordReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.ResetUserPassword(ctx, req)
}

func (v validation) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	return v.next.GetTotpRecoveryCodes(ctx, tenantID, userID)
}

func (v validation) DeleteTotpRecoveryCode(ctx context.Context, req *service.DeleteTotpRecoveryCodeReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.DeleteTotpRecoveryCode(ctx, req)
}

func (v validation) UpsertGrants(ctx context.Context, grants []model.UserGrant) error {
	if err := v.validate.Var(grants, "dive"); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.UpsertGrants(ctx, grants)
}

func (v validation) DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error {
	return v.next.DeleteGrants(ctx, tenantID, userID, clientID)
}

func (v validation) GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error) {
	return v.next.GetUserGrants(ctx, tenantID, userID, clientID)
}

func contains(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
