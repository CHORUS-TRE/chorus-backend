package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/mailer"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/crypto"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/helper"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *notification_model.Notification, userIDs []uint64) error
}

type Userer interface {
	ListUsers(ctx context.Context, req ListUsersReq) ([]*model.User, *common.PaginationResult, error)
	GetUser(ctx context.Context, req GetUserReq) (*model.User, error)
	CreateUser(ctx context.Context, req CreateUserReq) (*model.User, error)
	CreateRole(ctx context.Context, role string) error
	CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []model.UserRole) error
	RemoveUserRoles(ctx context.Context, tenantID, userID uint64, userRoleIDs []uint64) error
	RemoveRolesByContext(ctx context.Context, contextDimension, contextValue string) ([]uint64, error)
	GetRoles(ctx context.Context) ([]*model.Role, error)
	// GetRolesWithContext(ctx context.Context, roleContext map[string]string) ([]*model.Role, error)
	SoftDeleteUser(ctx context.Context, req DeleteUserReq) error
	UpdateUser(ctx context.Context, req UpdateUserReq) (*model.User, error)
	UpdateUserPassword(ctx context.Context, req UpdateUserPasswordReq) error
	EnableUserTotp(ctx context.Context, req EnableTotpReq) error
	ResetUserTotp(ctx context.Context, req ResetTotpReq) (string, []string, error)
	ResetUserPassword(ctx context.Context, req ResetUserPasswordReq) error

	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error)
	DeleteTotpRecoveryCode(ctx context.Context, req *DeleteTotpRecoveryCodeReq) error

	UpsertGrants(ctx context.Context, grants []model.UserGrant) error
	DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error
	GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error)
}

type UserStore interface {
	ListUsers(ctx context.Context, tenantID uint64, pagination *common.Pagination, filter *UserFilter) ([]*model.User, *common.PaginationResult, error)
	GetUser(ctx context.Context, tenantID uint64, userID uint64) (*model.User, error)
	CreateUser(ctx context.Context, tenantID uint64, user *model.User) (*model.User, error)
	SoftDeleteUser(ctx context.Context, tenantID uint64, userID uint64) error
	UpdateUser(ctx context.Context, tenantID uint64, user *model.User) (*model.User, error)

	CreateRole(ctx context.Context, role string) error
	CreateUserRoles(ctx context.Context, userID uint64, roles []model.UserRole) error
	RemoveUserRoles(ctx context.Context, userID uint64, userRoleIDs []uint64) error
	RemoveRolesByContext(ctx context.Context, contextDimension, contextValue string) ([]uint64, error)
	GetRoles(ctx context.Context) ([]*model.Role, error)

	GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error)
	UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) (*model.User, error)
	DeleteTotpRecoveryCode(ctx context.Context, tenantID, codeID uint64) error

	UpsertGrants(ctx context.Context, grants []model.UserGrant) error
	DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error
	GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error)
}

type UserService struct {
	totpNumRecoveryCodes int
	daemonEncryptionKey  *crypto.Secret
	store                UserStore
	mailer               mailer.Mailer
	notificationStore    NotificationStore
}

func NewUserService(totpNumRecoveryCodes int, daemonEncryptionKey *crypto.Secret, store UserStore, mailer mailer.Mailer, notificationStore NotificationStore) *UserService {
	return &UserService{
		totpNumRecoveryCodes: totpNumRecoveryCodes,
		daemonEncryptionKey:  daemonEncryptionKey,
		store:                store,
		mailer:               mailer,
		notificationStore:    notificationStore,
	}
}

func (u *UserService) ListUsers(ctx context.Context, req ListUsersReq) ([]*model.User, *common.PaginationResult, error) {
	users, pagination, err := u.store.ListUsers(ctx, req.TenantID, req.Pagination, req.Filter)
	if err != nil {
		return nil, nil, cerr.ErrInternal.Wrap(err, "Unable to query users")
	}
	return users, pagination, nil
}

