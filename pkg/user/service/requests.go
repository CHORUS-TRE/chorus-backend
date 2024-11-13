package service

import (
	"time"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

type UserReq struct {
	ID       uint64
	TenantID uint64

	FirstName       string `validate:"required,generalstring"`
	LastName        string `validate:"required,generalstring"`
	Username        string `validate:"required,generalstring"`
	Source          string `validate:"required,generalstring"`
	Password        string
	PasswordChanged bool
	Status          model.UserStatus `validate:"required"`

	TotpEnabled       bool
	TotpSecret        *string
	TotpRecoveryCodes []string

	Roles []model.UserRole `validate:"min=1"` // 1 element required min

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserUpdateReq struct {
	ID uint64

	FirstName string           `validate:"required,generalstring"`
	LastName  string           `validate:"required,generalstring"`
	Username  string           `validate:"required,generalstring"`
	Source    string           `validate:"required,generalstring"`
	Status    model.UserStatus `validate:"required"`

	Roles []model.UserRole `validate:"min=1"`
}

type GetUsersReq struct {
	TenantID uint64
}

type GetUserReq struct {
	TenantID uint64
	ID       uint64
}

type UpdateUserPasswordReq struct {
	TenantID        uint64 `validate:"required"`
	UserID          uint64 `validate:"required"`
	CurrentPassword string `validate:"required,max=254,generalstring"`
	NewPassword     string `validate:"required,max=254,generalstring"`
}

type DeleteUserReq struct {
	TenantID uint64
	ID       uint64
}

type UpdateUserReq struct {
	TenantID uint64
	User     *UserUpdateReq
}

type CreateUserReq struct {
	TenantID uint64
	User     *UserReq
}

type EnableTotpReq struct {
	TenantID uint64 `validate:"required,min=1"`
	UserID   uint64 `validate:"required,min=1"`
	Totp     string `validate:"required,min=6,max=10,safestring"`
}

type ResetTotpReq struct {
	TenantID uint64 `validate:"required,min=1"`
	UserID   uint64 `validate:"required,min=1"`
	Password string `validate:"required,max=254"`
}

type ResetUserPasswordReq struct {
	TenantID uint64 `validate:"required"`
	UserID   uint64 `validate:"required"`
}

type DeleteTotpRecoveryCodeReq struct {
	TenantID uint64 `validate:"required,min=1"`
	CodeID   uint64 `validate:"required,min=1"`
}

func reqToUserBusiness(req *UserReq) *model.User {
	return &model.User{
		ID:              req.ID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Username:        req.Username,
		Source:          req.Source,
		Password:        req.Password,
		PasswordChanged: req.PasswordChanged,
		Status:          req.Status,
		Roles:           req.Roles,
		TotpEnabled:     req.TotpEnabled,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
	}
}
