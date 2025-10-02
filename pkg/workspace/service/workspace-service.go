package service

import (
	"context"
	"fmt"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type WorkspaceFilter struct {
	WorkspaceIDsIn *[]uint64
}

type Workspaceer interface {
	GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error)
	CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, tenantId, workspaceId uint64) error

	ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error
	RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error

	GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error)
	ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error)
	CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error)
	DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error
}

type Workbencher interface {
	DeleteWorkbenchsInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type WorkspaceStore interface {
	GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, IDIn *[]uint64, allowDeleted bool) ([]*model.Workspace, *common_model.PaginationResult, error)
	CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type Userer interface {
	CreateUserRoles(ctx context.Context, userID uint64, roles []user_model.UserRole) error
	RemoveUserRoles(ctx context.Context, userID uint64, roleIDs []uint64) error
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
}

type WorkspaceService struct {
	store       WorkspaceStore
	client      k8s.K8sClienter
	workbencher Workbencher
	userer      Userer
	minioClient minio.MinioClienter
}

func NewWorkspaceService(store WorkspaceStore, client k8s.K8sClienter, workbencher Workbencher, userer Userer, minioClient minio.MinioClienter) *WorkspaceService {
	ws := &WorkspaceService{
		store:       store,
		client:      client,
		workbencher: workbencher,
		userer:      userer,
		minioClient: minioClient,
	}

	go func() {
		if err := ws.updateAllWorkspaces(context.Background()); err != nil {
			logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to update workspaces: %v", err))
		}
	}()

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
				if err := s.client.DeleteWorkspace(model.GetWorkspaceClusterName(workspace.ID)); err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to update workbench %v: %v", workspace.ID, err))
				}
			}()
		} else {
			go func() {
				if err := s.client.CreateWorkspace(workspace.TenantID, model.GetWorkspaceClusterName(workspace.ID)); err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to create workspace %v: %v", workspace.ID, err))
				}
			}()
		}
	}

	return nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error) {
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

	err = s.client.DeleteWorkspace(model.GetWorkspaceClusterName(workspaceID))
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
	err = s.userer.CreateUserRoles(ctx, workspace.UserID, []user_model.UserRole{{Role: r}})
	if err != nil {
		return nil, fmt.Errorf("unable to assign workspace admin role to user %v for workspace %v: %w", workspace.UserID, newWorkspace.ID, err)
	}

	err = s.client.CreateWorkspace(workspace.TenantID, model.GetWorkspaceClusterName(newWorkspace.ID))
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
		err = s.userer.RemoveUserRoles(ctx, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workspace roles for user %v for workspace %v: %w", userID, tenantID, err)
		}
	}

	err = s.userer.CreateUserRoles(ctx, userID, []user_model.UserRole{role})
	if err != nil {
		return fmt.Errorf("unable to assign workspace admin role to user %v for workspace %v: %w", userID, tenantID, err)
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
		err = s.userer.RemoveUserRoles(ctx, userID, matchingRolesIDs)
		if err != nil {
			return fmt.Errorf("unable to remove existing workspace roles for user %v for workspace %v: %w", userID, workspaceID, err)
		}
	}

	return nil
}
