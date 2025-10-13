package v1

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	authentication_service "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

var _ chorus.UserServiceServer = (*UserController)(nil)

// UserController is the user service controller handler.
type UserController struct {
	user          service.Userer
	authenticator authentication_service.Authenticator
	cfg           config.Config
}

// NewUserController returns a fresh admin service controller instance.
func NewUserController(user service.Userer, cfg config.Config, authenticator authentication_service.Authenticator) UserController {
	return UserController{user: user, cfg: cfg, authenticator: authenticator}
}

func (c UserController) GetUserMe(ctx context.Context, req *chorus.GetUserMeRequest) (*chorus.GetUserMeReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	user, err := c.user.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       userID,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetUser': %v", err.Error())
	}

	tgUser, err := converter.UserFromBusiness(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}
	return &chorus.GetUserMeReply{Result: &chorus.GetUserMeResult{Me: tgUser}}, nil
}

func (c UserController) GetUser(ctx context.Context, req *chorus.GetUserRequest) (*chorus.GetUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	user, err := c.user.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       req.Id,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetUser': %v", err.Error())
	}

	tgUser, err := converter.UserFromBusiness(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetUserReply{Result: &chorus.GetUserResult{User: tgUser}}, nil
}

func (c UserController) UpdatePassword(ctx context.Context, req *chorus.UpdatePasswordRequest) (*chorus.UpdatePasswordReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	err = c.user.UpdateUserPassword(ctx, service.UpdateUserPasswordReq{
		TenantID:        tenantID,
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdatePassword': %v", err.Error())
	}

	return &chorus.UpdatePasswordReply{Result: &chorus.UpdateUserResult{}}, nil
}

func (c UserController) UpdateUser(ctx context.Context, req *chorus.User) (*chorus.UpdateUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	user, err := userToUpdateServiceRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	updatedUser, err := c.user.UpdateUser(ctx, service.UpdateUserReq{
		TenantID: tenantID,
		User:     user,
	})

	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateUser': %v", err.Error())
	}
	updatedUserProto, err := converter.UserFromBusiness(updatedUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.UpdateUserReply{Result: &chorus.UpdateUserResult{User: updatedUserProto}}, nil
}

func (c UserController) DeleteUser(ctx context.Context, req *chorus.DeleteUserRequest) (*chorus.DeleteUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.user.SoftDeleteUser(ctx, service.DeleteUserReq{
		TenantID: tenantID,
		ID:       req.Id,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteUser': %v", err.Error())
	}
	return &chorus.DeleteUserReply{Result: &chorus.DeleteUserResult{}}, nil
}

// ListUsers extracts the retrieved users from the service and inserts them into a reply object.
// Note that an admin role is required to call this procedure.
func (c UserController) ListUsers(ctx context.Context, req *chorus.ListUsersRequest) (*chorus.ListUsersReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)
	filter := UserFilterToBusiness(req.Filter)
	res, paginationRes, err := c.user.ListUsers(ctx, service.ListUsersReq{TenantID: tenantID, Pagination: &pagination, Filter: filter})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListUsers': %v", err.Error())
	}

	var users []*chorus.User
	for _, r := range res {
		user, err := converter.UserFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		users = append(users, user)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		result := converter.PaginationResultFromBusiness(paginationRes)
		paginationResult = result
	}

	return &chorus.ListUsersReply{Result: &chorus.ListUsersResult{Users: users}, Pagination: paginationResult}, nil
}

func UserFilterToBusiness(aFilter *chorus.UserFilter) *service.UserFilter {
	if aFilter == nil {
		return nil
	}
	return &service.UserFilter{
		IDsIn:        aFilter.IdsIn,
		WorkspaceIDs: aFilter.WorkspaceIDs,
		WorkbenchIDs: aFilter.WorkbenchIDs,
		Search:       aFilter.Search,
	}
}

// CreateUser extracts the user from the request and passes it to the user service.
func (c UserController) CreateUser(ctx context.Context, req *chorus.User) (*chorus.CreateUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// modes := c.authenticator.GetAuthenticationModes()
	// internalModePublicRegistration := false
	// for _, mode := range modes {
	// 	if mode.Type == "internal" && mode.Internal.PublicRegistrationEnabled {
	// 		internalModePublicRegistration = true
	// 		break
	// 	}
	// }
	// if !internalModePublicRegistration {
	// 	return nil, status.Error(codes.PermissionDenied, "public registration is disabled")
	// }

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		if c.cfg.Daemon.TenantID != 0 {
			tenantID = c.cfg.Daemon.TenantID
		} else {
			tenantID = 1
		}
	}

	user, err := userToServiceRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	user.Source = "internal"

	res, err := c.user.CreateUser(ctx, service.CreateUserReq{TenantID: tenantID, User: user})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateUser': %v", err.Error())
	}

	err = c.user.CreateUserRoles(ctx, res.ID, []model.UserRole{{
		Role: authorization_model.NewRole(
			authorization_model.RoleAuthenticated,
			authorization_model.WithUser(res.ID),
		),
	}})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateUserRoles': %v", err.Error())
	}

	tgUser, err := converter.UserFromBusiness(res)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateUserReply{Result: &chorus.CreateUserResult{User: tgUser}}, nil
}

