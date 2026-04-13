package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.WorkspaceServiceInstanceServiceServer = (*WorkspaceServiceInstanceController)(nil)

func NewWorkspaceServiceInstanceController(workspaceer service.Workspaceer) WorkspaceServiceInstanceController {
	return WorkspaceServiceInstanceController{workspaceer: workspaceer}
}

type WorkspaceServiceInstanceController struct {
	workspaceer service.Workspaceer
}

func (c WorkspaceServiceInstanceController) GetWorkspaceServiceInstance(ctx context.Context, req *chorus.GetWorkspaceServiceInstanceRequest) (*chorus.GetWorkspaceServiceInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	svc, err := c.workspaceer.GetWorkspaceServiceInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetWorkspaceServiceInstance': %v", err.Error())
	}

	pbSvc, err := converter.WorkspaceServiceInstanceFromBusiness(svc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetWorkspaceServiceInstanceReply{Result: &chorus.GetWorkspaceServiceInstanceResult{WorkspaceServiceInstance: pbSvc}}, nil
}

func (c WorkspaceServiceInstanceController) ListWorkspaceServiceInstances(ctx context.Context, req *chorus.ListWorkspaceServiceInstancesRequest) (*chorus.ListWorkspaceServiceInstancesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter := service.WorkspaceServiceInstanceFilter{}
	if req.Filter != nil {
		filter.WorkspaceIDsIn = &req.Filter.WorkspaceIdsIn
	}

	res, paginationRes, err := c.workspaceer.ListWorkspaceServiceInstances(ctx, tenantID, &pagination, filter)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListWorkspaceServiceInstances': %v", err.Error())
	}

	var instances []*chorus.WorkspaceServiceInstance
	for _, r := range res {
		pbSvc, err := converter.WorkspaceServiceInstanceFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		instances = append(instances, pbSvc)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListWorkspaceServiceInstancesReply{Result: &chorus.ListWorkspaceServiceInstancesResult{WorkspaceServiceInstances: instances}, Pagination: paginationResult}, nil
}

func (c WorkspaceServiceInstanceController) CreateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.CreateWorkspaceServiceInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	svc, err := converter.WorkspaceServiceInstanceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	svc.TenantID = tenantID

	created, err := c.workspaceer.CreateWorkspaceServiceInstance(ctx, svc)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateWorkspaceServiceInstance': %v", err.Error())
	}

	pbSvc, err := converter.WorkspaceServiceInstanceFromBusiness(created)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateWorkspaceServiceInstanceReply{Result: &chorus.CreateWorkspaceServiceInstanceResult{WorkspaceServiceInstance: pbSvc}}, nil
}

func (c WorkspaceServiceInstanceController) UpdateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.UpdateWorkspaceServiceInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	svc, err := converter.WorkspaceServiceInstanceToBusiness(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	svc.TenantID = tenantID

	updated, err := c.workspaceer.UpdateWorkspaceServiceInstance(ctx, svc)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'UpdateWorkspaceServiceInstance': %v", err.Error())
	}

	pbSvc, err := converter.WorkspaceServiceInstanceFromBusiness(updated)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.UpdateWorkspaceServiceInstanceReply{Result: &chorus.UpdateWorkspaceServiceInstanceResult{WorkspaceServiceInstance: pbSvc}}, nil
}

func (c WorkspaceServiceInstanceController) DeleteWorkspaceServiceInstance(ctx context.Context, req *chorus.DeleteWorkspaceServiceInstanceRequest) (*chorus.DeleteWorkspaceServiceInstanceReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	err = c.workspaceer.DeleteWorkspaceServiceInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteWorkspaceServiceInstance': %v", err.Error())
	}

	return &chorus.DeleteWorkspaceServiceInstanceReply{Result: &chorus.DeleteWorkspaceServiceInstanceResult{}}, nil
}
