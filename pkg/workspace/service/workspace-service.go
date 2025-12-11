package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
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
	RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error
}

type Workbencher interface {
	DeleteWorkbenchsInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type WorkspaceStore interface {
	GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, IDIn *[]uint64, allowDeleted bool) ([]*model.Workspace, *common_model.PaginationResult, error)
	DeleteOldWorkspaces(ctx context.Context, duration time.Duration) ([]*model.Workspace, error)
	CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type Userer interface {
	CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []user_model.UserRole) error
	RemoveUserRoles(ctx context.Context, tenantID, userID uint64, roleIDs []uint64) error
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
}

type WorkspaceService struct {
	cfg               config.Config
	store             WorkspaceStore
	k8sClient         k8s.K8sClienter
	workbencher       Workbencher
	userer            Userer
	notificationStore NotificationStore
}

func NewWorkspaceService(cfg config.Config, store WorkspaceStore, client k8s.K8sClienter, workbencher Workbencher, userer Userer, notificationStore NotificationStore) *WorkspaceService {
	ws := &WorkspaceService{
		cfg:               cfg,
		store:             store,
		k8sClient:         client,
		workbencher:       workbencher,
		userer:            userer,
		notificationStore: notificationStore,
	}

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
				if err := s.k8sClient.CreateWorkspace(workspace.TenantID, model.GetWorkspaceClusterName(workspace.ID)); err != nil {
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
	err := s.workbencher.DeleteWorkbenchsInWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to delete workbenchs in workspace %v: %w", workspaceID, err)
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

	return updatedWorkspace, nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	newWorkspace, err := s.store.CreateWorkspace(ctx, workspace.TenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace %v: %w", workspace.ID, err)
	}

	r := authorization_model.NewRole(authorization_model.RoleWorkspaceAdmin, authorization_model.WithWorkspace(newWorkspace.ID))
	err = s.userer.CreateUserRoles(ctx, workspace.TenantID, workspace.UserID, []user_model.UserRole{{Role: r}})
	if err != nil {
		return nil, fmt.Errorf("unable to assign workspace admin role to user %v for workspace %v: %w", workspace.UserID, newWorkspace.ID, err)
	}

	err = s.k8sClient.CreateWorkspace(workspace.TenantID, model.GetWorkspaceClusterName(newWorkspace.ID))
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
			return fmt.Errorf("unable to remove existing workspace roles for user %v for workspace %v: %w", userID, tenantID, err)
		}
	}

	err = s.userer.CreateUserRoles(ctx, tenantID, userID, []user_model.UserRole{role})
	if err != nil {
		return fmt.Errorf("unable to assign workspace admin role to user %v for workspace %v: %w", userID, tenantID, err)
	}

	err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
		TenantID: tenantID,
		Message:  fmt.Sprintf("You have been assigned the role %v in workspace %v", role.Role, role.Context["workspace"]),
		Content: notification_model.NotificationContent{
			Type: "SystemNotification",
			SystemNotification: notification_model.SystemNotification{
				RefreshJWTRequired: true,
			},
		},
	}, []uint64{userID})
	if err != nil {
		return fmt.Errorf("unable to create notification for user %v about new role %v in workspace %v: %w", userID, role.Role, role.Context["workspace"], err)
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
