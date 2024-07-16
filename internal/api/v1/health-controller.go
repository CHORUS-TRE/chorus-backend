package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
)

type HealthController struct{}

func NewHealthController() chorus.HealthServiceServer {
	return &HealthController{}
}

func (c HealthController) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	// Endpoints in HealthController don't need to be authenticated
	return ctx, nil
}

func (c HealthController) GetHealthCheck(ctx context.Context, req *chorus.GetHealthCheckRequest) (*chorus.GetHealthCheckReply, error) {
	return &chorus.GetHealthCheckReply{}, nil
}
