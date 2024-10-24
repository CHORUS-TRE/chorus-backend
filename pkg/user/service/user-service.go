package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/mailer"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/helper"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

type Userer interface {
	GetUsers(ctx context.Context, req GetUsersReq) ([]*model.User, error)
	GetUser(ctx context.Context, req GetUserReq) (*model.User, error)
	CreateUser(ctx context.Context, req CreateUserReq) (uint64, error)
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*model.Role, error)
	SoftDeleteUser(ctx context.Context, req DeleteUserReq) error
	UpdateUser(ctx context.Context, req UpdateUserReq) error
	UpdateUserPassword(ctx context.Context, req UpdateUserPasswordReq) error
	EnableUserTotp(ctx context.Context, req EnableTotpReq) error
	ResetUserTotp(ctx context.Context, req ResetTotpReq) (string, []string, error)
	ResetUserPassword(ctx context.Context, req ResetUserPasswordReq) error

	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error)
	DeleteTotpRecoveryCode(ctx context.Context, req *DeleteTotpRecoveryCodeReq) error
}

type UserStore interface {
	GetUsers(ctx context.Context, tenantID uint64) ([]*model.User, error)
	GetUser(ctx context.Context, tenantID uint64, userID uint64) (*model.User, error)
	CreateUser(ctx context.Context, tenantID uint64, user *model.User) (uint64, error)
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*model.Role, error)
	SoftDeleteUser(ctx context.Context, tenantID uint64, userID uint64) error
	UpdateUser(ctx context.Context, tenantID uint64, user *model.User) error
	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error)
	UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) error
	DeleteTotpRecoveryCode(ctx context.Context, tenantID, codeID uint64) error
}

type ErrUnauthorized struct{}

func (e *ErrUnauthorized) Error() string {
	return "invalid credentials"
}

type Err2FARequired struct{}

func (e *Err2FARequired) Error() string {
	return "2FA_REQUIRED"
}

type ErrWeakPassword struct{}

func (e *ErrWeakPassword) Error() string {
	return fmt.Sprintf("password does not meet security requirements: complex password (not easily guessable) with at least 14 characters, among which 1 lowercase, 1 uppercase and 1 special character: %v", helper.SpecialChars)
}

type UserService struct {
	totpNumRecoveryCodes int
	daemonEncryptionKey  *crypto.Secret
	store                UserStore
	mailer               mailer.Mailer
}

func NewUserService(totpNumRecoveryCodes int, daemonEncryptionKey *crypto.Secret, store UserStore, mailer mailer.Mailer) *UserService {
	return &UserService{
		totpNumRecoveryCodes: totpNumRecoveryCodes,
		daemonEncryptionKey:  daemonEncryptionKey,
		store:                store,
		mailer:               mailer,
	}
}

func (u *UserService) GetUsers(ctx context.Context, req GetUsersReq) ([]*model.User, error) {
	users, err := u.store.GetUsers(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %w", err)
	}
	return users, nil
}

func (u *UserService) GetUser(ctx context.Context, req GetUserReq) (*model.User, error) {
	user, err := u.store.GetUser(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get user %v: %w", req.ID, err)
	}

	user.Password = ""
	user.TotpSecret = nil

	return user, nil
}

func (u *UserService) UpdateUserPassword(ctx context.Context, req UpdateUserPasswordReq) error {
	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", req.UserID, err)
	}

	if !helper.CheckPassHash(user.Password, req.CurrentPassword) {
		logger.SecLog.Warn(ctx, fmt.Sprintf("wrong password of user: %v", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return &ErrUnauthorized{}
	}

	if !helper.IsStrongPassword(req.NewPassword) {
		return &ErrWeakPassword{}
	}

	hashed, err := helper.HashPass(req.NewPassword)
	if err != nil {
		return fmt.Errorf("unable to hash password: %w", err)
	}

	user.Password = hashed
	user.PasswordChanged = true

	err = u.store.UpdateUser(ctx, req.TenantID, user)
	if err != nil {
		return fmt.Errorf("unable to update user %v: %w", req.UserID, err)
	}
	return nil

}

func (u *UserService) SoftDeleteUser(ctx context.Context, req DeleteUserReq) error {
	if err := u.store.SoftDeleteUser(ctx, req.TenantID, req.ID); err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return nil
}

func (u *UserService) UpdateUser(ctx context.Context, req UpdateUserReq) error {

	user, err := u.store.GetUser(ctx, req.TenantID, req.User.ID)
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", req.User.ID, err)
	}

	req.User.Roles = filterDuplicateRoles(req.User.Roles)

	user.FirstName = req.User.FirstName
	user.LastName = req.User.LastName
	user.Username = req.User.Username
	user.Source = req.User.Source
	user.Status = req.User.Status

	if err = verifyRoles(req.User.Roles); err != nil {
		return fmt.Errorf("role verification failed: %w", err)
	}
	user.Roles = req.User.Roles

	if err := u.store.UpdateUser(ctx, req.TenantID, user); err != nil {
		return fmt.Errorf("unable to update user %v: %w", req.User.ID, err)
	}

	return nil
}

