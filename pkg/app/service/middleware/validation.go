package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Apper
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Apper) service.Apper {
	return func(next service.Apper) service.Apper {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, err
	}
	return v.next.ListApps(ctx, tenantID, pagination)
}

func (v validation) GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error) {
	return v.next.GetApp(ctx, tenantID, appID)
}

func (v validation) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	return v.next.DeleteApp(ctx, tenantID, appID)
}

func (v validation) UpdateApp(ctx context.Context, app *model.App) error {
	if err := v.validate.Struct(app); err != nil {
		return v.next.UpdateApp(ctx, app)
	}
	return v.next.UpdateApp(ctx, app)
}

func (v validation) CreateApp(ctx context.Context, app *model.App) (uint64, error) {
	if err := v.validate.Struct(app); err != nil {
		return 0, err
	}
	return v.next.CreateApp(ctx, app)
}