func (u *UserService) GetUser(ctx context.Context, req GetUserReq) (*model.User, error) {
	user, err := u.store.GetUser(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.ID))
	}

	user.Password = ""
	user.TotpSecret = nil

	return user, nil
}

func (u *UserService) UpdateUserPassword(ctx context.Context, req UpdateUserPasswordReq) error {
	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.UserID))
	}

	if !helper.CheckPassHash(user.Password, req.CurrentPassword) {
		logger.SecLog.Warn(ctx, fmt.Sprintf("wrong password of user: %v", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return cerr.ErrInvalidCredentials
	}

	if !helper.IsStrongPassword(req.NewPassword) {
		return cerr.ErrValidation.WithMessage(fmt.Sprintf("Password does not meet security requirements: complex password (not easily guessable) with at least 14 characters, among which 1 lowercase, 1 uppercase and 1 special character: %v", helper.SpecialChars))
	}

	hashed, err := helper.HashPass(req.NewPassword)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, "Unable to hash password")
	}

	user.Password = hashed
	user.PasswordChanged = true

	_, err = u.store.UpdateUser(ctx, req.TenantID, user)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to update user %v", req.UserID))
	}

	return nil
}

func (u *UserService) SoftDeleteUser(ctx context.Context, req DeleteUserReq) error {
	if err := u.store.SoftDeleteUser(ctx, req.TenantID, req.ID); err != nil {
		return cerr.WrapStoreError(err, "Unable to delete user")
	}

	return nil
}

func (u *UserService) UpdateUser(ctx context.Context, req UpdateUserReq) (*model.User, error) {

	user, err := u.store.GetUser(ctx, req.TenantID, req.User.ID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.User.ID))
	}

	// req.User.Roles = filterDuplicateRoles(req.User.Roles)

	user.FirstName = req.User.FirstName
	user.LastName = req.User.LastName
	user.Username = req.User.Username
	user.Email = req.User.Email
	user.Source = req.User.Source
	user.Status = req.User.Status

	if err = verifyRoles(req.User.Roles); err != nil {
		return nil, cerr.ErrValidation.Wrap(err, "Role verification failed")
	}

	user.Roles = req.User.Roles
	updatedUser, err := u.store.UpdateUser(ctx, req.TenantID, user)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to update user %v", req.User.ID))
	}

	return updatedUser, nil
}

func (u *UserService) CreateUser(ctx context.Context, req CreateUserReq) (*model.User, error) {

	// req.User.Roles = filterDuplicateRoles(req.User.Roles)

	if err := verifyRoles(req.User.Roles); err != nil {
		return nil, cerr.ErrValidation.Wrap(err, "Role verification failed")
	}

	if req.User.Password != "" {
		return u.createUserWithPassword(ctx, req)
	}

	password, err := helper.GeneratePassword(20)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to generate password")
	}

	hash, err := helper.HashPass(password)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to hash password")
	}

	user := reqToUserBusiness(req.User)
	user.Password = hash

	newUser, err := u.store.CreateUser(ctx, req.TenantID, user)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create user %v", user.Username))
	}

	go u.sendMailWithTempPassword("Please change your password", req.TenantID, user, password, mailer.TemporaryPasswordKey)

	return newUser, nil
}

func (u *UserService) createUserWithPassword(ctx context.Context, req CreateUserReq) (*model.User, error) {
	user := req.User
	if user.TotpEnabled {

		secret, err := crypto.CreateTotpSecret(user.Username, u.daemonEncryptionKey)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, "Unable to create totp secret")
		}
		user.TotpSecret = &secret

		recoveryCodes, err := crypto.CreateTotpRecoveryCodes(u.totpNumRecoveryCodes, u.daemonEncryptionKey)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, "Unable to create totp recovery codes")
		}
		user.TotpRecoveryCodes = recoveryCodes
		user.TotpEnabled = true
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to hash password")
	}
	user.Password = string(hash)
	user.PasswordChanged = true

	newUser, err := u.store.CreateUser(ctx, req.TenantID, reqToUserBusiness(req.User))
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to store user %v", user.Username))
	}
	return newUser, nil
}