func (u *UserService) CreateUser(ctx context.Context, req CreateUserReq) (uint64, error) {

	req.User.Roles = filterDuplicateRoles(req.User.Roles)

	if err := verifyRoles(req.User.Roles); err != nil {
		return 0, err
	}

	if req.User.Password != "" {
		return u.createUserWithPassword(ctx, req)
	}

	password, err := helper.GeneratePassword(20)
	if err != nil {
		return 0, fmt.Errorf("unable to generate password: %w", err)
	}

	hash, err := helper.HashPass(password)
	if err != nil {
		return 0, fmt.Errorf("unable to hash password: %w", err)
	}

	user := reqToUserBusiness(req.User)
	user.Password = hash

	id, err := u.store.CreateUser(ctx, req.TenantID, user)
	if err != nil {
		return 0, fmt.Errorf("unable to create user %v: %w", user.Username, err)
	}

	go u.sendMailWithTempPassword("Please change your password", req.TenantID, user, password, mailer.TemporaryPasswordKey)

	return id, nil
}

func (u *UserService) createUserWithPassword(ctx context.Context, req CreateUserReq) (uint64, error) {
	user := req.User
	if user.TotpEnabled {

		secret, err := crypto.CreateTotpSecret(user.Username, u.daemonEncryptionKey)
		if err != nil {
			return 0, fmt.Errorf("unable to create totp secret: %w", err)
		}
		user.TotpSecret = &secret

		recoveryCodes, err := crypto.CreateTotpRecoveryCodes(u.totpNumRecoveryCodes, u.daemonEncryptionKey)
		if err != nil {
			return 0, fmt.Errorf("unable to create totp recovery codes: %w", err)
		}
		user.TotpRecoveryCodes = recoveryCodes
		user.TotpEnabled = true
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("unable to hash password: %w", err)
	}
	user.Password = string(hash)
	user.PasswordChanged = true

	id, err := u.store.CreateUser(ctx, req.TenantID, reqToUserBusiness(req.User))
	if err != nil {
		return 0, fmt.Errorf("unable to store user: %w", err)
	}
	return id, nil
}

