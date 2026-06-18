package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
)

func TenantFromBusiness(t *tenant_model.Tenant) (*chorus.InitializeTenantResult, error) {
	ca, err := ToProtoTimestamp(t.CreationDate)
	if err != nil {
		return nil, fmt.Errorf("unable to convert created_at timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(t.UpdateDate)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updated_at timestamp: %w", err)
	}

	return &chorus.InitializeTenantResult{
		Id:        t.ID,
		Name:      t.Name,
		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
