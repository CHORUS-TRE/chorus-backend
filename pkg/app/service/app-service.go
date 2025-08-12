package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
)

type Apper interface {
	GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error)
	CreateApp(ctx context.Context, app *model.App) (*model.App, error)
	UpdateApp(ctx context.Context, app *model.App) (*model.App, error)
	DeleteApp(ctx context.Context, tenantId, appId uint64) error
}

type AppStore interface {
	GetApp(ctx context.Context, tenantID uint64, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error)
	CreateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error)
	UpdateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error)
	DeleteApp(ctx context.Context, tenantID uint64, appID uint64) error
}

type AppService struct {
	store  AppStore
	client k8s.K8sClienter
}

func NewAppService(store AppStore, client k8s.K8sClienter) *AppService {
	return &AppService{
		store:  store,
		client: client,
	}
}

func (u *AppService) ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error) {
	apps, paginationRes, err := u.store.ListApps(ctx, tenantID, pagination)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query apps: %w", err)
	}
	return apps, paginationRes, nil
}

func (u *AppService) GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error) {
	app, err := u.store.GetApp(ctx, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app %v: %w", appID, err)
	}

	return app, nil
}

func (u *AppService) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	err := u.store.DeleteApp(ctx, tenantID, appID)
	if err != nil {
		return fmt.Errorf("unable to get app %v: %w", appID, err)
	}

	return nil
}

func (u *AppService) UpdateApp(ctx context.Context, app *model.App) (*model.App, error) {
	updatedApp, err := u.store.UpdateApp(ctx, app.TenantID, app)
	if err != nil {
		return nil, fmt.Errorf("unable to update app %v: %w", app.ID, err)
	}

	go func() {
		u.client.PrePullImageOnAllNodes(dockerImageToString(app))
	}()

	return updatedApp, nil
}

func (u *AppService) CreateApp(ctx context.Context, app *model.App) (*model.App, error) {
	newApp, err := u.store.CreateApp(ctx, app.TenantID, app)
	if err != nil {
		return nil, fmt.Errorf("unable to create app %v: %w", app.Name, err)
	}

	go func() {
		u.client.PrePullImageOnAllNodes(dockerImageToString(app))
	}()

	return newApp, nil
}

// dockerImageToString constructs the full Docker image name
func dockerImageToString(app *model.App) string {
	return app.DockerImageRegistry + "/" + app.DockerImageName + ":" + app.DockerImageTag
}
