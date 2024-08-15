package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"

	"github.com/pkg/errors"
)

type Workbencher interface {
	GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error)
	ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error)
	CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error)
	ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error
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
	store      WorkbenchStore
	client     helm.HelmClienter
	mutex      sync.Mutex
	proxyCache map[proxyID]*proxy
}

func NewWorkbenchService(store WorkbenchStore, client helm.HelmClienter) *WorkbenchService {
	return &WorkbenchService{
		store:      store,
		client:     client,
		proxyCache: make(map[proxyID]*proxy),
	}
}

func (s *WorkbenchService) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	workbenchs, err := s.store.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query workbenchs")
	}
	return workbenchs, nil
}

func (s *WorkbenchService) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get workbench %v", workbench.ID)
	}

	return workbench, nil
}

func (s *WorkbenchService) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return errors.Wrapf(err, "unable to get workbench %v", workbench.ID)
	}

	err = s.store.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return errors.Wrapf(err, "unable to delete workbench %v", workbenchID)
	}

	err = s.client.DeleteWorkbench(s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(workbenchID))
	if err != nil {
		return errors.Wrapf(err, "unable to delete workbench %v", workbenchID)
	}

	return nil
}

func (s *WorkbenchService) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	if err := s.store.UpdateWorkbench(ctx, workbench.TenantID, workbench); err != nil {
		return errors.Wrapf(err, "unable to update workbench %v", workbench.ID)
	}

	return nil
}

func (s *WorkbenchService) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	id, err := s.store.CreateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create workbench %v", workbench.ID)
	}

	namespace, workbenchName := s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(id)

	err = s.client.CreateWorkbench(namespace, workbenchName)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to create workbench %v", workbench.ID)
	}

	return id, nil
}

func (s *WorkbenchService) getProxy(proxyID proxyID) (*proxy, error) {
	// TODO error handling, port forwarding re-creation, cache eviction, cleaning on cache evit and sig stop
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if proxy, exists := s.proxyCache[proxyID]; exists {
		return proxy, nil
	}

	port, stopChan, err := s.client.CreatePortForward(proxyID.namespace, proxyID.workbench)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create port forward: %v", err)
	}

	targetURL, err := url.Parse(fmt.Sprintf("http://localhost:%v", port))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse url: %v", err)

	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := reverseProxy.Director

	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

		reg := regexp.MustCompile(`^/api/rest/v1/workbenchs/[0-9]+/stream`)
		req.URL.Path = reg.ReplaceAllString(req.URL.Path, "")
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
		return errors.Wrapf(err, "unable to get workbench %v", workbench.ID)
	}

	namespace, workbenchName := s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(workbenchID)

	proxyID := proxyID{
		namespace: namespace,
		workbench: workbenchName,
	}

	proxy, err := s.getProxy(proxyID)
	if err != nil {
		return errors.Wrapf(err, "unable to get proxy %v", proxyID)
	}

	proxy.reverseProxy.ServeHTTP(w, r)

	return nil
}

func (s *WorkbenchService) getWorkspaceName(id uint64) string {
	return "workspace" + fmt.Sprintf("%v", id)
}
func (s *WorkbenchService) getWorkbenchName(id uint64) string {
	return "workbench" + fmt.Sprintf("%v", id)
}
