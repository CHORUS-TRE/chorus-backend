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

	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/protocol/rest/middleware"
	app_service "github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	auth_helper "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/helper"
	authentication_service "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
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

var streamPathRegex = regexp.MustCompile(`^/api/rest/v1/workbenches/[0-9]+/stream`)

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

type WorkspaceReader interface {
	GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*workspace_model.Workspace, error)
}

type Workbencher interface {
	GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkbenchFilter) ([]*model.Workbench, *common_model.PaginationResult, error)
	CreateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error)
	ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error
	UpdateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error)
	DeleteWorkbench(ctx context.Context, tenantId, workbenchId uint64) (*model.Workbench, error)
	DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error

	AddUserRoleInWorkbench(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error
	RemoveUserFromWorkbench(ctx context.Context, tenantID, userID, workbenchID uint64) error

	GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error)
	ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.AppInstanceFilter) ([]*model.AppInstance, *common_model.PaginationResult, error)
	CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error)
	UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error)
	DeleteAppInstance(ctx context.Context, tenantId, appInstanceId uint64) (*model.AppInstance, error)
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
	workspaceReader   WorkspaceReader
	auditWriter       audit_service.AuditWriter

	proxyRWMutex     sync.RWMutex
	proxyCache       map[proxyID]*proxy
	proxyHitMutex    sync.Mutex
	proxyHitCountMap map[uint64]uint64
	proxyHitDateMap  map[uint64]time.Time
}

