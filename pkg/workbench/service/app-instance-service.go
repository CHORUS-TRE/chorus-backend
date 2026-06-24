package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
)

func (s *WorkbenchService) ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.AppInstanceFilter) ([]*model.AppInstance, *common_model.PaginationResult, error) {
	appInstances, paginationRes, err := s.store.ListAppInstances(ctx, tenantID, pagination, filter.WorkbenchIDsIn)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, "Unable to list appInstances")
	}
	return appInstances, paginationRes, nil
}

func (s *WorkbenchService) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get appInstance %v", appInstanceID))
	}

	return appInstance, nil
}

func (s *WorkbenchService) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get appInstance %v", appInstanceID))
	}

	err = s.store.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to delete appInstance %v", appInstanceID))
	}

	// Set appInstance state to Stopped
	appInstance.K8sState = model.K8sAppInstanceStateStopped

	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(ctx, appInstance)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get k8s app instance %v", appInstance.AppID))
	}

	err = s.client.UpdateAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to delete app instance %v", appInstance.ID))
	}

	return appInstance, nil
}

func (s *WorkbenchService) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	updatedAppInstance, err := s.store.UpdateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to update appInstance %v", appInstance.ID))
	}

	return updatedAppInstance, nil
}

func (s *WorkbenchService) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	app, err := s.apper.GetApp(ctx, appInstance.TenantID, appInstance.AppID)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get app %v", appInstance.AppID))
	}

	if app.BrowserConfigJWTOIDCClientID != "" {
		token, _, err := s.authenticator.GetShortLivedTokenForClient(ctx, app.BrowserConfigJWTOIDCClientID, appInstance.WorkspaceID)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get short lived token for app %v", appInstance.AppID))
		}
		appInstance.BrowserConfigJWTToken = token
	}

	newAppInstance, err := s.store.CreateAppInstance(ctx, appInstance.TenantID, appInstance)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to create appInstance %v", appInstance.ID))
	}

	wsName := s.getWorkspaceName(newAppInstance.WorkspaceID)
	wbName := s.getWorkbenchName(newAppInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(ctx, newAppInstance)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get app %v", newAppInstance.ID))
	}

	err = s.client.CreateAppInstance(wsName, wbName, clientApp)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create app instance %v", newAppInstance.ID))
	}

	return newAppInstance, nil
}

func (s *WorkbenchService) getK8sAppInstance(ctx context.Context, appInstance *model.AppInstance) (k8s.AppInstance, error) {
	app, err := s.apper.GetApp(ctx, appInstance.TenantID, appInstance.AppID)
	if err != nil {
		return k8s.AppInstance{}, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get app %v", appInstance.AppID))
	}

	clientApp := k8s.AppInstance{
		ID:      appInstance.ID,
		AppName: app.Name,

		AppRegistry: app.DockerImageRegistry,
		AppImage:    app.DockerImageName,
		AppTag:      app.DockerImageTag,

		K8sState: appInstance.K8sState.String(),

		BrowserConfigURL:      app.BrowserConfigURL,
		BrowserConfigJWTURL:   app.BrowserConfigJWTURL,
		BrowserConfigJWTToken: appInstance.BrowserConfigJWTToken,

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
		return 0, cerr.ErrInternal.Wrap(err, "Unable to compile regex")
	}

	matches := re.FindStringSubmatch(name)
	if len(matches) != 2 {
		return 0, cerr.ErrInternal.WithMessage(fmt.Sprintf("No match found for regex with prefix %q in name %q", prefix, name))
	}

	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, cerr.ErrInternal.Wrap(err, "Unable to parse id")
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
