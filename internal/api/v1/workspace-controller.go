package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"
	workspace_service "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.WorkspaceServiceServer = (*WorkspaceController)(nil)

// WorkspaceController is the workspace service controller handler.
type WorkspaceController struct {
	workspace service.Workspaceer
}

// NewWorkspaceController returns a fresh admin service controller instance.
func NewWorkspaceController(workspace service.Workspaceer) WorkspaceController {
	return WorkspaceController{workspace: workspace}
}

func (c WorkspaceController) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	workspace, err := c.workspace.GetWorkspace(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetWorkspace': %v", err.Error())
	}

	tgWorkspace, err := converter.WorkspaceFromBusiness(workspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetWorkspaceReply{Result: &chorus.GetWorkspaceResult{Workspace: tgWorkspace}}, nil
}

func (c WorkspaceController) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	workspace, err := converter.WorkspaceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	workspace.TenantID = tenantID

	updatedWorkspace, err := c.workspace.UpdateWorkspace(ctx, workspace)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateWorkspace': %v", err.Error())
	}

	tgWorkspace, err := converter.WorkspaceFromBusiness(updatedWorkspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.UpdateWorkspaceReply{Result: &chorus.UpdateWorkspaceResult{Workspace: tgWorkspace}}, nil
}

func (c WorkspaceController) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.workspace.DeleteWorkspace(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteWorkspace': %v", err.Error())
	}
	return &chorus.DeleteWorkspaceReply{Result: &chorus.DeleteWorkspaceResult{}}, nil
}

// ListWorkspaces extracts the retrieved workspaces from the service and inserts them into a reply object.
func (c WorkspaceController) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter := workspace_service.WorkspaceFilter{}
	if req.Filter != nil {
		filter.WorkspaceIDsIn = &req.Filter.WorkspaceIdsIn
	}

	res, paginationRes, err := c.workspace.ListWorkspaces(ctx, tenantID, &pagination, filter)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListWorkspaces': %v", err.Error())
	}

	var workspaces []*chorus.Workspace
	for _, r := range res {
		workspace, err := converter.WorkspaceFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		workspaces = append(workspaces, workspace)
	}

	paginationResult := converter.PaginationResultFromBusiness(paginationRes)

	return &chorus.ListWorkspacesReply{Result: &chorus.ListWorkspacesResult{Workspaces: workspaces}, Pagination: paginationResult}, nil
}

// CreateWorkspace extracts the workspace from the request and passes it to the workspace service.
func (c WorkspaceController) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
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

	workspace, err := converter.WorkspaceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	workspace.TenantID = tenantID
	workspace.UserID = userID

	newWorkspace, err := c.workspace.CreateWorkspace(ctx, workspace)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateWorkspace': %v", err.Error())
	}

	tgWorkspace, err := converter.WorkspaceFromBusiness(newWorkspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateWorkspaceReply{Result: &chorus.CreateWorkspaceResult{Workspace: tgWorkspace}}, nil
}