func NewWorkbenchService(cfg config.Config, store WorkbenchStore, client k8s.K8sClienter, apper app_service.Apper, userer user_service.Userer, authenticator authentication_service.Authenticator, notificationStore NotificationStore, workspaceReader WorkspaceReader, auditWriter audit_service.AuditWriter) *WorkbenchService {
	s := &WorkbenchService{
		cfg:    cfg,
		store:  store,
		client: client,

		apper:             apper,
		userer:            userer,
		authenticator:     authenticator,
		notificationStore: notificationStore,
		workspaceReader:   workspaceReader,
		auditWriter:       auditWriter,

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
		logger.TechLog.Debug(ctx, "watcher received a workbench update", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Int64("currentGeneration", k8sWorkbench.CurrentGeneration), zap.Int64("observedGeneration", k8sWorkbench.ObservedGeneration))

		// Skip updates if operator has not reconciled yet
		if k8sWorkbench.ObservedGeneration != k8sWorkbench.CurrentGeneration {
			logger.TechLog.Debug(ctx, "skipping updates - operator has not reconciled", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Int64("currentGeneration", k8sWorkbench.CurrentGeneration), zap.Int64("observedGeneration", k8sWorkbench.ObservedGeneration))
			return nil
		}

		// Get workbench ID and workspace ID from cluster names
		workbenchID, err := model.GetIDFromClusterName(k8sWorkbench.Name)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get workbench ID from cluster name", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Error(err))
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workbench ID from cluster name %s", k8sWorkbench.Name))
		}

		workspaceID, err := workspace_model.GetIDFromClusterName(k8sWorkbench.Namespace)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get namespace ID from cluster name", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Error(err))
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get namespace ID from cluster name %s", k8sWorkbench.Namespace))
		}

		// Fetch existing workbench
		workbench, err := s.store.GetWorkbench(ctx, k8sWorkbench.TenantID, workbenchID)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get workbench for k8s status update", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Error(err))
			return cerr.WrapStoreError(err, fmt.Sprintf("Unable to get workbench %d", workbenchID))
		}

		// Overwrite k8s-derived statuses
		workbench.ServerPodStatus = model.WorkbenchServerPodStatus(k8sWorkbench.ServerPodStatus)
		workbench.ServerPodMessage = model.WorbenchServerPodMessage(k8sWorkbench.ServerPodMessage)
		workbench.K8sStatus = model.K8sWorkbenchStatus(k8sWorkbench.Status)

		switch k8sWorkbench.Status {
		case string(k8s.WorkbenchStatusServerStatusRunning):
			workbench.Status = model.WorkbenchActive
		case string(k8s.WorkbenchStatusServerStatusProgressing):
			workbench.Status = model.WorkbenchActive
		case string(k8s.WorkbenchStatusServerStatusFailed):
			workbench.Status = model.WorkbenchDeleted
		}

		// Update workbench in DB
		logger.TechLog.Debug(ctx, "updating workbench", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("workbench", workbench))
		_, err = s.store.UpdateWorkbench(ctx, k8sWorkbench.TenantID, workbench)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update workbench", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("workbench", workbench), zap.Error(err))
			return err
		}

		// Build app instances to update with correct status and K8sState
		appInstancesToUpdate := make([]*model.AppInstance, 0, len(k8sWorkbench.Apps))
		appInstancesToDelete := make([]k8s.AppInstance, 0)
		appsNeedingK8sUpdate := make([]k8s.AppInstance, 0)

		for _, app := range k8sWorkbench.Apps {
			k8sStatus := model.K8sAppInstanceStatus(app.K8sStatus)

			switch k8sStatus {
			case model.K8sAppInstanceStatusComplete:
				// App completed - update status and set K8sState to Stopped
				logger.TechLog.Info(ctx, "app instance completed", zap.Uint64("appInstanceID", app.ID), zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name))
				appInstancesToUpdate = append(appInstancesToUpdate, &model.AppInstance{
					ID:         app.ID,
					Status:     k8sStatus.ToAppInstanceStatus(),
					K8sStatus:  k8sStatus,
					K8sMessage: model.K8sAppInstanceMessage(app.K8sMessage),
					K8sState:   model.K8sAppInstanceStateStopped,
				})
				// Queue for K8s spec update
				updatedApp := app
				updatedApp.K8sState = string(model.K8sAppInstanceStateStopped)
				appsNeedingK8sUpdate = append(appsNeedingK8sUpdate, updatedApp)

			case model.K8sAppInstanceStatusStopped:
				// App stopped - delete from DB
				logger.TechLog.Info(ctx, "app instance stopped", zap.Uint64("appInstanceID", app.ID), zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name))
				appInstancesToDelete = append(appInstancesToDelete, app)

			case model.K8sAppInstanceStatusFailed:
				// App failed - update status but keep K8sState as Running
				logger.TechLog.Warn(ctx, "app instance failed, keeping desired state Running", zap.Uint64("appInstanceID", app.ID), zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name))
				appInstancesToUpdate = append(appInstancesToUpdate, &model.AppInstance{
					ID:         app.ID,
					Status:     k8sStatus.ToAppInstanceStatus(),
					K8sStatus:  k8sStatus,
					K8sMessage: model.K8sAppInstanceMessage(app.K8sMessage),
					K8sState:   model.K8sAppInstanceStateRunning,
				})

			default:
				// Other statuses (Running, Progressing, Unknown) - update status, keep K8sState as Running
				appInstancesToUpdate = append(appInstancesToUpdate, &model.AppInstance{
					ID:         app.ID,
					Status:     k8sStatus.ToAppInstanceStatus(),
					K8sStatus:  k8sStatus,
					K8sMessage: model.K8sAppInstanceMessage(app.K8sMessage),
					K8sState:   model.K8sAppInstanceStateRunning,
				})
			}
		}

		// Update K8s spec for apps that need state changes
		for _, clientApp := range appsNeedingK8sUpdate {
			err = s.client.UpdateAppInstance(k8sWorkbench.Namespace, k8sWorkbench.Name, clientApp)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to update app instance in K8s", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Uint64("appInstanceID", clientApp.ID), zap.Error(err))
				return err
			}
		}

		// Update app instances in store
		if len(appInstancesToUpdate) > 0 {
			err = s.store.UpdateAppInstances(ctx, k8sWorkbench.TenantID, appInstancesToUpdate)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to update app instances", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Any("appInstanceIDs", appInstancesToUpdate), zap.Error(err))
				return err
			}
		}

		// Delete stopped app instances from store
		for _, ai := range appInstancesToDelete {
			if err := s.store.DeleteAppInstance(ctx, k8sWorkbench.TenantID, ai.ID); err != nil {
				logger.TechLog.Error(ctx, "unable to delete app instance", zap.String("namespace", k8sWorkbench.Namespace), zap.String("workbenchName", k8sWorkbench.Name), zap.Uint64("appInstanceID", ai.ID), zap.Error(err))
				continue
			}

			audit.Record(ctx, s.auditWriter, audit_model.AuditActionAppInstanceDelete,
				audit.WithTenantID(k8sWorkbench.TenantID),
				audit.WithActorID(k8sWorkbench.UserID),
				audit.WithActorUsername(k8sWorkbench.Username),
				audit.WithWorkspaceID(workspaceID),
				audit.WithWorkbenchID(workbenchID),
				audit.WithDescription(fmt.Sprintf("Terminated instance of '%s' (version %s).", ai.AppName, ai.AppTag)),
				audit.WithDetail("app_instance_id", ai.ID),
				audit.WithDetail("app_name", ai.AppName),
				audit.WithDetail("app_image_registry", ai.AppRegistry),
				audit.WithDetail("app_image_name", ai.AppImage),
				audit.WithDetail("app_image_tag", ai.AppTag),
				audit.WithDetail("trigger", "k8s_watcher"),
			)
		}

		return nil
	}

	s.client.RegisterOnNewWorkbenchHandler(watcher)
	s.client.RegisterOnUpdateWorkbenchHandler(watcher)
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

	// Record audit logs for deleted workbenches
	for _, workbench := range workbenches {
		audit.Record(ctx, s.auditWriter, audit_model.AuditActionWorkbenchDelete,
			audit.WithTenantID(workbench.TenantID),
			audit.WithActorUsername("system"),
			audit.WithWorkspaceID(workbench.WorkspaceID),
			audit.WithWorkbenchID(workbench.ID),
			audit.WithDescription(fmt.Sprintf("Workbench with ID %d auto-deleted due to idle timeout.", workbench.ID)),
			audit.WithDetail("trigger", "idle_cleanup"),
		)
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
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workbench %v", workbenchID))
	}

	err = s.syncWorkbench(ctx, workbench)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to sync workbench %v", workbenchID))
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

		clipboard := ""
		ws, wsErr := s.workspaceReader.GetWorkspace(ctx, workbench.TenantID, workbench.WorkspaceID)
		if wsErr != nil {
			logger.TechLog.Warn(ctx, "unable to get workspace for clipboard", zap.Error(wsErr), zap.Uint64("workspaceID", workbench.WorkspaceID))
		} else {
			clipboard = string(ws.Clipboard)
		}

		err = s.client.UpdateWorkbench(k8s.Workbench{
			TenantID:                workbench.TenantID,
			Namespace:               namespace,
			Username:                username,
			UserID:                  user.ID,
			Name:                    workbenchName,
			Apps:                    clientApps,
			Clipboard:               clipboard,
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

func (s *WorkbenchService) ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkbenchFilter) ([]*model.Workbench, *common_model.PaginationResult, error) {
	workbenches, paginationRes, err := s.store.ListWorkbenches(ctx, tenantID, pagination, filter.WorkspaceIDsIn)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, "Unable to list workbenches")
	}
	return workbenches, paginationRes, nil
}

