package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.AppInstanceServiceServer = (*AppInstanceController)(nil)

// NewAppInstanceController returns a fresh admin service controller instance.
func NewAppInstanceController(workbencher service.Workbencher) AppInstanceController {
	return AppInstanceController{workbencher: workbencher}
}

// AppInstanceController is the appInstance service controller handler.
type AppInstanceController struct {
	workbencher service.Workbencher
}

func (c AppInstanceController) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := c.workbencher.GetAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetAppInstance': %v", err.Error())
	}

	tgAppInstance, err := converter.AppInstanceFromBusiness(appInstance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetAppInstanceReply{Result: &chorus.GetAppInstanceResult{AppInstance: tgAppInstance}}, nil
}

func (c AppInstanceController) UpdateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.UpdateAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := converter.AppInstanceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	appInstance.TenantID = tenantID

	updatedAppInstance, err := c.workbencher.UpdateAppInstance(ctx, appInstance)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateAppInstance': %v", err.Error())
	}

	updatedAppInstanceProto, err := converter.AppInstanceFromBusiness(updatedAppInstance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.UpdateAppInstanceReply{Result: &chorus.UpdateAppInstanceResult{AppInstance: updatedAppInstanceProto}}, nil
}

func (c AppInstanceController) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.workbencher.DeleteAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteAppInstance': %v", err.Error())
	}
	return &chorus.DeleteAppInstanceReply{Result: &chorus.DeleteAppInstanceResult{}}, nil
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

	filter := service.AppInstanceFilter{}
	if req.Filter != nil {
		filter.WorkbenchIDsIn = &req.Filter.WorkbenchIdsIn
	}

	res, paginationRes, err := c.workbencher.ListAppInstances(ctx, tenantID, &pagination, filter)
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

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListAppInstancesReply{Result: &chorus.ListAppInstancesResult{AppInstances: appInstances}, Pagination: paginationResult}, nil
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

	res, err := c.workbencher.CreateAppInstance(ctx, appInstance)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateAppInstance': %v", err.Error())
	}

	appInstanceProto, err := converter.AppInstanceFromBusiness(res)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateAppInstanceReply{Result: &chorus.CreateAppInstanceResult{AppInstance: appInstanceProto}}, nil
}
