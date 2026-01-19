package middleware

import (
	"context"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	service "github.com/CHORUS-TRE/chorus-backend/pkg/common/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/model"
	request_service "github.com/CHORUS-TRE/chorus-backend/pkg/request/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     request_service.Requester
	validate *val.Validate
}

func Validation(validate *val.Validate) func(request_service.Requester) request_service.Requester {
	return func(next request_service.Requester) request_service.Requester {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) GetRequest(ctx context.Context, tenantID, requestID uint64) (*model.Request, error) {
	if requestID == 0 {
		return nil, &service.InvalidParametersErr{}
	}
	return v.next.GetRequest(ctx, tenantID, requestID)
}

func (v validation) ListRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter request_service.RequestFilter) ([]*model.Request, *common_model.PaginationResult, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, nil, err
	}
	return v.next.ListRequests(ctx, tenantID, userID, pagination, filter)
}

func (v validation) CreateRequest(ctx context.Context, request *model.Request, filePaths []string) (*model.Request, error) {
	if request == nil {
		return nil, &service.InvalidParametersErr{}
	}
	if request.SourceWorkspaceID == 0 {
		return nil, &service.InvalidParametersErr{}
	}
	if len(filePaths) == 0 {
		return nil, &service.InvalidParametersErr{}
	}
	if request.Type == model.RequestTypeUnspecified {
		return nil, &service.InvalidParametersErr{}
	}
	if request.Type == model.RequestTypeCopyToWorkspace && request.DestinationWorkspaceID == nil {
		return nil, &service.InvalidParametersErr{}
	}
	return v.next.CreateRequest(ctx, request, filePaths)
}

func (v validation) ApproveRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.Request, error) {
	if requestID == 0 {
		return nil, &service.InvalidParametersErr{}
	}
	return v.next.ApproveRequest(ctx, tenantID, requestID, userID, approve)
}

func (v validation) DeleteRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	if requestID == 0 {
		return &service.InvalidParametersErr{}
	}
	return v.next.DeleteRequest(ctx, tenantID, requestID, userID)
}
