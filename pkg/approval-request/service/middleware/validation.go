package middleware

import (
	"context"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	approval_request_service "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     approval_request_service.ApprovalRequester
	validate *val.Validate
}

func Validation(validate *val.Validate) func(approval_request_service.ApprovalRequester) approval_request_service.ApprovalRequester {
	return func(next approval_request_service.ApprovalRequester) approval_request_service.ApprovalRequester {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	if requestID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Request ID is required")
	}
	return v.next.GetApprovalRequest(ctx, tenantID, requestID)
}

func (v validation) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter approval_request_service.ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, nil, cerr.WrapValidationError(err)
	}
	return v.next.ListApprovalRequests(ctx, tenantID, userID, pagination, filter)
}

func (v validation) CountMyApprovalRequests(ctx context.Context, tenantID, userID uint64) (*model.ApprovalRequestCounts, error) {
	return v.next.CountMyApprovalRequests(ctx, tenantID, userID)
}

func (v validation) CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	if request == nil {
		return nil, cerr.ErrValidation.WithMessage("Request is required")
	}
	details := request.Details.DataExtractionDetails
	if details == nil {
		return nil, cerr.ErrValidation.WithMessage("Data extraction details are required")
	}
	if details.SourceWorkspaceID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Source workspace ID is required")
	}
	if len(filePaths) == 0 {
		return nil, cerr.ErrValidation.WithMessage("At least one file path is required")
	}
	return v.next.CreateDataExtractionRequest(ctx, request, filePaths)
}

func (v validation) CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	if request == nil {
		return nil, cerr.ErrValidation.WithMessage("Request is required")
	}
	details := request.Details.DataTransferDetails
	if details == nil {
		return nil, cerr.ErrValidation.WithMessage("Data transfer details are required")
	}
	if details.SourceWorkspaceID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Source workspace ID is required")
	}
	if details.DestinationWorkspaceID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Destination workspace ID is required")
	}
	if len(filePaths) == 0 {
		return nil, cerr.ErrValidation.WithMessage("At least one file path is required")
	}
	return v.next.CreateDataTransferRequest(ctx, request, filePaths)
}

func (v validation) ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error) {
	if requestID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Request ID is required")
	}
	return v.next.ApproveApprovalRequest(ctx, tenantID, requestID, userID, approve)
}

func (v validation) DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	if requestID == 0 {
		return cerr.ErrValidation.WithMessage("Request ID is required")
	}
	return v.next.DeleteApprovalRequest(ctx, tenantID, requestID, userID)
}

func (v validation) DownloadApprovalRequestFile(ctx context.Context, tenantID, requestID uint64, filePath string) (*model.ApprovalRequestFile, []byte, error) {
	if requestID == 0 {
		return nil, nil, cerr.ErrValidation.WithMessage("Request ID is required")
	}
	if filePath == "" {
		return nil, nil, cerr.ErrValidation.WithMessage("File path is required")
	}
	return v.next.DownloadApprovalRequestFile(ctx, tenantID, requestID, filePath)
}
