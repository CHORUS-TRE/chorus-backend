package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/rest/middleware"
	app_service "github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	auth_helper "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/helper"
	authentication_service "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var _ Workbencher = (*WorkbenchService)(nil)

var (
	workbenchProxyRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "workbench_service_proxy_request",
		Help: "The total number of request proxied to a workbench via the backend",
	}, []string{"workbench_id"})

	_ = prometheus.DefaultRegisterer.Register(workbenchProxyRequest)
)

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *notification_model.Notification, userIDs []uint64) error
}

type WorkbenchFilter struct {
	WorkspaceIDsIn *[]uint64
}

type AppInstanceFilter struct {
	WorkbenchIDsIn *[]uint64
}

type Workbencher interface {
	GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkbenchFilter) ([]*model.Workbench, *common_model.PaginationResult, error)
	CreateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error)
	ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error
	UpdateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error)
	DeleteWorkbench(ctx context.Context, tenantId, workbenchId uint64) error
	DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error

	ManageUserRoleInWorkbench(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error
	RemoveUserFromWorkbench(ctx context.Context, tenantID, userID, workbenchID uint64) error

	GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error)
	ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter AppInstanceFilter) ([]*model.AppInstance, *common_model.PaginationResult, error)
	CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error)
	UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error)
	DeleteAppInstance(ctx context.Context, tenantId, appInstanceId uint64) error
}

type WorkbenchStore interface {
	GetWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workspaceIDsIn *[]uint64) ([]*model.Workbench, *common_model.PaginationResult, error)
	ListWorkbenchAppInstances(ctx context.Context, workbenchID uint64) ([]*model.AppInstance, error)
	ListAllWorkbenches(ctx context.Context) ([]*model.Workbench, error)
	SaveBatchProxyHit(ctx context.Context, proxyHitCountMap map[uint64]uint64, proxyHitDateMap map[uint64]time.Time) error
	CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (*model.Workbench, error)
	UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (*model.Workbench, error)
	DeleteWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) error
	DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
	DeleteIdleWorkbenches(ctx context.Context, idleTimeout time.Duration) ([]*model.Workbench, error)

	GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error)
	ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workbenchIDsIn *[]uint64) ([]*model.AppInstance, *common_model.PaginationResult, error)
	CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (*model.AppInstance, error)
	UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (*model.AppInstance, error)
	UpdateAppInstances(ctx context.Context, tenantID uint64, appInstances []*model.AppInstance) error
	DeleteAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) error
	DeleteAppInstances(ctx context.Context, tenantID uint64, appInstanceIDs []uint64) error
}

type proxyID struct {
	namespace string
	workbench string
}

type proxy struct {
	reverseProxy    *httputil.ReverseProxy
	forwardStopChan chan struct{}
	forwardPort     uint16
}

type WorkbenchService struct {
	cfg    config.Config
	store  WorkbenchStore
	client k8s.K8sClienter

	apper             app_service.Apper
	userer            user_service.Userer
	authenticator     authentication_service.Authenticator
	notificationStore NotificationStore

	proxyRWMutex     sync.RWMutex
	proxyCache       map[proxyID]*proxy
	proxyHitMutex    sync.Mutex
	proxyHitCountMap map[uint64]uint64
	proxyHitDateMap  map[uint64]time.Time
}

func NewWorkbenchService(cfg config.Config, store WorkbenchStore, client k8s.K8sClienter, apper app_service.Apper, userer user_service.Userer, authenticator authentication_service.Authenticator, notificationStore NotificationStore) *WorkbenchService {
	s := &WorkbenchService{
		cfg:    cfg,
		store:  store,
		client: client,

		apper:             apper,
		userer:            userer,
		authenticator:     authenticator,
		notificationStore: notificationStore,

		proxyCache:       make(map[proxyID]*proxy),
		proxyHitCountMap: make(map[uint64]uint64),
		proxyHitDateMap:  make(map[uint64]time.Time),
	}

	go func() {
		s.updateAllWorkbenches(context.Background())
	}()

	go func() {
		for {
			s.saveBatchProxyHit(context.Background())
			randomDelayToAvoidCollision := time.Duration(rand.Int64N(int64(10 * time.Second)))
			time.Sleep(cfg.Services.WorkbenchService.ProxyHitSaveBatchInterval + randomDelayToAvoidCollision)
		}
	}()

	if s.cfg.Services.WorkbenchService.WorkbenchIdleTimeout != nil {
		logger.TechLog.Info(context.Background(), "starting workbench idle cleaner", zap.Duration("idleTimeout", *s.cfg.Services.WorkbenchService.WorkbenchIdleTimeout), zap.Duration("checkInterval", s.cfg.Services.WorkbenchService.WorkbenchIdleCheckInterval))

		go func() {
			interval := cfg.Services.WorkbenchService.WorkbenchIdleCheckInterval
			// sleep a random jitter in initial interval to avoid all instances doing this at the same time
			jitter := time.Duration(rand.Int64N(int64(interval)))
			time.Sleep(jitter)

			for {
				s.cleanIdleWorkbenches(context.Background())
				time.Sleep(interval)
			}
		}()
	}

	s.SetClientWatchers()

	return s
}

