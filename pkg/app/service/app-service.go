package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
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
		return nil, fmt.Errorf("unable to query apps: %w", err)
	}
	return apps, nil
}

func (u *AppService) GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error) {
	app, err := u.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app %v: %w", app.ID, err)
	}

	return app, nil
}

func (u *AppService) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	err := u.store.DeleteApp(ctx, tenantID, appID)
	if err != nil {
		return fmt.Errorf("unable to get app %v, %w", appID, err)
	}

	return nil
}

func (u *AppService) UpdateApp(ctx context.Context, app *model.App) error {
	if err := u.store.UpdateApp(ctx, app.TenantID, app); err != nil {
		return fmt.Errorf("unable to update app %v: %w", app.ID, err)
	}

	return nil
}

func (u *AppService) CreateApp(ctx context.Context, app *model.App) (uint64, error) {
	id, err := u.store.CreateApp(ctx, app.TenantID, app)
	if err != nil {
		return 0, fmt.Errorf("unable to create app %v: %w", app.Name, err)
	}

	return id, nil
}