func (s *WorkbenchService) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get workbench %v", workbenchID))
	}

	return workbench, nil
}

func (s *WorkbenchService) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to get workbench %v", workbenchID))
	}

	err = s.store.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to delete workbench %v", workbenchID))
	}

	_, err = s.userer.RemoveRolesByContext(ctx, "workbench", fmt.Sprintf("%d", workbenchID))
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to remove roles for workbench %v", workbenchID))
	}

	err = s.client.DeleteWorkbench(workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbenchID))
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to delete workbench %v in K8s", workbenchID))
	}

	return workbench, nil
}

func (s *WorkbenchService) DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error {
	err := s.store.DeleteWorkbenchesInWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to delete workbenches in workspace %v", workspaceID))
	}

	return nil
}

func (s *WorkbenchService) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	updatedWorkbench, err := s.store.UpdateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return nil, cerr.WrapStoreError(err, fmt.Sprintf("Unable to update workbench %v", workbench.ID))
	}

	return updatedWorkbench, nil
}

func (s *WorkbenchService) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	newWorkbench, err := s.store.CreateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to create workbench")
	}

	r := authorization_model.NewRole(authorization_model.RoleWorkbenchAdmin,
		authorization_model.WithWorkbench(newWorkbench.ID),
		authorization_model.WithWorkspace(newWorkbench.WorkspaceID))
	err = s.userer.CreateUserRoles(ctx, workbench.TenantID, workbench.UserID, []user_model.UserRole{{Role: r}})
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to assign workbench admin role to user %v for workbench %v", workbench.UserID, newWorkbench.ID))
	}

	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: workbench.TenantID, ID: workbench.UserID})
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get user %v", workbench.UserID))
	}

	username := ""
	if user.Source == auth_helper.GetMainSourceID(s.cfg) {
		username = user.Username
	}

	namespace, workbenchName := workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(newWorkbench.ID)

	clipboard := ""
	ws, wsErr := s.workspaceReader.GetWorkspace(ctx, workbench.TenantID, workbench.WorkspaceID)
	if wsErr != nil {
		logger.TechLog.Warn(ctx, "unable to get workspace for clipboard", zap.Error(wsErr), zap.Uint64("workspaceID", workbench.WorkspaceID))
	} else {
		clipboard = string(ws.Clipboard)
	}

	err = s.client.CreateWorkbench(k8s.Workbench{
		TenantID:                workbench.TenantID,
		Namespace:               namespace,
		Username:                username,
		UserID:                  user.ID,
		Name:                    workbenchName,
		Clipboard:               clipboard,
		InitialResolutionWidth:  workbench.InitialResolutionWidth,
		InitialResolutionHeight: workbench.InitialResolutionHeight,
	})
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create workbench %v in K8s", workbench.ID))
	}

	return newWorkbench, nil
}