func (s *WorkbenchService) SetClientWatchers() {
	watcher := func(k8sWorkbench k8s.Workbench) error {
		ctx := context.Background()
		logger.TechLog.Debug(ctx, "new/update workbench",
			zap.String("namespace", k8sWorkbench.Namespace),
			zap.String("workbenchName", k8sWorkbench.Name),
			zap.Int64("currentGeneration", k8sWorkbench.CurrentGeneration),
			zap.Int64("observedGeneration", k8sWorkbench.ObservedGeneration),
			zap.Any("apps", k8sWorkbench.Apps))

		workbenchID, err := model.GetIDFromClusterName(k8sWorkbench.Name)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get workbench ID from cluster name", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Error(err))
			return fmt.Errorf("unable to get workbench ID from cluster name %s: %w", k8sWorkbench.Name, err)
		}

		workspaceID, err := workspace_model.GetIDFromClusterName(k8sWorkbench.Namespace)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get namespace ID from cluster name", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Error(err))
			return fmt.Errorf("unable to get namespace ID from cluster name %s: %w", k8sWorkbench.Namespace, err)
		}

		workbench := &model.Workbench{
			ID:                      workbenchID,
			TenantID:                k8sWorkbench.TenantID,
			WorkspaceID:             workspaceID,
			InitialResolutionWidth:  k8sWorkbench.InitialResolutionWidth,
			InitialResolutionHeight: k8sWorkbench.InitialResolutionHeight,
			ServerPodStatus:         model.WorkbenchServerPodStatus(k8sWorkbench.ServerPodStatus),
			K8sStatus:               model.K8sWorkbenchStatus(k8sWorkbench.Status),
		}

		switch k8sWorkbench.Status {
		case string(k8s.WorkbenchStatusServerStatusRunning):
			workbench.Status = model.WorkbenchActive
		case string(k8s.WorkbenchStatusServerStatusProgressing):
			workbench.Status = model.WorkbenchActive
		case string(k8s.WorkbenchStatusServerStatusFailed):
			workbench.Status = model.WorkbenchDeleted
		}

		// ALWAYS update DB status from K8s (regardless of generation)
		appInstancesToUpdate := make([]*model.AppInstance, 0, len(k8sWorkbench.Apps))
		for _, app := range k8sWorkbench.Apps {
			k8sStatus := model.K8sAppInstanceStatus(app.K8sStatus)
			appInstance := &model.AppInstance{
				ID:        app.ID,
				Status:    k8sStatus.ToAppInstanceStatus(),
				K8sStatus: k8sStatus,
			}
			appInstancesToUpdate = append(appInstancesToUpdate, appInstance)
		}

		logger.TechLog.Debug(ctx, "updating workbench", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("workbench", workbench))
		_, err = s.store.UpdateWorkbench(ctx, k8sWorkbench.TenantID, workbench)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update workbench", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("workbench", workbench), zap.Error(err))
			return err
		}

		logger.TechLog.Debug(ctx, "updating app instances", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("appInstances", appInstancesToUpdate))

		err = s.store.UpdateAppInstances(ctx, k8sWorkbench.TenantID, appInstancesToUpdate)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update app instances", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("apps", k8sWorkbench.Apps), zap.Error(err))
			return err
		}

		// Check if operator has reconciled before modifying spec
		isReconciled := k8sWorkbench.ObservedGeneration == k8sWorkbench.CurrentGeneration
		if !isReconciled {
			logger.TechLog.Debug(ctx, "skipping spec updates - operator has not reconciled",
				zap.String("namespace", k8sWorkbench.Namespace),
				zap.String("workbenchName", k8sWorkbench.Name),
				zap.Int64("currentGeneration", k8sWorkbench.CurrentGeneration),
				zap.Int64("observedGeneration", k8sWorkbench.ObservedGeneration))
			return nil
		}

		// Process apps that need spec updates based on their status
		for _, app := range k8sWorkbench.Apps {
			k8sStatus := model.K8sAppInstanceStatus(app.K8sStatus)

			switch k8sStatus {
			case model.K8sAppInstanceStatusComplete:
				// Update spec state to Stopped when app completes
				logger.TechLog.Info(ctx, "app instance completed, updating state to Stopped",
					zap.Uint64("appInstanceID", app.ID),
					zap.String("namespace", k8sWorkbench.Namespace),
					zap.String("workbenchName", k8sWorkbench.Name))
				err = s.updateAppInstanceStateFromWatcher(ctx, k8sWorkbench.TenantID, app.ID, model.K8sAppInstanceStateStopped)
				if err != nil {
					logger.TechLog.Error(ctx, "unable to update app instance state to Stopped",
						zap.Uint64("appInstanceID", app.ID),
						zap.Error(err))
					return err
				}

			case model.K8sAppInstanceStatusFailed:
				// Log warning, keep state as Running (timeout logic to be added later)
				logger.TechLog.Warn(ctx, "app instance failed but keeping desired state Running",
					zap.Uint64("appInstanceID", app.ID),
					zap.String("namespace", k8sWorkbench.Namespace),
					zap.String("workbenchName", k8sWorkbench.Name),
					zap.String("k8sStatus", string(k8sStatus)))
			}
		}

		return nil
	}

	s.client.RegisterOnNewWorkbenchHandler(watcher)
	s.client.RegisterOnUpdateWorkbenchHandler(watcher)
}

