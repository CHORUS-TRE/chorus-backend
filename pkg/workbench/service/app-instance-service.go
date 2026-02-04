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

func (s *WorkbenchService) ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter AppInstanceFilter) ([]*model.AppInstance, *common_model.PaginationResult, error) {
	appInstances, paginationRes, err := s.store.ListAppInstances(ctx, tenantID, pagination, filter.WorkbenchIDsIn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query appInstances: %w", err)
	}
	return appInstances, paginationRes, nil
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
		return fmt.Errorf("unable to delete appInstance %v: %w", appInstanceID, err)
	}

	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(ctx, appInstance)
	if err != nil {
		return fmt.Errorf("unable to get k8s app instance %v: %w", appInstance.AppID, err)
	}

	err = s.client.DeleteAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return fmt.Errorf("unable to delete app instance %v: %w", appInstance.ID, err)
	}

	return nil
}

func (s *WorkbenchService) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	updatedAppInstance, err := s.store.UpdateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return nil, fmt.Errorf("unable to update appInstance %v: %w", appInstance.ID, err)
	}

	return updatedAppInstance, nil
}

func (s *WorkbenchService) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	app, err := s.apper.GetApp(ctx, appInstance.TenantID, appInstance.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app %v: %w", appInstance.AppID, err)
	}

	if app.KioskConfigJWTOIDCClientID != "" {
		token, _, err := s.authenticator.GetShortLivedTokenForClient(ctx, app.KioskConfigJWTOIDCClientID, appInstance.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("unable to get short lived token for app %v: %w", appInstance.AppID, err)
		}
		appInstance.KioskConfigJWTToken = token
	}

	newAppInstance, err := s.store.CreateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return nil, fmt.Errorf("unable to create appInstance %v: %w", appInstance.ID, err)
	}

	wsName := s.getWorkspaceName(newAppInstance.WorkspaceID)
	wbName := s.getWorkbenchName(newAppInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(ctx, newAppInstance)
	if err != nil {
		return nil, fmt.Errorf("unable to get app %v: %w", newAppInstance.ID, err)
	}

	err = s.client.CreateAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return nil, fmt.Errorf("unable to create app instance %v: %w", newAppInstance.ID, err)
	}

	return newAppInstance, nil
}

func (s *WorkbenchService) getK8sAppInstance(ctx context.Context, appInstance *model.AppInstance) (k8s.AppInstance, error) {
	app, err := s.apper.GetApp(ctx, appInstance.TenantID, appInstance.AppID)
	if err != nil {
		return k8s.AppInstance{}, fmt.Errorf("unable to get app %v: %w", appInstance.AppID, err)
	}

	clientApp := k8s.AppInstance{
		ID:      appInstance.ID,
		AppName: app.Name,

		AppRegistry: app.DockerImageRegistry,
		AppImage:    app.DockerImageName,
		AppTag:      app.DockerImageTag,

		KioskConfigURL:      app.KioskConfigURL,
		KioskConfigJWTURL:   app.KioskConfigJWTURL,
		KioskConfigJWTToken: appInstance.KioskConfigJWTToken,

		ShmSize:             app.ShmSize,
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
