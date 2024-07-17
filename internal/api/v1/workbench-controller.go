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

// WorkbenchController is the workbench service controller handler.
type WorkbenchController struct {
	workbench service.Workbencher
}

func (c WorkbenchController) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	workbench, err := c.workbench.GetWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetWorkbench': %v", err.Error())
	}

	tgWorkbench, err := converter.WorkbenchFromBusiness(workbench)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetWorkbenchReply{Result: &chorus.GetWorkbenchResult{Workbench: tgWorkbench}}, nil
}

func (c WorkbenchController) UpdateWorkbench(ctx context.Context, req *chorus.UpdateWorkbenchRequest) (*chorus.UpdateWorkbenchReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	workbench, err := converter.WorkbenchToBusiness(req.Workbench)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	workbench.TenantID = tenantID

	err = c.workbench.UpdateWorkbench(ctx, workbench)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateWorkbench': %v", err.Error())
	}
	return &chorus.UpdateWorkbenchReply{Result: &chorus.UpdateWorkbenchResult{}}, nil
}

func (c WorkbenchController) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.workbench.DeleteWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteWorkbench': %v", err.Error())
	}
	return &chorus.DeleteWorkbenchReply{Result: &chorus.DeleteWorkbenchResult{}}, nil
}

// NewWorkbenchController returns a fresh admin service controller instance.
func NewWorkbenchController(workbench service.Workbencher) WorkbenchController {
	return WorkbenchController{workbench: workbench}
}

// ListWorkbenchs extracts the retrieved workbenchs from the service and inserts them into a reply object.
func (c WorkbenchController) ListWorkbenchs(ctx context.Context, req *chorus.ListWorkbenchsRequest) (*chorus.ListWorkbenchsReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	res, err := c.workbench.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListWorkbenchs': %v", err.Error())
	}

	var workbenchs []*chorus.Workbench
	for _, r := range res {
		workbench, err := converter.WorkbenchFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		workbenchs = append(workbenchs, workbench)
	}
	return &chorus.ListWorkbenchsReply{Result: workbenchs}, nil
}

// CreateWorkbench extracts the workbench from the request and passes it to the workbench service.
func (c WorkbenchController) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
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

	workbench, err := converter.WorkbenchToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	workbench.TenantID = tenantID
	workbench.UserID = userID

	res, err := c.workbench.CreateWorkbench(ctx, workbench)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateWorkbench': %v", err.Error())
	}
	return &chorus.CreateWorkbenchReply{Result: &chorus.CreateWorkbenchResult{Id: res}}, nil
}