// updateAppInstanceStateFromWatcher updates an app instance's K8s state in both DB and K8s spec.
// This is called from the watcher when the operator reports a status change that requires a spec update.
func (s *WorkbenchService) updateAppInstanceStateFromWatcher(ctx context.Context, tenantID, appInstanceID uint64, newState model.K8sAppInstanceState) error {
	appInstance, err := s.store.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		return fmt.Errorf("unable to get appInstance %v: %w", appInstanceID, err)
	}

	// Update state in DB
	appInstance.K8sState = newState
	_, err = s.store.UpdateAppInstance(ctx, tenantID, appInstance)
	if err != nil {
		return fmt.Errorf("unable to update appInstance state %v: %w", appInstanceID, err)
	}

	// Update K8s spec
	wsName := s.getWorkspaceName(appInstance.WorkspaceID)
	wbName := s.getWorkbenchName(appInstance.WorkbenchID)

	clientApp, err := s.getK8sAppInstance(ctx, appInstance)
	if err != nil {
		return fmt.Errorf("unable to get k8s app instance %v: %w", appInstance.AppID, err)
	}

	return s.client.UpdateAppInstance(wsName, wbName, clientApp)
}

func (s *WorkbenchService) updateAllWorkbenches(ctx context.Context) {
	workbenches, err := s.store.ListAllWorkbenches(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to query workbenches", zap.Error(err))
		return
	}

	for _, workbench := range workbenches {
		go func(workbench *model.Workbench) {
			logger.TechLog.Debug(ctx, "syncing workbench", zap.Uint64("workbenchID", workbench.ID), zap.String("status", string(workbench.Status)), zap.Any("workbench", workbench))
			err := s.syncWorkbench(ctx, workbench)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to sync workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			}
		}(workbench)
	}
}

func (s *WorkbenchService) cleanIdleWorkbenches(ctx context.Context) {
	workbenches, err := s.store.DeleteIdleWorkbenches(ctx, *s.cfg.Services.WorkbenchService.WorkbenchIdleTimeout)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to query idle workbenches", zap.Error(err))
		return
	}

	for _, workbench := range workbenches {
		go func(workbench *model.Workbench) {
			logger.TechLog.Debug(ctx, "cleaning idle workbench", zap.Uint64("workbenchID", workbench.ID), zap.String("status", string(workbench.Status)), zap.Any("workbench", workbench))
			// err := s.syncWorkbench(ctx, workbench)
			// if err != nil {
			// 	logger.TechLog.Error(ctx, "unable to clean idle workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			// }
			err = s.client.DeleteWorkbench(workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbench.ID))
			if err != nil {
				logger.TechLog.Error(ctx, "unable to delete idle workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			}
		}(workbench)
	}
}

