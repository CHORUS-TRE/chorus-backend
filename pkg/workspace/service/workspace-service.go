package service

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"

	"github.com/pkg/errors"
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
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error)
	CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (uint64, error)
	UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) error
	DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error
}

type WorkspaceService struct {
	store  WorkspaceStore
	client helm.HelmClienter
}

func NewWorkspaceService(store WorkspaceStore, client helm.HelmClienter) *WorkspaceService {
	return &WorkspaceService{
		store:  store,
		client: client,
	}
}

func (u *WorkspaceService) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error) {
	workspaces, err := u.store.ListWorkspaces(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query workspaces")
	}
	return workspaces, nil
}

func (u *WorkspaceService) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	workspace, err := u.store.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get workspace %v", workspace.ID)
	}

	return workspace, nil
}

func (u *WorkspaceService) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	err := u.store.DeleteWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		return errors.Wrapf(err, "unable to delete workspace %v", workspaceID)
	}

	// TODO implement delete all workspaces and appInstances

	// err = u.client.DeleteWorkbench(u.getWorkbenchName(workspaceID), u.getWorkbenchName(workspaceID))
	// if err != nil {
	// 	return errors.Wrapf(err, "unable to delete workbench %v", workspaceID)
	// }

	return nil
}

func (u *WorkspaceService) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) error {
	if err := u.store.UpdateWorkspace(ctx, workspace.TenantID, workspace); err != nil {
		return errors.Wrapf(err, "unable to update workspace %v", workspace.ID)
	}

	return nil
}

func (u *WorkspaceService) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (uint64, error) {
	id, err := u.store.CreateWorkspace(ctx, workspace.TenantID, workspace)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create workspace %v", workspace.ID)
	}

	// should we do something as we can lazily create namespace with helm if needed..?

	// err = u.client.CreateWorkbench(u.getWorkbenchName(id), u.getWorkbenchName(id))
	// if err != nil {
	// 	return 0, errors.Wrapf(err, "unable to create workbench %v", workspace.ID)
	// }

	return id, nil
}

// func (u *WorkspaceService) getWorkbenchName(id uint64) string {
// 	return "workbench" + fmt.Sprintf("%v", id)
// }