func (c UserController) CreateUserRole(ctx context.Context, req *chorus.CreateUserRoleRequest) (*chorus.CreateUserRoleReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	role, err := authorization_model.ToRole(req.Role.Name, req.Role.Context)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role: %v", err.Error())
	}

	err = c.user.CreateUserRoles(ctx, req.UserId, []model.UserRole{{
		Role: role,
	}})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateUserRoles': %v", err.Error())
	}

	user, err := c.user.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       req.UserId,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetUser': %v", err.Error())
	}

	u, err := converter.UserFromBusiness(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateUserRoleReply{Result: &chorus.CreateUserRoleResult{User: u}}, nil
}

func (c UserController) DeleteUserRole(ctx context.Context, req *chorus.DeleteUserRoleRequest) (*chorus.DeleteUserRoleReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.user.RemoveUserRoles(ctx, req.UserId, []uint64{req.RoleId})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteUserRole': %v", err.Error())
	}

	user, err := c.user.GetUser(ctx, service.GetUserReq{
		TenantID: tenantID,
		ID:       req.UserId,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetUser': %v", err.Error())
	}

	u, err := converter.UserFromBusiness(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.DeleteUserRoleReply{Result: &chorus.DeleteUserRoleResult{User: u}}, nil
}

func (c UserController) EnableTotp(ctx context.Context, req *chorus.EnableTotpRequest) (*chorus.EnableTotpReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	if err = c.user.EnableUserTotp(ctx, service.EnableTotpReq{
		TenantID: tenantID,
		UserID:   userID,
		Totp:     req.Totp,
	}); err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'EnableTotp': %v", err.Error())
	}

	return &chorus.EnableTotpReply{Result: &chorus.EnableTotpResult{}}, nil
}

func (c UserController) ResetTotp(ctx context.Context, req *chorus.ResetTotpRequest) (*chorus.ResetTotpReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	totpSecret, totpRecoveryCodes, err := c.user.ResetUserTotp(ctx, service.ResetTotpReq{
		TenantID: tenantID,
		UserID:   userID,
		Password: req.Password,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ResetTotp' : %v", err.Error())
	}

	return &chorus.ResetTotpReply{Result: &chorus.ResetTotpResult{
		TotpSecret:        totpSecret,
		TotpRecoveryCodes: totpRecoveryCodes,
	}}, nil
}

func (c UserController) ResetPassword(ctx context.Context, req *chorus.ResetPasswordRequest) (*chorus.ResetPasswordReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	if err = c.user.ResetUserPassword(ctx, service.ResetUserPasswordReq{
		TenantID: tenantID,
		UserID:   req.Id,
	}); err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ResetUserPassword': %v", err.Error())
	}

	return &chorus.ResetPasswordReply{Result: &chorus.ResetPasswordResult{}}, nil
}

// userToServiceRequest converts a chorus.User to a model.User.
func userToServiceRequest(user *chorus.User) (*service.UserReq, error) {
	ca, err := converter.FromProtoTimestamp(user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := converter.FromProtoTimestamp(user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}
	userStatus, err := model.ToUserStatus(user.Status)
	if err != nil {
		return nil, err
	}
	// roles, err := model.ToUserRoles(user.Roles)
	// if err != nil {
	// 	return nil, err
	// }

	return &service.UserReq{
		ID:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Source:    user.Source,
		Password:  user.Password,
		Status:    userStatus,
		// Roles:       roles,
		TotpEnabled: user.TotpEnabled,
		CreatedAt:   ca,
		UpdatedAt:   ua,
	}, nil
}

func userToUpdateServiceRequest(user *chorus.User) (*service.UserUpdateReq, error) {
	userStatus, err := model.ToUserStatus(user.Status)
	if err != nil {
		return nil, err
	}

	// roles := make([]authorization.Role, len(user.Roles))
	// for i, r := range user.Roles {
	// 	role, err := authorization.ToRole(r.Name, r.Context)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	roles[i] = role
	// }
	return &service.UserUpdateReq{
		ID:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Source:    user.Source,
		Status:    userStatus,
		// Roles:     roles,
	}, nil
}