func (u *UserService) EnableUserTotp(ctx context.Context, req EnableTotpReq) error {
	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.UserID))
	}

	isTotpValid, err := crypto.VerifyTotp(req.Totp, utils.ToString(user.TotpSecret), u.daemonEncryptionKey)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to verify totp", zap.Error(err))
		return cerr.ErrInvalidCredentials
	}
	if !isTotpValid {
		logger.SecLog.Warn(ctx, fmt.Sprintf("user %v has entered an invalid totp code", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return cerr.ErrInvalidCredentials
	}

	user.TotpEnabled = true
	if _, err := u.store.UpdateUser(ctx, req.TenantID, user); err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to update user %v", req.UserID))
	}

	return nil
}

func (u *UserService) ResetUserTotp(ctx context.Context, req ResetTotpReq) (string, []string, error) {

	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return "", nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.UserID))
	}

	if !helper.CheckPassHash(user.Password, req.Password) {
		logger.SecLog.Warn(ctx, fmt.Sprintf("wrong password of user: %v", req.UserID), zap.Uint64("tenant_id", req.TenantID))
		return "", nil, cerr.ErrInvalidCredentials
	}

	if u.totpNumRecoveryCodes == 0 {
		return "", nil, cerr.ErrInternal.WithMessage("Configuration value for totp num recovery codes is not set")
	}

	user.TotpEnabled = false

	totpSecret, err := crypto.CreateTotpSecret(user.Username, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, cerr.ErrInternal.Wrap(err, "Unable to create totp secret")
	}
	user.TotpSecret = &totpSecret

	decTotpSecret, err := crypto.DecryptTotpSecret(totpSecret, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, cerr.ErrInternal.Wrap(err, "Unable to decrypt totp secret")
	}

	recoveryCodes, err := crypto.CreateTotpRecoveryCodes(u.totpNumRecoveryCodes, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, cerr.ErrInternal.Wrap(err, "Unable to create totp recovery codes")
	}

	decRecoveryCodes, err := crypto.DecryptTotpRecoveryCodes(recoveryCodes, u.daemonEncryptionKey)
	if err != nil {
		return "", nil, cerr.ErrInternal.Wrap(err, "Unable to decrypt totp recovery codes")
	}

	if _, err = u.store.UpdateUserWithRecoveryCodes(ctx, req.TenantID, user, recoveryCodes); err != nil {
		return "", nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to update user %v", req.UserID))
	}

	return decTotpSecret, decRecoveryCodes, nil
}

func (u *UserService) ResetUserPassword(ctx context.Context, req ResetUserPasswordReq) error {

	password, err := helper.GeneratePassword(20)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, "Unable to generate password")
	}

	hash, err := helper.HashPass(password)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, "Unable to hash password")
	}

	user, err := u.store.GetUser(ctx, req.TenantID, req.UserID)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to get user %v", req.UserID))
	}

	user.Password = hash
	user.PasswordChanged = false
	user.TotpEnabled = false

	_, err = u.store.UpdateUser(ctx, req.TenantID, user)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to update user %v", req.UserID))
	}

	go u.sendMailWithTempPassword("Password reset, please change your password", req.TenantID, user, password, mailer.TemporaryPasswordKey)

	return nil
}

func (s *UserService) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	recoveryCodes, err := s.store.GetTotpRecoveryCodes(ctx, tenantID, userID)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to get totp recovery codes")
	}
	return recoveryCodes, nil
}

func (s *UserService) DeleteTotpRecoveryCode(ctx context.Context, req *DeleteTotpRecoveryCodeReq) error {
	err := s.store.DeleteTotpRecoveryCode(ctx, req.TenantID, req.CodeID)
	if err != nil {
		return cerr.WrapStoreError(err, "Unable to delete totp recovery code")
	}
	return nil
}

