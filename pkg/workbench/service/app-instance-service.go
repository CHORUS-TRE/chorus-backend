package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
)

func (s *WorkbenchService) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	appInstances, err := s.store.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		return nil, fmt.Errorf("unable to query appInstances: %w", err)
	}
	return appInstances, nil
}

func (s *WorkbenchService) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return nil, fmt.Errorf("unable to get appInstance %v: %w", appInstanceID, err)
	}

	return appInstance, nil
}

func (s *WorkbenchService) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return fmt.Errorf("unable to get appInstance %v: %w", appInstanceID, err)
	}

	err = s.store.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return fmt.Errorf("unable to get appInstance %v: %w", appInstanceID, err)
	}

	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(appInstance.TenantID, appInstance.AppID, appInstance.ID)
	if err != nil {
		return fmt.Errorf("unable to get k8s app instance %v: %w", appInstance.AppID, err)
	}

	err = s.client.DeleteAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return fmt.Errorf("unable to delete app instance %v: %w", appInstance.ID, err)
	}

	return nil
}

func (s *WorkbenchService) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error {
	if err := s.store.UpdateAppInstance(ctx, appInstance.TenantID, appInstance); err != nil {
		return fmt.Errorf("unable to update appInstance %v: %w", appInstance.ID, err)
	}

	return nil
}

func (s *WorkbenchService) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error) {
	id, err := s.store.CreateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return 0, fmt.Errorf("unable to create appInstance %v: %w", appInstance.ID, err)
	}

	appInstance.ID = id

	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(appInstance.TenantID, appInstance.AppID, appInstance.ID)
	if err != nil {
		return 0, fmt.Errorf("unable to get app %v: %w", id, err)
	}

	err = s.client.CreateAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return 0, fmt.Errorf("unable to create app instance %v: %w", id, err)
	}

	return id, nil
}

func (s *WorkbenchService) getK8sAppInstance(tenantID, appID, appInstanceID uint64) (k8s.AppInstance, error) {
	app, err := s.apper.GetApp(context.Background(), tenantID, appID)
	if err != nil {
		return k8s.AppInstance{}, fmt.Errorf("unable to get app %v: %w", appID, err)
	}

	clientApp := k8s.AppInstance{
		ID:      appInstanceID,
		AppName: app.Name,

		AppRegistry: app.DockerImageRegistry,
		AppImage:    app.DockerImageName,
		AppTag:      app.DockerImageTag,

		ShmSize:             app.ShmSize,
		KioskConfigURL:      app.KioskConfigURL,
		MaxCPU:              app.MaxCPU,
		MinCPU:              app.MinCPU,
		MaxMemory:           app.MaxMemory,
		MinMemory:           app.MinMemory,
		MaxEphemeralStorage: app.MaxEphemeralStorage,
		MinEphemeralStorage: app.MinEphemeralStorage,
	}

	return clientApp, nil
}

func (s *WorkbenchService) getIDWithPrefix(prefix, name string) (uint64, error) {
	re, err := regexp.Compile("^" + prefix + "([0-9]+)$")
	if err != nil {
		return 0, fmt.Errorf("unable to compile regex: %w", err)
	}

	matches := re.FindStringSubmatch(name)
	if len(matches) != 2 {
		return 0, fmt.Errorf("no match found for regex with prefix %q in name %q", prefix, name)
	}

	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse id: %w", err)
	}

	return id, nil
}

func (s *WorkbenchService) getWorkspaceName(id uint64) string {
	return fmt.Sprintf("workspace%v", id)
}

func (s *WorkbenchService) getWorkspaceID(name string) (uint64, error) {
	return s.getIDWithPrefix("workspace", name)
}

func (s *WorkbenchService) getWorkbenchName(id uint64) string {
	return fmt.Sprintf("workbench%v", id)
}

func (s *WorkbenchService) getWorkbenchID(name string) (uint64, error) {
	return s.getIDWithPrefix("workbench", name)
}
