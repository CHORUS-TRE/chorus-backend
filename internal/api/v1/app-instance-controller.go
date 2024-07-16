package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AppInstanceController is the appInstance service controller handler.
type AppInstanceController struct {
	appInstance service.AppInstanceer
}

func (c AppInstanceController) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := c.appInstance.GetAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetAppInstance': %v", err.Error())
	}

	tgAppInstance, err := converter.AppInstanceFromBusiness(appInstance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetAppInstanceReply{Result: &chorus.GetAppInstanceResult{AppInstance: tgAppInstance}}, nil
}

func (c AppInstanceController) UpdateAppInstance(ctx context.Context, req *chorus.UpdateAppInstanceRequest) (*chorus.UpdateAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := converter.AppInstanceToBusiness(req.AppInstance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	appInstance.TenantID = tenantID

	err = c.appInstance.UpdateAppInstance(ctx, appInstance)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateAppInstance': %v", err.Error())
	}
	return &chorus.UpdateAppInstanceReply{Result: &chorus.UpdateAppInstanceResult{}}, nil
}

func (c AppInstanceController) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.appInstance.DeleteAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteAppInstance': %v", err.Error())
	}
	return &chorus.DeleteAppInstanceReply{Result: &chorus.DeleteAppInstanceResult{}}, nil
}

// NewAppInstanceController returns a fresh admin service controller instance.
func NewAppInstanceController(appInstance service.AppInstanceer) AppInstanceController {
	return AppInstanceController{appInstance: appInstance}
}

// ListAppInstances extracts the retrieved appInstances from the service and inserts them into a reply object.
func (c AppInstanceController) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	res, err := c.appInstance.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListAppInstances': %v", err.Error())
	}

	var appInstances []*chorus.AppInstance
	for _, r := range res {
		appInstance, err := converter.AppInstanceFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		appInstances = append(appInstances, appInstance)
	}
	return &chorus.ListAppInstancesReply{Result: appInstances}, nil
}

// CreateAppInstance extracts the appInstance from the request and passes it to the appInstance service.
func (c AppInstanceController) CreateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.CreateAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		tenantID = 1
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		tenantID = 1
	}

	appInstance, err := converter.AppInstanceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	appInstance.TenantID = tenantID
	appInstance.UserID = userID

	res, err := c.appInstance.CreateAppInstance(ctx, appInstance)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateAppInstance': %v", err.Error())
	}
	return &chorus.CreateAppInstanceReply{Result: &chorus.CreateAppInstanceResult{Id: res}}, nil
}
