package service

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/pkg/errors"
)

type Apper interface {
	GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error)
	CreateApp(ctx context.Context, app *model.App) (uint64, error)
	UpdateApp(ctx context.Context, app *model.App) error
	DeleteApp(ctx context.Context, tenantId, appId uint64) error
}

type AppStore interface {
	GetApp(ctx context.Context, tenantID uint64, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error)
	CreateApp(ctx context.Context, tenantID uint64, app *model.App) (uint64, error)
	UpdateApp(ctx context.Context, tenantID uint64, app *model.App) error
	DeleteApp(ctx context.Context, tenantID uint64, appID uint64) error
}

type AppService struct {
	store AppStore
}

func NewAppService(store AppStore) *AppService {
	return &AppService{
		store: store,
	}
}

func (u *AppService) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error) {
	apps, err := u.store.ListApps(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query apps")
	}
	return apps, nil
}

func (u *AppService) GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error) {
	app, err := u.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get app %v", app.ID)
	}

	return app, nil
}

func (u *AppService) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	err := u.store.DeleteApp(ctx, tenantID, appID)
	if err != nil {
		return errors.Wrapf(err, "unable to get app %v", appID)
	}

	return nil
}

func (u *AppService) UpdateApp(ctx context.Context, app *model.App) error {
	if err := u.store.UpdateApp(ctx, app.TenantID, app); err != nil {
		return errors.Wrapf(err, "unable to update app %v", app.ID)
	}

	return nil
}

func (u *AppService) CreateApp(ctx context.Context, app *model.App) (uint64, error) {
	id, err := u.store.CreateApp(ctx, app.TenantID, app)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create app %v", app.Name)
	}

	return id, nil
}