func (s *WorkbenchService) syncWorkbenchWithID(ctx context.Context, tenantID, workbenchID uint64) error {
	workbench, err := s.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to get workbench %v: %w", workbenchID, err)
	}

	err = s.syncWorkbench(ctx, workbench)
	if err != nil {
		return fmt.Errorf("unable to sync workbench %v: %w", workbenchID, err)
	}
	return nil
}

func (s *WorkbenchService) syncWorkbench(ctx context.Context, workbench *model.Workbench) error {
	switch workbench.Status {
	case model.WorkbenchActive:
		apps, err := s.store.ListWorkbenchAppInstances(ctx, workbench.ID)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to list app instances", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			return err
		}

		clientApps := []k8s.AppInstance{}
		for _, app := range apps {
			clientApps = append(clientApps, app.ToK8sAppInstance())
		}

		user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: workbench.TenantID, ID: workbench.UserID})
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get user", zap.Error(err), zap.Uint64("userID", workbench.UserID))
			return err
		}

		username := ""
		if user.Source == auth_helper.GetMainSourceID(s.cfg) {
			username = user.Username
		}

		namespace, workbenchName := workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbench.ID)

		err = s.client.UpdateWorkbench(k8s.Workbench{
			TenantID:                workbench.TenantID,
			Namespace:               namespace,
			Username:                username,
			UserID:                  user.ID,
			Name:                    workbenchName,
			Apps:                    clientApps,
			InitialResolutionWidth:  workbench.InitialResolutionWidth,
			InitialResolutionHeight: workbench.InitialResolutionHeight,
		})
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			return err
		}

		return nil
	case model.WorkbenchDeleted:
		err := s.client.DeleteWorkbench(workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbench.ID))
		if err != nil {
			logger.TechLog.Error(ctx, "unable to delete workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			return err
		}

		logger.TechLog.Debug(ctx, "deleted workbench", zap.Uint64("workbenchID", workbench.ID))
		return nil
	}

	logger.TechLog.Debug(ctx, "skipping workbench update", zap.Uint64("workbenchID", workbench.ID), zap.String("status", string(workbench.Status)))
	return nil
}

func (s *WorkbenchService) ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkbenchFilter) ([]*model.Workbench, *common_model.PaginationResult, error) {
	workbenches, paginationRes, err := s.store.ListWorkbenches(ctx, tenantID, pagination, filter.WorkspaceIDsIn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query workbenches: %w", err)
	}
	return workbenches, paginationRes, nil
}

func (s *WorkbenchService) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workbench %v: %w", workbenchID, err)
	}

	return workbench, nil
}

func (s *WorkbenchService) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to get workbench %v: %w", workbenchID, err)
	}

	err = s.store.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workbenchID, err)
	}

	err = s.client.DeleteWorkbench(workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbenchID))
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workbenchID, err)
	}

	return nil
}

func (s *WorkbenchService) DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error {
	err := s.store.DeleteWorkbenchesInWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to delete workbenches in workspace %v: %w", workspaceID, err)
	}

	return nil
}

func (s *WorkbenchService) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	updatedWorkbench, err := s.store.UpdateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return nil, fmt.Errorf("unable to update workbench %v: %w", workbench.ID, err)
	}

	return updatedWorkbench, nil
}

func (s *WorkbenchService) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	newWorkbench, err := s.store.CreateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return nil, fmt.Errorf("unable to create workbench: %w", err)
	}

	r := authorization_model.NewRole(authorization_model.RoleWorkbenchAdmin,
		authorization_model.WithWorkbench(newWorkbench.ID),
		authorization_model.WithWorkspace(newWorkbench.WorkspaceID))
	err = s.userer.CreateUserRoles(ctx, workbench.TenantID, workbench.UserID, []user_model.UserRole{{Role: r}})
	if err != nil {
		return nil, fmt.Errorf("unable to assign workbench admin role to user %v for workbench %v: %w", workbench.UserID, newWorkbench.ID, err)
	}

	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: workbench.TenantID, ID: workbench.UserID})
	if err != nil {
		return nil, fmt.Errorf("unable to get user %v: %w", workbench.UserID, err)
	}

	username := ""
	if user.Source == auth_helper.GetMainSourceID(s.cfg) {
		username = user.Username
	}

	namespace, workbenchName := workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(newWorkbench.ID)

	err = s.client.CreateWorkbench(k8s.Workbench{
		TenantID:                workbench.TenantID,
		Namespace:               namespace,
		Username:                username,
		UserID:                  user.ID,
		Name:                    workbenchName,
		InitialResolutionWidth:  workbench.InitialResolutionWidth,
		InitialResolutionHeight: workbench.InitialResolutionHeight,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create workbench %v: %w", workbench.ID, err)
	}

	return newWorkbench, nil
}