func (s *WorkbenchService) AddUserRoleInWorkbench(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	// Verify that the workbench exists
	workbenchID, err := strconv.ParseUint(role.Context["workbench"], 10, 64)
	if err != nil {
		return cerr.ErrInvalidRequest.Wrap(err, fmt.Sprintf("Unable to parse workbench ID %v", role.Context["workbench"]))
	}

	workbench, err := s.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workbench %v", role.Context["workbench"]))
	}

	// Verify that the user exists and get its roles
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get user %v", userID))
	}

	// Remove existing role in workbench
	existingRoleID := uint64(0)
	for _, r := range user.Roles {
		if r.Context["workbench"] == role.Context["workbench"] {
			existingRoleID = r.ID
			break
		}
	}

	if existingRoleID != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, []uint64{existingRoleID})
		if err != nil {
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to remove existing workbench roles for user %v for workbench %v", userID, tenantID))
		}
	}

	// Add the new role to the user
	role.Context["workspace"] = fmt.Sprintf("%d", workbench.WorkspaceID)
	logger.TechLog.Debug(ctx, "assigning role to user", zap.Uint64("userID", userID), zap.Any("role", role))

	err = s.userer.CreateUserRoles(ctx, tenantID, userID, []user_model.UserRole{role})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to assign workbench admin role to user %v for workbench %v", userID, tenantID))
	}

	// Notify the user about the new role
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
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create notification for user %v about new role %v in workspace %v", userID, role.Role, role.Context["workspace"]))
	}

	return nil
}

