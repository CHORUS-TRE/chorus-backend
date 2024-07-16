package v1

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserController is the user service controller handler.
type UserController struct {
	user service.Userer
}

func (c UserController) GetUserMe(ctx context.Context, empty *empty.Empty) (*chorus.GetUserMeReply, error) {
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

func (c UserController) UpdateUser(ctx context.Context, req *chorus.UpdateUserRequest) (*chorus.UpdateUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	user, err := userToUpdateServiceRequest(req.User)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	err = c.user.UpdateUser(ctx, service.UpdateUserReq{
		TenantID: tenantID,
		User:     user,
	})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateUser': %v", err.Error())
	}
	return &chorus.UpdateUserReply{Result: &chorus.UpdateUserResult{}}, nil
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

// NewUserController returns a fresh admin service controller instance.
func NewUserController(user service.Userer) UserController {
	return UserController{user: user}
}

// GetUsers extracts the retrieved users from the service and inserts them into a reply object.
// Note that an admin role is required to call this procedure.
func (c UserController) GetUsers(ctx context.Context, req *chorus.GetUsersRequest) (*chorus.GetUsersReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	res, err := c.user.GetUsers(ctx, service.GetUsersReq{TenantID: tenantID})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetUsers': %v", err.Error())
	}

	var users []*chorus.User
	for _, r := range res {
		user, err := converter.UserFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		users = append(users, user)
	}
	return &chorus.GetUsersReply{Result: users}, nil
}

// CreateUser extracts the user from the request and passes it to the user service.
func (c UserController) CreateUser(ctx context.Context, req *chorus.User) (*chorus.CreateUserReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		tenantID = 1
	}

	user, err := userToServiceRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	res, err := c.user.CreateUser(ctx, service.CreateUserReq{TenantID: tenantID, User: user})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateUser': %v", err.Error())
	}
	return &chorus.CreateUserReply{Result: &chorus.CreateUserResult{Id: res}}, nil
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
		return nil, errors.Wrap(err, "unable to convert createdAt timestamp")
	}
	ua, err := converter.FromProtoTimestamp(user.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert updatedAt timestamp")
	}
	userStatus, err := model.ToUserStatus(user.Status)
	if err != nil {
		return nil, err
	}
	roles, err := model.ToUserRoles(user.Roles)
	if err != nil {
		return nil, err
	}

	return &service.UserReq{
		ID:          user.Id,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Username:    user.Username,
		Password:    user.Password,
		Status:      userStatus,
		Roles:       roles,
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
	roles, err := model.ToUserRoles(user.Roles)
	if err != nil {
		return nil, err
	}

	return &service.UserUpdateReq{
		ID:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Status:    userStatus,
		Roles:     roles,
	}, nil
}