func (s *WorkbenchService) ManageUserRoleInWorkbench(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", userID, err)
	}

	workbenchID, err := strconv.ParseUint(role.Context["workbench"], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse workbench ID %v: %w", role.Context["workbench"], err)
	}

	workbench, err := s.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to get workbench %v: %w", role.Context["workbench"], err)
	}

	matchingRolesIDs := []uint64{}
	for _, r := range user.Roles {
		if r.Context["workbench"] == role.Context["workbench"] {
			matchingRolesIDs = append(matchingRolesIDs, r.ID)
		}
	}

	if len(matchingRolesIDs) != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workbench roles for user %v for workbench %v: %w", userID, tenantID, err)
		}
	}

	role.Context["workspace"] = fmt.Sprintf("%d", workbench.WorkspaceID)

	logger.TechLog.Debug(ctx, "assigning role to user", zap.Uint64("userID", userID), zap.Any("role", role))

	err = s.userer.CreateUserRoles(ctx, tenantID, userID, []user_model.UserRole{role})
	if err != nil {
		return fmt.Errorf("unable to assign workbench admin role to user %v for workbench %v: %w", userID, tenantID, err)
	}

	err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  fmt.Sprintf("You have been assigned the role %v in workbench %v", role.Role, workbench.Name),
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: &notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return fmt.Errorf("unable to create notification for user %v about new role %v in workspace %v: %w", userID, role.Role, role.Context["workspace"], err)
	}

	return nil
}

func (s *WorkbenchService) RemoveUserFromWorkbench(ctx context.Context, tenantID, userID uint64, workbenchID uint64) error {
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", userID, err)
	}

	matchingRolesIDs := []uint64{}
	for _, r := range user.Roles {
		if r.Context["workbench"] == fmt.Sprintf("%d", workbenchID) {
			matchingRolesIDs = append(matchingRolesIDs, r.ID)
		}
	}

	if len(matchingRolesIDs) != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workbench roles for user %v for workbench %v: %w", userID, workbenchID, err)
		}
	}

	return nil
}

type retryRT struct {
	rt  http.RoundTripper
	cfg config.Config
}

func (r retryRT) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	for i := 0; i < r.cfg.Services.WorkbenchService.RoundTripper.MaxTransientRetry; i++ {
		resp, err := r.rt.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		// retry on common transient network errors
		if ne, ok := err.(net.Error); ok && ne.Temporary() {
			lastErr = err
			logger.TechLog.Warn(context.Background(), "transient network error, retrying", zap.Error(err), zap.Int("attempt", i+1))
			continue
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "connection reset by peer") || strings.Contains(msg, "broken pipe") || strings.Contains(msg, "unexpected eof") || strings.Contains(msg, "read: connection timed out") {
			lastErr = err
			logger.TechLog.Warn(context.Background(), "transient network error, retrying", zap.Error(err), zap.Int("attempt", i+1))
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func (s *WorkbenchService) getRoundtripper() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   s.cfg.Services.WorkbenchService.RoundTripper.DialTimeout,
			KeepAlive: s.cfg.Services.WorkbenchService.RoundTripper.DialKeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     s.cfg.Services.WorkbenchService.RoundTripper.ForceAttemptHTTP2,
		MaxIdleConns:          s.cfg.Services.WorkbenchService.RoundTripper.MaxIdleConns,
		MaxIdleConnsPerHost:   s.cfg.Services.WorkbenchService.RoundTripper.MaxIdleConnsPerHost,
		IdleConnTimeout:       s.cfg.Services.WorkbenchService.RoundTripper.IdleConnTimeout,
		TLSHandshakeTimeout:   s.cfg.Services.WorkbenchService.RoundTripper.TLSHandshakeTimeout,
		ResponseHeaderTimeout: s.cfg.Services.WorkbenchService.RoundTripper.ResponseHeaderTimeout,
	}
}

