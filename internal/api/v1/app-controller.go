package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.AppServiceServer = (*AppController)(nil)

// AppController is the app service controller handler.
type AppController struct {
	app service.Apper
}

// NewAppController returns a fresh admin service controller instance.
func NewAppController(app service.Apper) AppController {
	return AppController{app: app}
}

func (c AppController) GetApp(ctx context.Context, req *chorus.GetAppRequest) (*chorus.GetAppReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	app, err := c.app.GetApp(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetApp': %v", err.Error())
	}

	tgApp, err := converter.AppFromBusiness(app)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetAppReply{Result: &chorus.GetAppResult{App: tgApp}}, nil
}

func (c AppController) UpdateApp(ctx context.Context, req *chorus.App) (*chorus.UpdateAppReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	app, err := converter.AppToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	app.TenantID = tenantID

	updatedApp, err := c.app.UpdateApp(ctx, app)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateApp': %v", err.Error())
	}

	updatedAppProto, err := converter.AppFromBusiness(updatedApp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}
	return &chorus.UpdateAppReply{Result: &chorus.UpdateAppResult{App: updatedAppProto}}, nil
}

func (c AppController) DeleteApp(ctx context.Context, req *chorus.DeleteAppRequest) (*chorus.DeleteAppReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.app.DeleteApp(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteApp': %v", err.Error())
	}
	return &chorus.DeleteAppReply{Result: &chorus.DeleteAppResult{}}, nil
}

// ListApps extracts the retrieved apps from the service and inserts them into a reply object.
func (c AppController) ListApps(ctx context.Context, req *chorus.ListAppsRequest) (*chorus.ListAppsReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	res, paginationRes, err := c.app.ListApps(ctx, tenantID, &pagination)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListApps': %v", err.Error())
	}

	var apps []*chorus.App
	for _, r := range res {
		app, err := converter.AppFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		apps = append(apps, app)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListAppsReply{Result: &chorus.ListAppsResult{Apps: apps}, Pagination: paginationResult}, nil
}

// CreateApp extracts the app from the request and passes it to the app service.
func (c AppController) CreateApp(ctx context.Context, req *chorus.App) (*chorus.CreateAppReply, error) {
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

	app, err := converter.AppToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	app.TenantID = tenantID
	app.UserID = userID

	newApp, err := c.app.CreateApp(ctx, app)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateApp': %v", err.Error())
	}

	newAppProto, err := converter.AppFromBusiness(newApp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}
	return &chorus.CreateAppReply{Result: &chorus.CreateAppResult{App: newAppProto}}, nil
}

// BulkCreateApps extracts the apps from the request and passes them to the app service.
func (c AppController) BulkCreateApps(ctx context.Context, req *chorus.BulkCreateAppsRequest) (*chorus.BulkCreateAppsReply, error) {
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

	var apps []*model.App
	for _, r := range req.Apps {
		app, err := converter.AppToBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		app.TenantID = tenantID
		app.UserID = userID
		apps = append(apps, app)
	}

	newApps, err := c.app.BulkCreateApps(ctx, apps)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'BulkCreateApps': %v", err.Error())
	}

	var newAppsProto []*chorus.App
	for _, newApp := range newApps {
		newAppProto, err := converter.AppFromBusiness(newApp)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		newAppsProto = append(newAppsProto, newAppProto)
	}
	return &chorus.BulkCreateAppsReply{Apps: newAppsProto}, nil
}
