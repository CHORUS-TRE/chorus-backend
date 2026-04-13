package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"go.uber.org/zap"
)

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *notification_model.Notification, userIDs []uint64) error
}

type Workspaceer interface {
	GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error)
	CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, tenantId, workspaceId uint64) error

	ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error
	RemoveUserRoleInWorkspace(ctx context.Context, tenantID, userID, workspaceID uint64, roleName authorization_model.RoleName) error
	RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error

	GetWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) (*model.WorkspaceServiceInstance, error)
	ListWorkspaceServiceInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkspaceServiceInstanceFilter) ([]*model.WorkspaceServiceInstance, *common_model.PaginationResult, error)
	CreateWorkspaceServiceInstance(ctx context.Context, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error)
	UpdateWorkspaceServiceInstance(ctx context.Context, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error)
	DeleteWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) error
}

type WorkspaceServiceInstanceFilter struct {
	WorkspaceIDsIn *[]uint64
}

type Workbencher interface {
	DeleteWorkbenchesInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type WorkspaceStore interface {
	GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, IDIn *[]uint64, allowDeleted bool) ([]*model.Workspace, *common_model.PaginationResult, error)
	DeleteOldWorkspaces(ctx context.Context, duration time.Duration) ([]*model.Workspace, error)
	CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
	UpdateWorkspaceStatus(ctx context.Context, tenantID uint64, workspaceID uint64, networkPolicyStatus, networkPolicyMessage string) error

	GetWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) (*model.WorkspaceServiceInstance, error)
	ListWorkspaceServiceInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workspaceIDsIn *[]uint64) ([]*model.WorkspaceServiceInstance, *common_model.PaginationResult, error)
	ListWorkspaceServiceInstancesByWorkspace(ctx context.Context, workspaceID uint64) ([]*model.WorkspaceServiceInstance, error)
	CreateWorkspaceServiceInstance(ctx context.Context, tenantID uint64, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error)
	UpdateWorkspaceServiceInstance(ctx context.Context, tenantID uint64, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error)
	DeleteWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) error
	UpdateWorkspaceServiceInstanceStatuses(ctx context.Context, workspaceID uint64, statuses map[string]model.WorkspaceServiceInstanceStatusUpdate) error
}

type Userer interface {
	CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []user_model.UserRole) error
	RemoveUserRoles(ctx context.Context, tenantID, userID uint64, roleIDs []uint64) error
	RemoveRolesByContext(ctx context.Context, contextDimension, contextValue string) ([]uint64, error)
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
}

type WorkspaceService struct {
	cfg               config.Config
	store             WorkspaceStore
	k8sClient         k8s.K8sClienter
	workbencher       Workbencher
	userer            Userer
	notificationStore NotificationStore
	auditWriter       audit_service.AuditWriter
}

func NewWorkspaceService(cfg config.Config, store WorkspaceStore, client k8s.K8sClienter, workbencher Workbencher, userer Userer, notificationStore NotificationStore, auditWriter audit_service.AuditWriter) *WorkspaceService {
	ws := &WorkspaceService{
		cfg:               cfg,
		store:             store,
		k8sClient:         client,
		workbencher:       workbencher,
		userer:            userer,
		notificationStore: notificationStore,
		auditWriter:       auditWriter,
	}

	ws.SetClientWatchers()

	go func() {
		if err := ws.updateAllWorkspaces(context.Background()); err != nil {
			logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to update workspaces: %v", err))
		}
	}()

	if ws.cfg.Services.WorkspaceService.EnableKillFixedTimeout {
		logger.TechLog.Info(context.Background(), "starting workspace idle cleaner", zap.Duration("idleTimeout", ws.cfg.Services.WorkspaceService.KillFixedTimeout), zap.Duration("checkInterval", ws.cfg.Services.WorkspaceService.KillFixedCheckInterval))

		go func() {
			interval := ws.cfg.Services.WorkspaceService.KillFixedTimeout
			// sleep a random jitter in initial interval to avoid all instances doing this at the same time
			jitter := time.Duration(rand.Int64N(int64(interval)))
			time.Sleep(jitter)

			for {
				ws.cleanOldWorkspaces(context.Background())
				time.Sleep(interval)
			}
		}()
	}

	return ws
}

func (s *WorkspaceService) updateAllWorkspaces(ctx context.Context) error {
	workspaces, _, err := s.store.ListWorkspaces(ctx, 0, &common_model.Pagination{}, nil, true)
	if err != nil {
		return fmt.Errorf("unable to list workspaces: %w", err)
	}

	for _, workspace := range workspaces {
		if workspace.Status == model.WorkspaceDeleted {
			go func() {
				if err := s.k8sClient.DeleteWorkspace(model.GetWorkspaceClusterName(workspace.ID)); err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to update workbench %v: %v", workspace.ID, err))
				}
			}()
		} else {
			go func() {
				svcs, err := s.store.ListWorkspaceServiceInstancesByWorkspace(context.Background(), workspace.ID)
				if err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to list workspace service instances for workspace %v: %v", workspace.ID, err))
					return
				}
				input := workspaceToK8sInput(workspace, svcs)
				if err := s.k8sClient.CreateWorkspace(input); err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to create workspace %v: %v", workspace.ID, err))
				}
			}()
		}
	}

	return nil
}