func (s *WorkbenchService) RemoveUserFromWorkbench(ctx context.Context, tenantID, userID uint64, workbenchID uint64) error {
	// Verify that the workbench exists
	_, err := s.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get workbench %v", workbenchID))
	}

	// Verify that the user exists and get its roles
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get user %v", userID))
	}

	// Get the user's role in the workbench
	workbenchRoleID := uint64(0)
	for _, r := range user.Roles {
		if r.Context["workbench"] == fmt.Sprintf("%d", workbenchID) {
			workbenchRoleID = r.ID
			break
		}
	}

	// Remove workbench role from the user
	if workbenchRoleID != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, []uint64{workbenchRoleID})
		if err != nil {
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to remove existing workbench roles for user %v for workbench %v", userID, workbenchID))
		}
	}

	// Notify the user about being removed from the workbench
	err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  fmt.Sprintf("You have been removed from workbench %v", workbenchID),
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: &notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to create notification for user %v about being removed from workbench %v", userID, workbenchID))
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
		var scheme string
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
		proto := req.Proto
		method := req.Method
		remoteAddr := req.RemoteAddr
		userAgent := req.UserAgent()
		uri := strings.Join([]string{scheme, "://", req.Host, req.RequestURI}, "")

		ctx := req.Context()

		logger.TechLog.Debug(ctx, "request started",
			zap.String("http-scheme", scheme),
			zap.String("http-proto", proto),
			zap.String("http-method", method),
			zap.String("remote-addr", remoteAddr),
			zap.String("user-agent", userAgent),
			zap.String("uri", uri),
			zap.Int("attempt", i+1),
		)

		t1 := time.Now()

		resp, err := r.rt.RoundTrip(req)
		if err == nil {
			logger.TechLog.Debug(ctx, "request completed",
				zap.String("http-scheme", scheme),
				zap.String("http-proto", proto),
				zap.String("http-method", method),
				zap.String("remote-addr", remoteAddr),
				zap.String("user-agent", userAgent),
				zap.String("uri", uri),
				zap.Int("attempt", i+1),
				zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(t1).Nanoseconds())/1000000.0),
			)

			return resp, nil
		}

		logger.TechLog.Debug(ctx, "request completed",
			zap.String("http-scheme", scheme),
			zap.String("http-proto", proto),
			zap.String("http-method", method),
			zap.String("remote-addr", remoteAddr),
			zap.String("user-agent", userAgent),
			zap.String("uri", uri),
			zap.Int("attempt", i+1),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(t1).Nanoseconds())/1000000.0),
		)

		// retry on common transient network errors
		if ne, ok := err.(net.Error); ok && ne.Temporary() {
			lastErr = err
			logger.TechLog.Warn(context.Background(), "transient network error, retrying", zap.Error(err), zap.Int("attempt", i+1))
			continue
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "connection reset by peer") || strings.Contains(msg, "broken pipe") || strings.Contains(msg, "unexpected eof") || strings.Contains(msg, "read: connection timed out") || strings.Contains(msg, "connect: operation not permitted") {
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

	logger.TechLog.Debug(context.Background(), "proxy cache miss, acquiring write lock", zap.String("workbench", proxyID.workbench), zap.String("namespace", proxyID.namespace))

	lockStart := time.Now()
	s.proxyRWMutex.Lock()
	defer s.proxyRWMutex.Unlock()

	logger.TechLog.Debug(context.Background(), "write lock acquired", zap.String("workbench", proxyID.workbench), zap.String("namespace", proxyID.namespace), zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(lockStart).Nanoseconds())/1000000.0))

	if p, exists := s.proxyCache[proxyID]; exists {
		logger.TechLog.Debug(context.Background(), "proxy created by another goroutine while waiting for lock", zap.String("workbench", proxyID.workbench), zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(lockStart).Nanoseconds())/1000000.0))
		return p, nil
	}

	var xpraUrl string
	var port uint16
	var stopChan chan struct{}
	var err error
	if !s.cfg.Services.WorkbenchService.BackendInK8S {
		port, stopChan, err = s.client.CreatePortForward(proxyID.namespace, proxyID.workbench)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, "Failed to create port forward")
		}

		xpraUrl = fmt.Sprintf("http://localhost:%v", port)
	} else {
		xpraUrl = fmt.Sprintf("http://%v.%v:8080", proxyID.workbench, proxyID.namespace)
	}
	logger.TechLog.Debug(context.Background(), "targetUrl", zap.String("xpraUrl", xpraUrl))

	targetURL, err := url.Parse(xpraUrl)
	if err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "Failed to parse xpra url")
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.FlushInterval = -1
	tr := s.getRoundtripper()
	reverseProxy.Transport = retryRT{rt: tr, cfg: s.cfg}
	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, e error) {
		logger.TechLog.Error(context.Background(), "proxy error, evicting proxy", zap.Error(e), zap.String("workbench", proxyID.workbench), zap.String("namespace", proxyID.namespace))
		lockStart := time.Now()
		s.proxyRWMutex.Lock()
		delete(s.proxyCache, proxyID)
		logger.TechLog.Info(context.Background(), "proxy evicted and port-forward closed", zap.String("workbench", proxyID.workbench), zap.Int("remainingProxies", len(s.proxyCache)), zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(lockStart).Nanoseconds())/1000000.0))
		s.proxyRWMutex.Unlock()
		http.Error(rw, "Proxy Error: "+e.Error(), http.StatusBadGateway)
	}

	originalDirector := reverseProxy.Director

	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.URL.Path = streamPathRegex.ReplaceAllString(req.URL.Path, "")
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

	logger.TechLog.Info(context.Background(), "new proxy created",
		zap.String("workbench", proxyID.workbench), zap.String("namespace", proxyID.namespace),
		zap.Uint16("forwardPort", port), zap.Int("cacheSize", len(s.proxyCache)))

	return proxy, nil
}

func (s *WorkbenchService) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return cerr.WrapStoreError(err, fmt.Sprintf("Unable to get workbench %v", workbenchID))
	}

	namespace, workbenchName := workspace_model.GetWorkspaceClusterName(workbench.WorkspaceID), model.GetWorkbenchClusterName(workbenchID)

	proxyID := proxyID{
		namespace: namespace,
		workbench: workbenchName,
	}

	proxy, err := s.getProxy(proxyID)
	if err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get proxy %v", proxyID))
	}

	go s.addWorkbenchHit(workbenchID)

	// Audit the initial stream navigation (user opening the workbench in their browser).
	// Sec-Fetch-Mode: navigate distinguishes user-initiated page loads from script-initiated fetches.
	remainingPath := streamPathRegex.ReplaceAllString(r.URL.Path, "")
	if remainingPath == "/" && r.Header.Get("Sec-Fetch-Mode") == "navigate" {
		audit.Record(ctx, s.auditWriter, audit_model.AuditActionWorkbenchStream,
			audit.WithWorkbenchID(workbenchID),
			audit.WithWorkspaceID(workbench.WorkspaceID),
			audit.WithDescription(fmt.Sprintf("Accessed stream of session '%s' (ID %d) in workspace %d.", workbench.Name, workbenchID, workbench.WorkspaceID)),
			audit.WithDetail("workbench_id", workbenchID),
			audit.WithDetail("workbench_name", workbench.Name),
			audit.WithDetail("workspace_id", workbench.WorkspaceID),
		)
	}

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
