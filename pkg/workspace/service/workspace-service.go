package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type Workspaceer interface {
	GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error)
	CreateWorkspace(ctx context.Context, workspace *model.Workspace) (uint64, error)
	UpdateWorkspace(ctx context.Context, workspace *model.Workspace) error
	DeleteWorkspace(ctx context.Context, tenantId, workspaceId uint64) error
}

type WorkspaceStore interface {
	GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error)
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination, allowDeleted bool) ([]*model.Workspace, error)
	CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (uint64, error)
	UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) error
	DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type WorkspaceService struct {
	store  WorkspaceStore
	client k8s.K8sClienter
}

func NewWorkspaceService(store WorkspaceStore, client k8s.K8sClienter) *WorkspaceService {
	ws := &WorkspaceService{
		store:  store,
		client: client,
	}

	go func() {
		if err := ws.updateAllWorkspaces(context.Background()); err != nil {
			logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to update workspaces: %v", err))
		}
	}()

	return ws
}

func (s *WorkspaceService) updateAllWorkspaces(ctx context.Context) error {
	workspaces, err := s.store.ListWorkspaces(ctx, 0, common_model.Pagination{}, true)
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
				if err := s.client.CreateWorkspace(model.GetWorkspaceClusterName(workspace.ID)); err != nil {
					logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to create workspace %v: %v", workspace.ID, err))
				}
			}()
		}
	}

	return nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error) {
	workspaces, err := s.store.ListWorkspaces(ctx, tenantID, pagination, false)
	if err != nil {
		return nil, fmt.Errorf("unable to query workspaces: %w", err)
	}
	return workspaces, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	workspace, err := s.store.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace %v: %w", workspaceID, err)
	}

	return workspace, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	err := s.store.DeleteWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return fmt.Errorf("unable to delete workspace %v: %w", workspaceID, err)
	}

	err = s.client.DeleteWorkspace(model.GetWorkspaceClusterName(workspaceID))
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workspaceID, err)
	}

	return nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) error {
	if err := s.store.UpdateWorkspace(ctx, workspace.TenantID, workspace); err != nil {
		return fmt.Errorf("unable to update workspace %v: %w", workspace.ID, err)
	}

	return nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (uint64, error) {
	id, err := s.store.CreateWorkspace(ctx, workspace.TenantID, workspace)
	if err != nil {
		return 0, fmt.Errorf("unable to create workspace %v: %w", workspace.ID, err)
	}

	err = s.client.CreateWorkspace(model.GetWorkspaceClusterName(id))
	if err != nil {
		return 0, fmt.Errorf("unable to create workbench %v: %w", workspace.ID, err)
	}

	return id, nil
}