func (s *WorkspaceService) cleanOldWorkspaces(ctx context.Context) {
	workspaces, err := s.store.DeleteOldWorkspaces(ctx, s.cfg.Services.WorkspaceService.KillFixedTimeout)
	if err != nil {
		logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to list workspaces: %v", err))
		return
	}

	for _, workspace := range workspaces {
		audit.Record(ctx, s.auditWriter, audit_model.AuditActionWorkspaceDelete,
			audit.WithTenantID(workspace.TenantID),
			audit.WithActorUsername("system"),
			audit.WithWorkspaceID(workspace.ID),
			audit.WithDescription(fmt.Sprintf("Workspace '%s' (ID %d) auto-deleted due to fixed timeout.", workspace.Name, workspace.ID)),
			audit.WithDetail("trigger", "auto_cleanup_timeout"),
		)
	}

	for _, workspace := range workspaces {
		go func(workspaceID uint64) {
			if err := s.k8sClient.DeleteWorkspace(model.GetWorkspaceClusterName(workspaceID)); err != nil {
				logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to delete workspace %v: %v", workspaceID, err))
			}
		}(workspace.ID)
	}
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error) {
	workspaces, paginationRes, err := s.store.ListWorkspaces(ctx, tenantID, pagination, filter.WorkspaceIDsIn, false)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to query workspaces: %w", err)
	}
	return workspaces, paginationRes, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	workspace, err := s.store.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace %v: %w", workspaceID, err)
	}

	return workspace, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	err := s.workbencher.DeleteWorkbenchesInWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to delete workbenches in workspace %v: %w", workspaceID, err)
	}

	_, err = s.userer.RemoveRolesByContext(ctx, "workspace", fmt.Sprintf("%d", workspaceID))
	if err != nil {
		return fmt.Errorf("unable to remove roles for workspace %v: %w", workspaceID, err)
	}

	err = s.store.DeleteWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to delete workspace %v: %w", workspaceID, err)
	}

	err = s.k8sClient.DeleteWorkspace(model.GetWorkspaceClusterName(workspaceID))
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workspaceID, err)
	}

	return nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	updatedWorkspace, err := s.store.UpdateWorkspace(ctx, workspace.TenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("unable to update workspace %v: %w", workspace.ID, err)
	}

	svcs, err := s.store.ListWorkspaceServiceInstancesByWorkspace(ctx, updatedWorkspace.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to list workspace service instances for workspace %v: %w", updatedWorkspace.ID, err)
	}

	input := workspaceToK8sInput(updatedWorkspace, svcs)
	if err := s.k8sClient.UpdateWorkspace(input); err != nil {
		return nil, fmt.Errorf("unable to sync workspace %v to K8s: %w", workspace.ID, err)
	}

	return updatedWorkspace, nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	newWorkspace, err := s.store.CreateWorkspace(ctx, workspace.TenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace %v: %w", workspace.ID, err)
	}

	var rolesToAssign []user_model.UserRole
	if s.cfg.Services.WorkspaceService.CreatorIsAdmin {
		r := authorization_model.NewRole(authorization_model.RoleWorkspaceAdmin, authorization_model.WithWorkspace(newWorkspace.ID))
		rolesToAssign = append(rolesToAssign, user_model.UserRole{Role: r})
	}
	if s.cfg.Services.WorkspaceService.CreatorIsDataManager {
		r := authorization_model.NewRole(authorization_model.RoleWorkspaceDataManager, authorization_model.WithWorkspace(newWorkspace.ID))
		rolesToAssign = append(rolesToAssign, user_model.UserRole{Role: r})
	}

	if len(rolesToAssign) > 0 {
		err = s.userer.CreateUserRoles(ctx, workspace.TenantID, workspace.UserID, rolesToAssign)
		if err != nil {
			return nil, fmt.Errorf("unable to assign workspace roles to user %v for workspace %v: %w", workspace.UserID, newWorkspace.ID, err)
		}
	}

	input := workspaceToK8sInput(newWorkspace, nil)
	err = s.k8sClient.CreateWorkspace(input)
	if err != nil {
		return nil, fmt.Errorf("unable to create workbench %v: %w", workspace.ID, err)
	}

	return newWorkspace, nil
}

