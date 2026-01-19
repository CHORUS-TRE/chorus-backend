package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.RequestServiceServer = (*RequestController)(nil)

type RequestController struct {
	request service.Requester
}

func NewRequestController(request service.Requester) RequestController {
	return RequestController{request: request}
}

func (c RequestController) GetRequest(ctx context.Context, req *chorus.GetRequestRequest) (*chorus.GetRequestReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	request, err := c.request.GetRequest(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetRequest': %v", err.Error())
	}

	protoRequest, err := converter.RequestFromBusiness(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetRequestReply{Result: &chorus.GetRequestResult{Request: protoRequest}}, nil
}

func (c RequestController) ListRequests(ctx context.Context, req *chorus.ListRequestsRequest) (*chorus.ListRequestsReply, error) {
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

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter := service.RequestFilter{}
	if req.Filter != nil {
		if len(req.Filter.StatusesIn) > 0 {
			statuses := make([]model.RequestStatus, len(req.Filter.StatusesIn))
			for i, s := range req.Filter.StatusesIn {
				statuses[i] = converter.RequestStatusToBusiness(s)
			}
			filter.StatusesIn = &statuses
		}
		if len(req.Filter.TypesIn) > 0 {
			types := make([]model.RequestType, len(req.Filter.TypesIn))
			for i, t := range req.Filter.TypesIn {
				types[i] = converter.RequestTypeToBusiness(t)
			}
			filter.TypesIn = &types
		}
		if req.Filter.SourceWorkspaceId != nil {
			filter.SourceWorkspaceID = req.Filter.SourceWorkspaceId
		}
		if req.Filter.PendingApproval != nil {
			filter.PendingApproval = req.Filter.PendingApproval
		}
	}

	res, paginationRes, err := c.request.ListRequests(ctx, tenantID, userID, &pagination, filter)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListRequests': %v", err.Error())
	}

	var requests []*chorus.Request
	for _, r := range res {
		protoRequest, err := converter.RequestFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		requests = append(requests, protoRequest)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListRequestsReply{Result: &chorus.ListRequestsResult{Requests: requests}, Pagination: paginationResult}, nil
}

func (c RequestController) CreateRequest(ctx context.Context, req *chorus.CreateRequestRequest) (*chorus.CreateRequestReply, error) {
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

	request := &model.Request{
		TenantID:               tenantID,
		RequesterID:            userID,
		SourceWorkspaceID:      req.SourceWorkspaceId,
		DestinationWorkspaceID: req.DestinationWorkspaceId,
		Type:                   converter.RequestTypeToBusiness(req.Type),
		Title:                  req.Title,
		Description:            req.Description,
		ApproverIDs:            req.ApproverIds,
	}

	createdRequest, err := c.request.CreateRequest(ctx, request, req.FilePaths)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateRequest': %v", err.Error())
	}

	protoRequest, err := converter.RequestFromBusiness(createdRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateRequestReply{Result: &chorus.CreateRequestResult{Request: protoRequest}}, nil
}

func (c RequestController) ApproveRequest(ctx context.Context, req *chorus.ApproveRequestRequest) (*chorus.ApproveRequestReply, error) {
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

	updatedRequest, err := c.request.ApproveRequest(ctx, tenantID, req.Id, userID, req.Approve)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ApproveRequest': %v", err.Error())
	}

	protoRequest, err := converter.RequestFromBusiness(updatedRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.ApproveRequestReply{Result: &chorus.ApproveRequestResult{Request: protoRequest}}, nil
}

func (c RequestController) DeleteRequest(ctx context.Context, req *chorus.DeleteRequestRequest) (*chorus.DeleteRequestReply, error) {
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

	err = c.request.DeleteRequest(ctx, tenantID, req.Id, userID)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteRequest': %v", err.Error())
	}

	return &chorus.DeleteRequestReply{Result: &chorus.DeleteRequestResult{}}, nil
}