func (u *UserService) EnableUserTotp(ctx context.Context, req EnableTotpReq) error {
	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("unable to get user: %v: %w", req.UserID, err)
	}

	isTotpValid, err := crypto.VerifyTotp(req.Totp, utils.ToString(user.TotpSecret), u.daemonEncryptionKey)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to verify totp", zap.Error(err))
		return &ErrUnauthorized{}
	}
	if !isTotpValid {
		logger.SecLog.Warn(ctx, fmt.Sprintf("user %v has entered an invalid totp code", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return &ErrUnauthorized{}
	}

	user.TotpEnabled = true
	if err := u.store.UpdateUser(ctx, req.TenantID, user); err != nil {
		return fmt.Errorf("unable to update user %v: %w", req.UserID, err)
	}

	return nil
}

func (u *UserService) ResetUserTotp(ctx context.Context, req ResetTotpReq) (string, []string, error) {

	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return "", nil, fmt.Errorf("unable to get user %v: %w", req.UserID, err)
	}

	if !helper.CheckPassHash(user.Password, req.Password) {
		logger.SecLog.Warn(ctx, fmt.Sprintf("wrong password of user: %v", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return "", nil, &ErrUnauthorized{}
	}

	if u.totpNumRecoveryCodes == 0 {
		return "", nil, errors.New("configuration value for totp num recovery codes is not set")
	}

	user.TotpEnabled = false

	totpSecret, err := crypto.CreateTotpSecret(user.Username, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create totp secret: %w", err)
	}
	user.TotpSecret = &totpSecret

	decTotpSecret, err := crypto.DecryptTotpSecret(totpSecret, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, fmt.Errorf("unable to decrypt totp secret: %w", err)
	}

	recoveryCodes, err := crypto.CreateTotpRecoveryCodes(u.totpNumRecoveryCodes, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create totp recovery codes: %w", err)
	}

	decRecoveryCodes, err := crypto.DecryptTotpRecoveryCodes(recoveryCodes, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, fmt.Errorf("unable to decrypt totp recovery codes: %w", err)
	}

	if err = u.store.UpdateUserWithRecoveryCodes(ctx, req.TenantID, user, recoveryCodes); err != nil {
		return "", nil, fmt.Errorf("unable to update user %v: %w", req.UserID, err)
	}

	return decTotpSecret, decRecoveryCodes, nil
}

func (u *UserService) ResetUserPassword(ctx context.Context, req ResetUserPasswordReq) error {

	password, err := helper.GeneratePassword(20)
	if err != nil {
		return fmt.Errorf("unable to generate password: %w", err)
	}

	hash, err := helper.HashPass(password)
	if err != nil {
		return fmt.Errorf("unable to hash password: %w", err)
	}

	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", req.UserID, err)
	}

	user.Password = hash
	user.PasswordChanged = false
	user.TotpEnabled = false

	err = u.store.UpdateUser(ctx, req.TenantID, user)
	if err != nil {
		return fmt.Errorf("unable to update user %v: %w", req.UserID, err)
	}

	go u.sendMailWithTempPassword("Password reset, please change your password", req.TenantID, user, password, mailer.TemporaryPasswordKey)

	return nil
}

func (s *UserService) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	recoveryCodes, err := s.store.GetTotpRecoveryCodes(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get totp recovery codes for user: %w", err)
	}
	return recoveryCodes, nil
}

func (s *UserService) DeleteTotpRecoveryCode(ctx context.Context, req *DeleteTotpRecoveryCodeReq) error {
	err := s.store.DeleteTotpRecoveryCode(ctx, req.TenantID, req.CodeID)
	if err != nil {
		return fmt.Errorf("unable to delete totp recovery code: %w", err)
	}
	return nil
}

func (u *UserService) sendMailWithTempPassword(subjectMessage string, tenantID uint64, user *model.User, password string, templateKey mailer.TemplateKey) {
	ctx := context.Background()
	subject := u.mailer.GetSubject(ctx, tenantID, "temporaryPassword")
	if subject == "" {
		subject = subjectMessage
	}
	err := u.mailer.Send(ctx, tenantID, []string{user.Username}, subject, u.mailer.GetTemplate(ctx, tenantID, templateKey), mailer.TemporaryPassword{
		Email:    user.Username,
		Password: password,
	})
	if err != nil {
		logger.BizLog.Error(ctx, fmt.Sprintf("unable to send temporary password to user: %v", user.Username), zap.Uint64("tenant_id", tenantID), zap.Error(err))
	} else {
		logger.BizLog.Info(ctx, fmt.Sprintf("temporary password sent to user: %v", user.Username), zap.Uint64("tenant_id", tenantID))
	}
}

func (u *UserService) CreateRole(ctx context.Context, role string) error {
	return u.store.CreateRole(ctx, role)
}

func (u *UserService) GetRoles(ctx context.Context) ([]*model.Role, error) {
	return u.store.GetRoles(ctx)
}

func verifyRoles(roles []model.UserRole) error {
	for _, role := range roles {
		if _, ok := model.ValidRoles[role]; !ok {
			err := &service.InvalidParametersErr{}
			return fmt.Errorf("invalid role: %s: %w", role, err)
		}
	}

	return nil
}

func filterDuplicateRoles(roles []model.UserRole) []model.UserRole {
	roleSet := make(map[model.UserRole]struct{})
	var filteredRoles []model.UserRole

	for _, role := range roles {
		if _, ok := roleSet[role]; !ok {
			roleSet[role] = struct{}{}
			filteredRoles = append(filteredRoles, role)
		}
	}

	return filteredRoles
}