func (s *WorkbenchService) getProxy(proxyID proxyID) (*proxy, error) {
	// TODO error handling, port forwarding re-creation, cache eviction, cleaning on cache evit and sig stop
	s.proxyRWMutex.RLock()
	if proxy, exists := s.proxyCache[proxyID]; exists {
		s.proxyRWMutex.RUnlock()
		return proxy, nil
	}
	s.proxyRWMutex.RUnlock()

	s.proxyRWMutex.Lock()
	defer s.proxyRWMutex.Unlock()

	var xpraUrl string
	var port uint16
	var stopChan chan struct{}
	var err error
	if !s.cfg.Services.WorkbenchService.BackendInK8S {
		port, stopChan, err = s.client.CreatePortForward(proxyID.namespace, proxyID.workbench)
		if err != nil {
			return nil, fmt.Errorf("failed to create port forward: %w", err)
		}

		xpraUrl = fmt.Sprintf("http://localhost:%v", port)
	} else {
		xpraUrl = fmt.Sprintf("http://%v.%v:8080", proxyID.workbench, proxyID.namespace)
	}
	logger.TechLog.Debug(context.Background(), "targetUrl", zap.String("xpraUrl", xpraUrl))

	targetURL, err := url.Parse(xpraUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	reg := regexp.MustCompile(`^/api/rest/v1/workbenches/[0-9]+/stream`)

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	tr := s.getRoundtripper()
	reverseProxy.Transport = retryRT{rt: tr, cfg: s.cfg}
	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
		logger.TechLog.Error(context.Background(), "proxy error", zap.Error(e), zap.String("workbench", proxyID.workbench), zap.String("namespace", proxyID.namespace))
		s.proxyRWMutex.Lock()
		delete(s.proxyCache, proxyID)
		s.proxyRWMutex.Unlock()
		http.Error(rw, "Proxy Error: "+e.Error(), http.StatusBadGateway)
	}

	originalDirector := reverseProxy.Director

	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.URL.Path = reg.ReplaceAllString(req.URL.Path, "")
	}

	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		middleware.SetCORSHeaders(resp.Request, resp.Header, s.cfg)
		return nil
	}

	proxy := &proxy{
		reverseProxy:    reverseProxy,
		forwardPort:     port,
		forwardStopChan: stopChan,
	}

	s.proxyCache[proxyID] = proxy

	return proxy, nil
}

func (s *WorkbenchService) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to get workbench %v: %w", workbenchID, err)
	}

	namespace, workbenchName := workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbenchID)

	proxyID := proxyID{
		namespace: namespace,
		workbench: workbenchName,
	}

	proxy, err := s.getProxy(proxyID)
	if err != nil {
		return fmt.Errorf("unable to get proxy %v: %w", proxyID, err)
	}

	go s.addWorkbenchHit(workbenchID)

	proxy.reverseProxy.ServeHTTP(w, r)

	return nil
}

func (s *WorkbenchService) addWorkbenchHit(workbenchID uint64) {
	workbenchProxyRequest.WithLabelValues(fmt.Sprintf("workbench%v", workbenchID)).Inc()

	s.proxyHitMutex.Lock()
	defer s.proxyHitMutex.Unlock()

	if _, ok := s.proxyHitCountMap[workbenchID]; !ok {
		s.proxyHitCountMap[workbenchID] = 0
	}
	s.proxyHitCountMap[workbenchID]++
	s.proxyHitDateMap[workbenchID] = time.Now()
}

func (s *WorkbenchService) saveBatchProxyHit(ctx context.Context) {
	s.proxyHitMutex.Lock()
	countMapToSave := s.proxyHitCountMap
	dateMapToSave := s.proxyHitDateMap
	s.proxyHitCountMap = make(map[uint64]uint64)
	s.proxyHitDateMap = make(map[uint64]time.Time)
	s.proxyHitMutex.Unlock()

	err := s.store.SaveBatchProxyHit(ctx, countMapToSave, dateMapToSave)
	if err != nil {
		hits := uint64(0)
		numWorkbenches := len(countMapToSave)
		for _, count := range countMapToSave {
			hits += count
		}
		logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to save batch proxy hit, losing %v hits to %v workbenches", hits, numWorkbenches), zap.Error(err))
	}
}
