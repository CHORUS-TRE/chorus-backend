package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/pkg/errors"
)

type AppInstanceer interface {
	GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error)
	ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error)
	CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error)
	UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error
	DeleteAppInstance(ctx context.Context, tenantId, appInstanceId uint64) error
}

type AppInstanceStore interface {
	GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error)
	ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error)
	CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (uint64, error)
	UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) error
	DeleteAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) error
}

type AppInstanceService struct {
	store  AppInstanceStore
	client helm.HelmClienter
	apper  service.Apper
}

func NewAppInstanceService(store AppInstanceStore, client helm.HelmClienter, apper service.Apper) *AppInstanceService {
	return &AppInstanceService{
		store:  store,
		client: client,
		apper:  apper,
	}
}

func (s *AppInstanceService) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	appInstances, err := s.store.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query appInstances")
	}
	return appInstances, nil
}

func (s *AppInstanceService) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get appInstance %v", appInstanceID)
	}

	return appInstance, nil
}

func (s *AppInstanceService) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	err := s.store.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return errors.Wrapf(err, "unable to get appInstance %v", appInstanceID)
	}

	return nil
}

func (s *AppInstanceService) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error {
	if err := s.store.UpdateAppInstance(ctx, appInstance.TenantID, appInstance); err != nil {
		return errors.Wrapf(err, "unable to update appInstance %v", appInstance.ID)
	}

	return nil
}

func (s *AppInstanceService) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error) {
	id, err := s.store.CreateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create appInstance %v", appInstance.ID)
	}

	app, err := s.apper.GetApp(ctx, appInstance.TenantID, appInstance.AppID)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get app %v", appInstance.AppID)
	}

	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	err = s.client.CreateAppInstance(wsName, wbName, app.Name, app.GetImage())
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create app instance %v", id)
	}

	return id, nil
}

func (s *AppInstanceService) getWorkspaceName(id uint64) string {
	return "workspace" + fmt.Sprintf("%v", id)
}

func (s *AppInstanceService) getWorkbenchName(id uint64) string {
	return "workbench" + fmt.Sprintf("%v", id)
}
