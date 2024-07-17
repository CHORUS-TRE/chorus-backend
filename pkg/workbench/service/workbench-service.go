package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"

	"github.com/pkg/errors"
)

type Workbencher interface {
	GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error)
	CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error)
	UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error
	DeleteWorkbench(ctx context.Context, tenantId, workbenchId uint64) error
}

type WorkbenchStore interface {
	GetWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error)
	CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (uint64, error)
	UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) error
	DeleteWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) error
}

type WorkbenchService struct {
	store  WorkbenchStore
	client helm.HelmClienter
}

func NewWorkbenchService(store WorkbenchStore, client helm.HelmClienter) *WorkbenchService {
	return &WorkbenchService{
		store:  store,
		client: client,
	}
}

func (u *WorkbenchService) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	workbenchs, err := u.store.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query workbenchs")
	}
	return workbenchs, nil
}

func (u *WorkbenchService) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := u.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get workbench %v", workbench.ID)
	}

	return workbench, nil
}

func (u *WorkbenchService) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	workbench, err := u.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return errors.Wrapf(err, "unable to get workbench %v", workbench.ID)
	}

	err = u.store.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return errors.Wrapf(err, "unable to delete workbench %v", workbenchID)
	}

	err = u.client.DeleteWorkbench(u.getWorkspaceName(workbench.WorkspaceID), u.getWorkbenchName(workbenchID))
	if err != nil {
		return errors.Wrapf(err, "unable to delete workbench %v", workbenchID)
	}

	return nil
}

func (u *WorkbenchService) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	if err := u.store.UpdateWorkbench(ctx, workbench.TenantID, workbench); err != nil {
		return errors.Wrapf(err, "unable to update workbench %v", workbench.ID)
	}

	return nil
}

func (u *WorkbenchService) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	id, err := u.store.CreateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create workbench %v", workbench.ID)
	}

	err = u.client.CreateWorkbench(u.getWorkspaceName(workbench.WorkspaceID), u.getWorkbenchName(id))
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create workbench %v", workbench.ID)
	}

	return id, nil
}

func (u *WorkbenchService) getWorkspaceName(id uint64) string {
	return "workspace" + fmt.Sprintf("%v", id)
}
func (u *WorkbenchService) getWorkbenchName(id uint64) string {
	return "workbench" + fmt.Sprintf("%v", id)
}
