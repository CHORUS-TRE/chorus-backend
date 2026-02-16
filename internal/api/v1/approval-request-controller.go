package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.ApprovalRequestServiceServer = (*ApprovalRequestController)(nil)

type ApprovalRequestController struct {
	approvalRequest service.ApprovalRequester
}

func NewApprovalRequestController(approvalRequest service.ApprovalRequester) ApprovalRequestController {
	return ApprovalRequestController{approvalRequest: approvalRequest}
}

func (c ApprovalRequestController) GetApprovalRequest(ctx context.Context, req *chorus.GetApprovalRequestRequest) (*chorus.GetApprovalRequestReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	request, err := c.approvalRequest.GetApprovalRequest(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'GetApprovalRequest': %v", err.Error())
	}

	protoRequest, err := converter.ApprovalRequestFromBusiness(request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.GetApprovalRequestReply{Result: &chorus.GetApprovalRequestResult{ApprovalRequest: protoRequest}}, nil
}

func (c ApprovalRequestController) ListApprovalRequests(ctx context.Context, req *chorus.ListApprovalRequestsRequest) (*chorus.ListApprovalRequestsReply, error) {
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

	filter := service.ApprovalRequestFilter{}
	if req.Filter != nil {
		if len(req.Filter.StatusesIn) > 0 {
			statuses := make([]model.ApprovalRequestStatus, len(req.Filter.StatusesIn))
			for i, s := range req.Filter.StatusesIn {
				statuses[i] = converter.ApprovalRequestStatusToBusiness(s)
			}
			filter.StatusesIn = &statuses
		}
		if len(req.Filter.TypesIn) > 0 {
			types := make([]model.ApprovalRequestType, len(req.Filter.TypesIn))
			for i, t := range req.Filter.TypesIn {
				types[i] = converter.ApprovalRequestTypeToBusiness(t)
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

	res, paginationRes, err := c.approvalRequest.ListApprovalRequests(ctx, tenantID, userID, &pagination, filter)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListApprovalRequests': %v", err.Error())
	}

	var requests []*chorus.ApprovalRequest
	for _, r := range res {
		protoRequest, err := converter.ApprovalRequestFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		requests = append(requests, protoRequest)
	}

	var paginationResult *chorus.PaginationResult
	if paginationRes != nil {
		paginationResult = converter.PaginationResultFromBusiness(paginationRes)
	}

	return &chorus.ListApprovalRequestsReply{Result: &chorus.ListApprovalRequestsResult{ApprovalRequests: requests}, Pagination: paginationResult}, nil
}

func (c ApprovalRequestController) CreateDataExtractionRequest(ctx context.Context, req *chorus.CreateDataExtractionRequestRequest) (*chorus.CreateDataExtractionRequestReply, error) {
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

	request := &model.ApprovalRequest{
		TenantID:    tenantID,
		RequesterID: userID,
		Title:       req.Title,
		Description: req.Description,
		Details: model.ApprovalRequestDetails{
			DataExtractionDetails: &model.DataExtractionDetails{
				SourceWorkspaceID: req.SourceWorkspaceId,
			},
		},
	}

	createdRequest, err := c.approvalRequest.CreateDataExtractionRequest(ctx, request, req.FilePaths)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateDataExtractionRequest': %v", err.Error())
	}

	protoRequest, err := converter.ApprovalRequestFromBusiness(createdRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateDataExtractionRequestReply{Result: &chorus.CreateDataExtractionRequestResult{ApprovalRequest: protoRequest}}, nil
}

func (c ApprovalRequestController) CreateDataTransferRequest(ctx context.Context, req *chorus.CreateDataTransferRequestRequest) (*chorus.CreateDataTransferRequestReply, error) {
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

	request := &model.ApprovalRequest{
		TenantID:    tenantID,
		RequesterID: userID,
		Title:       req.Title,
		Description: req.Description,
		Details: model.ApprovalRequestDetails{
			DataTransferDetails: &model.DataTransferDetails{
				SourceWorkspaceID:      req.SourceWorkspaceId,
				DestinationWorkspaceID: req.DestinationWorkspaceId,
			},
		},
	}

	createdRequest, err := c.approvalRequest.CreateDataTransferRequest(ctx, request, req.FilePaths)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'CreateDataTransferRequest': %v", err.Error())
	}

	protoRequest, err := converter.ApprovalRequestFromBusiness(createdRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.CreateDataTransferRequestReply{Result: &chorus.CreateDataTransferRequestResult{ApprovalRequest: protoRequest}}, nil
}

func (c ApprovalRequestController) ApproveApprovalRequest(ctx context.Context, req *chorus.ApproveApprovalRequestRequest) (*chorus.ApproveApprovalRequestReply, error) {
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

	updatedRequest, err := c.approvalRequest.ApproveApprovalRequest(ctx, tenantID, req.Id, userID, req.Approve)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ApproveApprovalRequest': %v", err.Error())
	}

	protoRequest, err := converter.ApprovalRequestFromBusiness(updatedRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
	}

	return &chorus.ApproveApprovalRequestReply{Result: &chorus.ApproveApprovalRequestResult{ApprovalRequest: protoRequest}}, nil
}

func (c ApprovalRequestController) DeleteApprovalRequest(ctx context.Context, req *chorus.DeleteApprovalRequestRequest) (*chorus.DeleteApprovalRequestReply, error) {
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

	err = c.approvalRequest.DeleteApprovalRequest(ctx, tenantID, req.Id, userID)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DeleteApprovalRequest': %v", err.Error())
	}

	return &chorus.DeleteApprovalRequestReply{Result: &chorus.DeleteApprovalRequestResult{}}, nil
}

func (c ApprovalRequestController) DownloadApprovalRequestFile(ctx context.Context, req *chorus.DownloadApprovalRequestFileRequest) (*chorus.DownloadApprovalRequestFileReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	file, content, err := c.approvalRequest.DownloadApprovalRequestFile(ctx, tenantID, req.Id, req.Path)
	if err != nil {
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'DownloadApprovalRequestFile': %v", err.Error())
	}

	return &chorus.DownloadApprovalRequestFileReply{
		Result: &chorus.DownloadApprovalRequestFileResult{
			File:    converter.FileFromBusiness(file),
			Content: content,
		},
	}, nil
}
