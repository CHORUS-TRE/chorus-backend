package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Userer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Userer) service.Userer {
	return func(next service.Userer) service.Userer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) CreateRole(ctx context.Context, role string) error {
	if !contains([]string{"admin", "authenticated", "chorus"}, role) {
		return fmt.Errorf("invalid role '%v', should be on of: %v", role, []string{"admin", "authenticated", "chorus"})
	}
	return v.next.CreateRole(ctx, role)
}

func (v validation) GetRoles(ctx context.Context) ([]*model.Role, error) {
	return v.next.GetRoles(ctx)
}

func (v validation) GetUsers(ctx context.Context, req service.GetUsersReq) ([]*model.User, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, err
	}
	return v.next.GetUsers(ctx, req)
}

func (v validation) GetUser(ctx context.Context, req service.GetUserReq) (*model.User, error) {
	return v.next.GetUser(ctx, req)
}

func (v validation) SoftDeleteUser(ctx context.Context, req service.DeleteUserReq) error {
	if err := v.validate.Struct(req); err != nil {
		return v.next.SoftDeleteUser(ctx, req)
	}
	return v.next.SoftDeleteUser(ctx, req)
}

func (v validation) UpdateUser(ctx context.Context, req service.UpdateUserReq) error {
	if err := v.validate.Struct(req); err != nil {
		return v.next.UpdateUser(ctx, req)
	}
	return v.next.UpdateUser(ctx, req)
}

func (v validation) CreateUser(ctx context.Context, req service.CreateUserReq) (uint64, error) {
	if err := v.validate.Struct(req); err != nil {
		return 0, err
	}
	return v.next.CreateUser(ctx, req)
}

func (v validation) UpdateUserPassword(ctx context.Context, req service.UpdateUserPasswordReq) error {
	if err := v.validate.Struct(req); err != nil {
		return err
	}
	return v.next.UpdateUserPassword(ctx, req)
}

func (v validation) EnableUserTotp(ctx context.Context, req service.EnableTotpReq) error {
	if err := v.validate.Struct(req); err != nil {
		return err
	}
	return v.next.EnableUserTotp(ctx, req)
}

func (v validation) ResetUserTotp(ctx context.Context, req service.ResetTotpReq) (string, []string, error) {
	if err := v.validate.Struct(req); err != nil {
		return "", nil, err
	}
	return v.next.ResetUserTotp(ctx, req)
}

func (v validation) ResetUserPassword(ctx context.Context, req service.ResetUserPasswordReq) error {
	if err := v.validate.Struct(req); err != nil {
		return err
	}
	return v.next.ResetUserPassword(ctx, req)
}

func (v validation) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	return v.next.GetTotpRecoveryCodes(ctx, tenantID, userID)
}

func (v validation) DeleteTotpRecoveryCode(ctx context.Context, req *service.DeleteTotpRecoveryCodeReq) error {
	if err := v.validate.Struct(req); err != nil {
		return err
	}
	return v.next.DeleteTotpRecoveryCode(ctx, req)
}

func contains(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
