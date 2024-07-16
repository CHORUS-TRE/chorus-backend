package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.AppInstanceer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.AppInstanceer) service.AppInstanceer {
	return func(next service.AppInstanceer) service.AppInstanceer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	if err := v.validate.Struct(pagination); err != nil {
		return nil, err
	}
	return v.next.ListAppInstances(ctx, tenantID, pagination)
}

func (v validation) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	return v.next.GetAppInstance(ctx, tenantID, appInstanceID)
}

func (v validation) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	return v.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
}

func (v validation) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error {
	if err := v.validate.Struct(appInstance); err != nil {
		return v.next.UpdateAppInstance(ctx, appInstance)
	}
	return v.next.UpdateAppInstance(ctx, appInstance)
}

func (v validation) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error) {
	if err := v.validate.Struct(appInstance); err != nil {
		return 0, err
	}
	return v.next.CreateAppInstance(ctx, appInstance)
}
