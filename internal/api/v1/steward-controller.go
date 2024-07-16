package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/steward/service"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StewardController struct {
	stewarder service.Stewarder
}

func (s StewardController) InitializeTenant(ctx context.Context, req *chorus.InitializeTenantRequest) (*empty.Empty, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: nil")
	}

	err := s.stewarder.InitializeNewTenant(ctx, req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &empty.Empty{}, nil
}

func NewStewardController(stewarder service.Stewarder) StewardController {
	return StewardController{stewarder: stewarder}
}
