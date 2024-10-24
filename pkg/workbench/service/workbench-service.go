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
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"go.uber.org/zap"
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
	cfg        config.Config
	store      WorkbenchStore
	client     helm.HelmClienter
	rwMutex    sync.RWMutex
	proxyCache map[proxyID]*proxy
}

func NewWorkbenchService(cfg config.Config, store WorkbenchStore, client helm.HelmClienter) *WorkbenchService {
	return &WorkbenchService{
		cfg:        cfg,
		store:      store,
		client:     client,
		proxyCache: make(map[proxyID]*proxy),
	}
}

func (s *WorkbenchService) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	workbenchs, err := s.store.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		return nil, fmt.Errorf("unable to query workbenchs: %w", err)
	}
	return workbenchs, nil
}

func (s *WorkbenchService) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("unable to get workbench %v: %w", workbench.ID, err)
	}

	return workbench, nil
}

func (s *WorkbenchService) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	workbench, err := s.store.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to get workbench %v: %w", workbench.ID, err)
	}

	err = s.store.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workbenchID, err)
	}

	err = s.client.DeleteWorkbench(s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(workbenchID))
	if err != nil {
		return fmt.Errorf("unable to delete workbench %v: %w", workbenchID, err)
	}

	return nil
}

func (s *WorkbenchService) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	if err := s.store.UpdateWorkbench(ctx, workbench.TenantID, workbench); err != nil {
		return fmt.Errorf("unable to update workbench %v: %w", workbench.ID, err)
	}

	return nil
}

func (s *WorkbenchService) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	id, err := s.store.CreateWorkbench(ctx, workbench.TenantID, workbench)
	if err != nil {
		return 0, fmt.Errorf("unable to create workbench %v: %w", workbench.ID, err)
	}

	namespace, workbenchName := s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(id)

	err = s.client.CreateWorkbench(namespace, workbenchName)
	if err != nil {
		return 0, fmt.Errorf("unable to create workbench %v: %w", workbench.ID, err)
	}

	return id, nil
}

func (s *WorkbenchService) getProxy(proxyID proxyID) (*proxy, error) {
	// TODO error handling, port forwarding re-creation, cache eviction, cleaning on cache evit and sig stop
	s.rwMutex.RLock()
	if proxy, exists := s.proxyCache[proxyID]; exists {
		s.rwMutex.RUnlock()
		return proxy, nil
	}
	s.rwMutex.RUnlock()

	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	var xpraUrl string
	var port uint16
	var stopChan chan struct{}
	var err error
	if !s.cfg.Services.WorkbenchService.BackendInK8S {
		port, stopChan, err = s.client.CreatePortForward(proxyID.namespace, proxyID.workbench)
		if err != nil {
			return nil, fmt.Errorf("Failed to create port forward: %w", err)
		}

		xpraUrl = fmt.Sprintf("http://localhost:%v", port)
	} else {
		xpraUrl = fmt.Sprintf("http://%v.%v:8080", proxyID.workbench, proxyID.namespace)
	}
	logger.TechLog.Debug(context.Background(), "targetUrl", zap.String("xpraUrl", xpraUrl))

	targetURL, err := url.Parse(xpraUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse url: %w", err)
	}

	reg := regexp.MustCompile(`^/api/rest/v1/workbenchs/[0-9]+/stream`)

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := reverseProxy.Director

	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

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
		return fmt.Errorf("unable to get workbench %v: %w", workbench.ID, err)
	}

	namespace, workbenchName := s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(workbenchID)

	proxyID := proxyID{
		namespace: namespace,
		workbench: workbenchName,
	}

	proxy, err := s.getProxy(proxyID)
	if err != nil {
		return fmt.Errorf("unable to get proxy %v: %w", proxyID, err)
	}

	proxy.reverseProxy.ServeHTTP(w, r)

	return nil
}

func (s *WorkbenchService) getWorkspaceName(id uint64) string {
	return fmt.Sprintf("workspace%v", id)
}
func (s *WorkbenchService) getWorkbenchName(id uint64) string {
	return fmt.Sprintf("workbench%v", id)
}