func (s *WorkspaceService) ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", userID, err)
	}

	matchingRolesIDs := []uint64{}
	for _, r := range user.Roles {
		if r.Context["workspace"] == role.Context["workspace"] {
			matchingRolesIDs = append(matchingRolesIDs, r.ID)
		}
	}

	if len(matchingRolesIDs) != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workspace roles for user %v for workspace %v: %w", userID, role.Context["workspace"], err)
		}
	}

	err = s.userer.CreateUserRoles(ctx, tenantID, userID, []user_model.UserRole{role})
	if err != nil {
		return fmt.Errorf("unable to assign workspace admin role to user %v for workspace %v: %w", userID, tenantID, err)
	}

	err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  fmt.Sprintf("You have been assigned the role %v in workspace %v", role.Role, role.Context["workspace"]),
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

func (s *WorkspaceService) RemoveUserRoleInWorkspace(ctx context.Context, tenantID, userID, workspaceID uint64, roleName authorization_model.RoleName) error {
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", userID, err)
	}

	matchingRolesIDs := []uint64{}
	for _, r := range user.Roles {
		if r.Context["workspace"] == fmt.Sprintf("%d", workspaceID) && r.Role.Name == roleName {
			matchingRolesIDs = append(matchingRolesIDs, r.ID)
		}
	}

	if len(matchingRolesIDs) == 0 {
		return fmt.Errorf("user %v does not have role %v in workspace %v", userID, roleName, workspaceID)
	}

	err = s.userer.RemoveUserRoles(ctx, tenantID, userID, matchingRolesIDs)
	if err != nil {
		return fmt.Errorf("unable to remove role %v from user %v in workspace %v: %w", roleName, userID, workspaceID, err)
	}

	err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Message:  fmt.Sprintf("You have been removed the role %v in workspace %v", roleName, workspaceID),
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: &notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return fmt.Errorf("unable to create notification for user %v about removed role %v in workspace %v: %w", userID, roleName, workspaceID, err)
	}

	return nil
}

func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error {
	user, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: tenantID, ID: userID})
	if err != nil {
		return fmt.Errorf("unable to get user %v: %w", userID, err)
	}

	matchingRolesIDs := []uint64{}
	for _, r := range user.Roles {
		if r.Context["workspace"] == fmt.Sprintf("%d", workspaceID) {
			matchingRolesIDs = append(matchingRolesIDs, r.ID)
		}
	}

	if len(matchingRolesIDs) != 0 {
		err = s.userer.RemoveUserRoles(ctx, tenantID, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workspace roles for user %v for workspace %v: %w", userID, workspaceID, err)
		}
	}

	return nil
}

func (s *WorkspaceService) GetWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) (*model.WorkspaceServiceInstance, error) {
	svc, err := s.store.GetWorkspaceServiceInstance(ctx, tenantID, workspaceServiceInstanceID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace service instance %v: %w", workspaceServiceInstanceID, err)
	}
	return svc, nil
}

func (s *WorkspaceService) ListWorkspaceServiceInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkspaceServiceInstanceFilter) ([]*model.WorkspaceServiceInstance, *common_model.PaginationResult, error) {
	svcs, paginationRes, err := s.store.ListWorkspaceServiceInstances(ctx, tenantID, pagination, filter.WorkspaceIDsIn)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list workspace service instances: %w", err)
	}
	return svcs, paginationRes, nil
}

func (s *WorkspaceService) CreateWorkspaceServiceInstance(ctx context.Context, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error) {
	created, err := s.store.CreateWorkspaceServiceInstance(ctx, svc.TenantID, svc)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace service instance: %w", err)
	}

	if err := s.syncWorkspaceToK8s(ctx, created.WorkspaceID, created.TenantID); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *WorkspaceService) UpdateWorkspaceServiceInstance(ctx context.Context, svc *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error) {
	updated, err := s.store.UpdateWorkspaceServiceInstance(ctx, svc.TenantID, svc)
	if err != nil {
		return nil, fmt.Errorf("unable to update workspace service instance %v: %w", svc.ID, err)
	}

	if err := s.syncWorkspaceToK8s(ctx, updated.WorkspaceID, updated.TenantID); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *WorkspaceService) DeleteWorkspaceServiceInstance(ctx context.Context, tenantID, workspaceServiceInstanceID uint64) error {
	svc, err := s.store.GetWorkspaceServiceInstance(ctx, tenantID, workspaceServiceInstanceID)
	if err != nil {
		return fmt.Errorf("unable to get workspace service instance %v: %w", workspaceServiceInstanceID, err)
	}

	err = s.store.DeleteWorkspaceServiceInstance(ctx, tenantID, workspaceServiceInstanceID)
	if err != nil {
		return fmt.Errorf("unable to delete workspace service instance %v: %w", workspaceServiceInstanceID, err)
	}

	if err := s.syncWorkspaceToK8s(ctx, svc.WorkspaceID, svc.TenantID); err != nil {
		return err
	}

	return nil
}

