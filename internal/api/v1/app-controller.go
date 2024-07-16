package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AppController is the app service controller handler.
type AppController struct {
	app service.Apper
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

func (c AppController) UpdateApp(ctx context.Context, req *chorus.UpdateAppRequest) (*chorus.UpdateAppReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	app, err := converter.AppToBusiness(req.App)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	app.TenantID = tenantID

	err = c.app.UpdateApp(ctx, app)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateApp': %v", err.Error())
	}
	return &chorus.UpdateAppReply{Result: &chorus.UpdateAppResult{}}, nil
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

// NewAppController returns a fresh admin service controller instance.
func NewAppController(app service.Apper) AppController {
	return AppController{app: app}
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

	res, err := c.app.ListApps(ctx, tenantID, pagination)
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
	return &chorus.ListAppsReply{Result: apps}, nil
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

	res, err := c.app.CreateApp(ctx, app)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateApp': %v", err.Error())
	}
	return &chorus.CreateAppReply{Result: &chorus.CreateAppResult{Id: res}}, nil
}
