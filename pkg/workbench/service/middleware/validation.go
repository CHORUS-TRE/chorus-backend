package middleware

import (
	"context"
	"net/http"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Workbencher
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Workbencher) service.Workbencher {
	return func(next service.Workbencher) service.Workbencher {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, err
	}
	return v.next.ListWorkbenchs(ctx, tenantID, pagination)
}
func (v validation) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	return v.next.ProxyWorkbench(ctx, tenantID, workbenchID, w, r)
}

func (v validation) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	return v.next.GetWorkbench(ctx, tenantID, workbenchID)
}

func (v validation) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	return v.next.DeleteWorkbench(ctx, tenantID, workbenchID)
}

func (v validation) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	if err := v.validate.Struct(workbench); err != nil {
		return v.next.UpdateWorkbench(ctx, workbench)
	}
	return v.next.UpdateWorkbench(ctx, workbench)
}

func (v validation) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	if err := v.validate.Struct(workbench); err != nil {
		return 0, err
	}
	return v.next.CreateWorkbench(ctx, workbench)
}
