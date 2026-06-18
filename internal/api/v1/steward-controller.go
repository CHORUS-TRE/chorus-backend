package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	"github.com/CHORUS-TRE/chorus-backend/pkg/steward/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.StewardServiceServer = (*StewardController)(nil)

type StewardController struct {
	stewarder service.Stewarder
}

func NewStewardController(stewarder service.Stewarder) StewardController {
	return StewardController{stewarder: stewarder}
}

func (s StewardController) InitializeTenant(ctx context.Context, req *chorus.InitializeTenantRequest) (*chorus.InitializeTenantReply, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: nil")
	}

	tenant, err := s.stewarder.InitializeNewTenant(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	result, err := converter.TenantFromBusiness(tenant)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &chorus.InitializeTenantReply{Result: result}, nil
}
