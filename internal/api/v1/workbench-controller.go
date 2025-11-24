package v1

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.WorkbenchServiceServer = (*WorkbenchController)(nil)

// WorkbenchController is the workbench service controller handler.
type WorkbenchController struct {
	workbench service.Workbencher
}

// NewWorkbenchController returns a fresh admin service controller instance.
func NewWorkbenchController(workbench service.Workbencher) WorkbenchController {
	return WorkbenchController{workbench: workbench}
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

func (c WorkbenchController) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	workbench, err := converter.WorkbenchToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	workbench.TenantID = tenantID

	updatedWorkbench, err := c.workbench.UpdateWorkbench(ctx, workbench)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateWorkbench': %v", err.Error())
	}

	updatedWorkbenchProto, err := converter.WorkbenchFromBusiness(updatedWorkbench)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.UpdateWorkbenchReply{Result: &chorus.UpdateWorkbenchResult{Workbench: updatedWorkbenchProto}}, nil
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

	filter := service.WorkbenchFilter{}
	if req.Filter != nil {
		filter.WorkspaceIDsIn = &req.Filter.WorkspaceIdsIn
	}

	res, paginationRes, err := c.workbench.ListWorkbenchs(ctx, tenantID, &pagination, filter)
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

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListWorkbenchsReply{Result: &chorus.ListWorkbenchsResult{Workbenchs: workbenchs}, Pagination: paginationResult}, nil
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

	newWorkbench, err := c.workbench.CreateWorkbench(ctx, workbench)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'CreateWorkbench': %v", err.Error())
	}

	newWorkbenchProto, err := converter.WorkbenchFromBusiness(newWorkbench)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateWorkbenchReply{Result: &chorus.CreateWorkbenchResult{Workbench: newWorkbenchProto}}, nil
}

func (c WorkbenchController) ManageUserRoleInWorkbench(ctx context.Context, req *chorus.ManageUserRoleInWorkbenchRequest) (*chorus.ManageUserRoleInWorkbenchReply, error) {
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

	role, err := authorization.ToRole(req.Role.Name, map[string]string{"workbench": fmt.Sprintf("%d", req.Id)})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract role from request")
	}

	err = c.workbench.ManageUserRoleInWorkbench(ctx, tenantID, userID, user_model.UserRole{Role: role})
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ManageUserRoleInWorkbench': %v", err.Error())
	}

	return &chorus.ManageUserRoleInWorkbenchReply{Result: &chorus.ManageUserRoleInWorkbenchResult{}}, nil
}

func (c WorkbenchController) RemoveUserFromWorkbench(ctx context.Context, req *chorus.RemoveUserFromWorkbenchRequest) (*chorus.RemoveUserFromWorkbenchReply, error) {
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

	err = c.workbench.RemoveUserFromWorkbench(ctx, tenantID, userID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'RemoveUserFromWorkbench': %v", err.Error())
	}

	return &chorus.RemoveUserFromWorkbenchReply{Result: &chorus.RemoveUserFromWorkbenchResult{}}, nil
}
