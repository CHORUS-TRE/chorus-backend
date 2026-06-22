package v1

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"
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
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	workbench, err := c.workbench.GetWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	tgWorkbench, err := converter.WorkbenchFromBusiness(workbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.GetWorkbenchReply{Result: &chorus.GetWorkbenchResult{Workbench: tgWorkbench}}, nil
}

func (c WorkbenchController) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	workbench, err := converter.WorkbenchToBusiness(req)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	workbench.TenantID = tenantID

	updatedWorkbench, err := c.workbench.UpdateWorkbench(ctx, workbench)
	if err != nil {
		return nil, err
	}

	updatedWorkbenchProto, err := converter.WorkbenchFromBusiness(updatedWorkbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.UpdateWorkbenchReply{Result: &chorus.UpdateWorkbenchResult{Workbench: updatedWorkbenchProto}}, nil
}

func (c WorkbenchController) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	workbench, err := c.workbench.DeleteWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	workbenchRes, err := converter.WorkbenchFromBusiness(workbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.DeleteWorkbenchReply{Result: &chorus.DeleteWorkbenchResult{Workbench: workbenchRes}}, nil
}

// ListWorkbenches extracts the retrieved workbenches from the service and inserts them into a reply object.
func (c WorkbenchController) ListWorkbenches(ctx context.Context, req *chorus.ListWorkbenchesRequest) (*chorus.ListWorkbenchesReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter := model.WorkbenchFilter{}
	if req.Filter != nil {
		filter.WorkspaceIDsIn = &req.Filter.WorkspaceIdsIn
	}

	res, paginationRes, err := c.workbench.ListWorkbenches(ctx, tenantID, &pagination, filter)
	if err != nil {
		return nil, err
	}

	var workbenches []*chorus.Workbench
	for _, r := range res {
		workbench, err := converter.WorkbenchFromBusiness(r)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
		}
		workbenches = append(workbenches, workbench)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListWorkbenchesReply{Result: &chorus.ListWorkbenchesResult{Workbenches: workbenches}, Pagination: paginationResult}, nil
}

// CreateWorkbench extracts the workbench from the request and passes it to the workbench service.
func (c WorkbenchController) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
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
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	workbench.TenantID = tenantID
	workbench.UserID = userID

	newWorkbench, err := c.workbench.CreateWorkbench(ctx, workbench)
	if err != nil {
		return nil, err
	}

	newWorkbenchProto, err := converter.WorkbenchFromBusiness(newWorkbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.CreateWorkbenchReply{Result: &chorus.CreateWorkbenchResult{Workbench: newWorkbenchProto}}, nil
}

func (c WorkbenchController) ManageUserRoleInWorkbench(ctx context.Context, req *chorus.ManageUserRoleInWorkbenchRequest) (*chorus.ManageUserRoleInWorkbenchReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	role, err := authorization.ToRole(req.Role.Name, map[string]string{"workbench": fmt.Sprintf("%d", req.Id)})
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract role from request")
	}

	err = c.workbench.ManageUserRoleInWorkbench(ctx, tenantID, req.UserId, user_model.UserRole{Role: role})
	if err != nil {
		return nil, err
	}

	workbench, err := c.workbench.GetWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	workbenchRes, err := converter.WorkbenchFromBusiness(workbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.ManageUserRoleInWorkbenchReply{Result: &chorus.ManageUserRoleInWorkbenchResult{Workbench: workbenchRes}}, nil
}

func (c WorkbenchController) RemoveUserFromWorkbench(ctx context.Context, req *chorus.RemoveUserFromWorkbenchRequest) (*chorus.RemoveUserFromWorkbenchReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	err = c.workbench.RemoveUserFromWorkbench(ctx, tenantID, req.UserId, req.Id)
	if err != nil {
		return nil, err
	}

	workbench, err := c.workbench.GetWorkbench(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	workbenchRes, err := converter.WorkbenchFromBusiness(workbench)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert workbench")
	}

	return &chorus.RemoveUserFromWorkbenchReply{Result: &chorus.RemoveUserFromWorkbenchResult{Workbench: workbenchRes}}, nil
}
