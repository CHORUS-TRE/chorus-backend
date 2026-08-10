package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/harbor"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/ociregistry"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
)

type Apper interface {
	GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error)
	CreateApp(ctx context.Context, app *model.App) (*model.App, error)
	BulkCreateApps(ctx context.Context, apps []*model.App) ([]*model.App, error)
	UpdateApp(ctx context.Context, app *model.App) (*model.App, error)
	DeleteApp(ctx context.Context, tenantId, appId uint64) error
}

type AppStore interface {
	GetApp(ctx context.Context, tenantID uint64, appID uint64) (*model.App, error)
	ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error)
	CreateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error)
	BulkCreateApps(ctx context.Context, tenantID uint64, apps []*model.App) ([]*model.App, error)
	UpdateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error)
	DeleteApp(ctx context.Context, tenantID uint64, appID uint64) error
}

type AppService struct {
	store        AppStore
	k8sClient    k8s.K8sClienter
	ociClient    ociregistry.OCIClienter
	harborClient harbor.HarborClient
}

func NewAppService(store AppStore, k8sClient k8s.K8sClienter, ociClient ociregistry.OCIClienter, harborClient harbor.HarborClient) *AppService {
	return &AppService{
		store:        store,
		k8sClient:    k8sClient,
		ociClient:    ociClient,
		harborClient: harborClient,
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
	if err := u.validateRegistry(app); err != nil {
		return nil, err
	}

	imageRef := dockerImageToString(app)

	exists, err := u.ociClient.ImageExists(imageRef)
	if err != nil {
		return nil, fmt.Errorf("unable to verify app image existence %s: %w", imageRef, err)
	}
	if !exists {
		return nil, fmt.Errorf("app image %s does not exist", imageRef)
	}

	updatedApp, err := u.store.UpdateApp(ctx, app.TenantID, app)
	if err != nil {
		return nil, fmt.Errorf("unable to update app %s: %w", app.Name, err)
	}

	go func() {
		u.k8sClient.PrePullImageOnAllNodes(imageRef)
	}()

	return updatedApp, nil
}

func (u *AppService) CreateApp(ctx context.Context, app *model.App) (*model.App, error) {
	if err := u.validateRegistry(app); err != nil {
		return nil, err
	}

	imageRef := dockerImageToString(app)

	exists, err := u.ociClient.ImageExists(imageRef)
	if err != nil {
		return nil, fmt.Errorf("unable to verify app image existence %s: %w", imageRef, err)
	}
	if !exists {
		return nil, fmt.Errorf("app image %s does not exist", imageRef)
	}

	newApp, err := u.store.CreateApp(ctx, app.TenantID, app)
	if err != nil {
		return nil, fmt.Errorf("unable to create app %s: %w", app.Name, err)
	}

	go func() {
		u.k8sClient.PrePullImageOnAllNodes(imageRef)
	}()

	return newApp, nil
}

func (u *AppService) BulkCreateApps(ctx context.Context, apps []*model.App) ([]*model.App, error) {
	for _, app := range apps {
		if err := u.validateRegistry(app); err != nil {
			return nil, err
		}

		imageRef := dockerImageToString(app)

		exists, err := u.ociClient.ImageExists(imageRef)
		if err != nil {
			return nil, fmt.Errorf("unable to verify app image existence %s: %w", imageRef, err)
		}
		if !exists {
			return nil, fmt.Errorf("app image %s does not exist", imageRef)
		}
	}

	newApps, err := u.store.BulkCreateApps(ctx, apps[0].TenantID, apps)
	if err != nil {
		return nil, fmt.Errorf("unable to bulk create apps: %w", err)
	}

	go func() {
		for _, app := range apps {
			imageRef := dockerImageToString(app)
			u.k8sClient.PrePullImageOnAllNodes(imageRef)
		}
	}()

	return newApps, nil
}

// dockerImageToString constructs the full Docker image name
func dockerImageToString(app *model.App) string {
	return app.DockerImageRegistry + "/" + app.DockerImageName + ":" + app.DockerImageTag
}

// validateRegistry rejects an app whose registry doesn't match the one the
// OCI client is configured for
func (u *AppService) validateRegistry(app *model.App) error {
	if app.DockerImageRegistry != "" && app.DockerImageRegistry != u.ociClient.Host() {
		return fmt.Errorf("app image registry %q does not match configured registry %q", app.DockerImageRegistry, u.ociClient.Host())
	}
	return nil
}