func (s *WorkspaceService) syncWorkspaceToK8s(ctx context.Context, workspaceID, tenantID uint64) error {
	workspace, err := s.store.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to get workspace %v for K8s sync: %w", workspaceID, err)
	}

	svcs, err := s.store.ListWorkspaceServiceInstancesByWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to list workspace service instances for workspace %v: %w", workspaceID, err)
	}

	input := workspaceToK8sInput(workspace, svcs)
	if err := s.k8sClient.UpdateWorkspace(input); err != nil {
		return fmt.Errorf("unable to sync workspace %v to K8s: %w", workspaceID, err)
	}

	return nil
}

// SetClientWatchers registers a handler for Workspace CRD status updates from K8s.
func (s *WorkspaceService) SetClientWatchers() {
	watcher := func(wsOutput k8s.WorkspaceOutput) error {
		ctx := context.Background()
		logger.TechLog.Debug(ctx, "workspace watcher received update",
			zap.String("namespace", wsOutput.Namespace),
			zap.Int64("currentGeneration", wsOutput.CurrentGeneration),
			zap.Int64("observedGeneration", wsOutput.ObservedGeneration),
		)

		// Skip updates if operator has not reconciled yet
		if wsOutput.ObservedGeneration != wsOutput.CurrentGeneration {
			logger.TechLog.Debug(ctx, "skipping workspace update - operator has not reconciled",
				zap.String("namespace", wsOutput.Namespace),
				zap.Int64("currentGeneration", wsOutput.CurrentGeneration),
				zap.Int64("observedGeneration", wsOutput.ObservedGeneration),
			)
			return nil
		}

		workspaceID, err := model.GetIDFromClusterName(wsOutput.Namespace)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to get workspace ID from namespace", zap.String("namespace", wsOutput.Namespace), zap.Error(err))
			return fmt.Errorf("unable to get workspace ID from namespace %s: %w", wsOutput.Namespace, err)
		}

		serviceStatuses := map[string]model.WorkspaceServiceInstanceStatusUpdate{}
		for name, ss := range wsOutput.ServiceStatuses {
			serviceStatuses[name] = model.WorkspaceServiceInstanceStatusUpdate{
				Status:         ss.Status,
				StatusMessage:  ss.Message,
				ConnectionInfo: ss.ConnectionInfo,
				SecretName:     ss.SecretName,
			}
		}

		err = s.store.UpdateWorkspaceStatus(ctx, wsOutput.TenantID, workspaceID, wsOutput.NetworkPolicyStatus, wsOutput.NetworkPolicyMessage)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update workspace status from watcher",
				zap.Uint64("workspaceID", workspaceID), zap.Error(err))
			return fmt.Errorf("unable to update workspace status %v: %w", workspaceID, err)
		}

		if len(serviceStatuses) > 0 {
			err = s.store.UpdateWorkspaceServiceInstanceStatuses(ctx, workspaceID, serviceStatuses)
			if err != nil {
				logger.TechLog.Error(ctx, "unable to update workspace service instance statuses from watcher",
					zap.Uint64("workspaceID", workspaceID), zap.Error(err))
				return fmt.Errorf("unable to update workspace service instance statuses %v: %w", workspaceID, err)
			}
		}

		return nil
	}

	s.k8sClient.RegisterOnUpdateWorkspaceHandler(watcher)
}

// workspaceToK8sInput converts a workspace model and its services to a K8s WorkspaceInput.
func workspaceToK8sInput(ws *model.Workspace, svcs []*model.WorkspaceServiceInstance) k8s.WorkspaceInput {
	services := make(map[string]k8s.WorkspaceInputService, len(svcs))
	for _, svc := range svcs {
		var creds *k8s.WorkspaceServiceCredentials
		if svc.CredentialsSecretName != "" {
			creds = &k8s.WorkspaceServiceCredentials{
				SecretName: svc.CredentialsSecretName,
				Paths:      []string(svc.CredentialsPaths),
			}
		}
		services[svc.Name] = k8s.WorkspaceInputService{
			Chart: k8s.WorkspaceServiceChart{
				Registry:   svc.ChartRegistry,
				Repository: svc.ChartRepository,
				Tag:        svc.ChartTag,
			},
			Values:                 svc.Values,
			Credentials:            creds,
			ConnectionInfoTemplate: svc.ConnectionInfoTemplate,
			ComputedValues:         map[string]string(svc.ComputedValues),
		}
	}

	return k8s.WorkspaceInput{
		TenantID:      ws.TenantID,
		Namespace:     model.GetWorkspaceClusterName(ws.ID),
		NetworkPolicy: string(ws.NetworkPolicy),
		AllowedFQDNs:  []string(ws.AllowedFQDNs),
		Clipboard:     string(ws.Clipboard),
		Services:      services,
	}
}