func (u *UserService) sendMailWithTempPassword(subjectMessage string, tenantID uint64, user *model.User, password string, templateKey mailer.TemplateKey) {
	ctx := context.Background()
	subject := u.mailer.GetSubject(ctx, tenantID, "temporaryPassword")
	if subject == "" {
		subject = subjectMessage
	}
	recipient := user.Email
	if recipient == "" {
		recipient = user.Username
	}
	err := u.mailer.Send(ctx, tenantID, []string{recipient}, subject, u.mailer.GetTemplate(ctx, tenantID, templateKey), mailer.TemporaryPassword{
		Email:    recipient,
		Password: password,
	})
	if err != nil {
		logger.BizLog.Error(ctx, fmt.Sprintf("unable to send temporary password to user: %v", user.Username), zap.Uint64("tenant_id", tenantID), zap.Error(err))
	} else {
		logger.BizLog.Info(ctx, fmt.Sprintf("temporary password sent to user: %v", user.Username), zap.Uint64("tenant_id", tenantID))
	}
}

func (u *UserService) CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []model.UserRole) error {
	err := verifyRoles(roles)
	if err != nil {
		return cerr.ErrValidation.Wrap(err, "Role verification failed")
	}

	err = u.store.CreateUserRoles(ctx, userID, roles)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create user roles for user %v", userID))
	}

	err = u.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  "You have been assigned new roles",
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: &notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create notification for user %v", userID))
	}

	return nil
}

func (u *UserService) RemoveUserRoles(ctx context.Context, tenantID, userID uint64, userRoleIDs []uint64) error {
	err := u.store.RemoveUserRoles(ctx, userID, userRoleIDs)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to remove user roles for user %v", userID))
	}

	err = u.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  "You have been assigned new roles",
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: &notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create notification for user %v", userID))
	}

	return nil
}

func (u *UserService) RemoveRolesByContext(ctx context.Context, contextDimension, contextValue string) ([]uint64, error) {
	userIDs, err := u.store.RemoveRolesByContext(ctx, contextDimension, contextValue)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to remove roles by context %s=%s", contextDimension, contextValue))
	}
	return userIDs, nil
}

func (u *UserService) CreateRole(ctx context.Context, role string) error {
	return u.store.CreateRole(ctx, role)
}

func (u *UserService) GetRoles(ctx context.Context) ([]*model.Role, error) {
	return u.store.GetRoles(ctx)
}

func (u *UserService) UpsertGrants(ctx context.Context, grants []model.UserGrant) error {
	err := u.store.UpsertGrants(ctx, grants)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, "Unable to upsert user grants")
	}
	return nil
}

func (u *UserService) DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error {
	err := u.store.DeleteGrants(ctx, tenantID, userID, clientID)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, "Unable to delete user grants")
	}
	return nil
}

func (u *UserService) GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error) {
	grants, err := u.store.GetUserGrants(ctx, tenantID, userID, clientID)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Unable to get user grants")
	}
	return grants, nil
}

// func (u *UserService) GetRolesWithContext(ctx context.Context, roleContext map[string]string) ([]*model.Role, error) {
// 	//TODO implement filter at DB level
// 	roles, err := u.store.GetRoles(ctx, roleContext)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to get roles: %w", err)
// 	}

// 	matchingRoles := []*model.Role{}
// 	for _, role := range roles {
// 		match := true
// 		for k, v := range roleContext {
// 			if roleVal, ok := role.Context[k]; !ok || roleVal != v {
// 				match = false
// 				break
// 			}
// 		}
// 	}

// 	return matchingRoles, nil
// }

func verifyRoles(roles []model.UserRole) error {
	// TODO: validate role & ctx

	// for _, role := range roles {
	// 	// if _, ok := model.ValidRoles[role]; !ok {
	// 	// 	err := &service.InvalidParametersErr{}
	// 	// 	return fmt.Errorf("invalid role: %s: %w", role, err)
	// 	// }

	// }

	return nil
}
