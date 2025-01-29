package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
	app_instance_model "github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	workbenchProxyRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "workbench_service_proxy_request",
		Help: "The total number of request proxied to a workbench via the backend",
	}, []string{"workbench_id"})

	_ = prometheus.DefaultRegisterer.Register(workbenchProxyRequest)
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
	ListWorkbenchAppInstances(ctx context.Context, workbenchID uint64) ([]*app_instance_model.AppInstance, error)
	ListAllActiveWorkbenchs(ctx context.Context) ([]*model.Workbench, error)
	SaveBatchProxyHit(ctx context.Context, proxyHitCountMap map[uint64]uint64, proxyHitDateMap map[uint64]time.Time) error
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
	cfg              config.Config
	store            WorkbenchStore
	client           helm.HelmClienter
	proxyRWMutex     sync.RWMutex
	proxyCache       map[proxyID]*proxy
	proxyHitMutex    sync.Mutex
	proxyHitCountMap map[uint64]uint64
	proxyHitDateMap  map[uint64]time.Time
}

func NewWorkbenchService(cfg config.Config, store WorkbenchStore, client helm.HelmClienter) *WorkbenchService {
	s := &WorkbenchService{
		cfg:              cfg,
		store:            store,
		client:           client,
		proxyCache:       make(map[proxyID]*proxy),
		proxyHitCountMap: make(map[uint64]uint64),
		proxyHitDateMap:  make(map[uint64]time.Time),
	}

	go func() {
		s.updateAllWorkbenchs(context.Background())
	}()

	go func() {
		for {
			s.saveBatchProxyHit(context.Background())
			randomDelayToAvoidCollision := time.Duration(rand.Int64N(int64(10 * time.Second)))
			time.Sleep(cfg.Services.WorkbenchService.ProxyHitSaveBatchInterval + randomDelayToAvoidCollision)
		}
	}()

	return s
}

func (s *WorkbenchService) updateAllWorkbenchs(ctx context.Context) {
	workbenchs, err := s.store.ListAllActiveWorkbenchs(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, "unable to query workbenchs", zap.Error(err))
		return
	}

	for _, workbench := range workbenchs {
		apps, err := s.store.ListWorkbenchAppInstances(ctx, workbench.ID)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to list app instances", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
			continue
		}
		clientApps := []helm.AppInstance{}
		for _, app := range apps {
			clientApps = append(clientApps, helm.AppInstance{
				AppName:     utils.ToString(app.AppName),
				AppRegistry: utils.ToString(app.AppDockerImageRegistry),
				AppImage:    utils.ToString(app.AppDockerImageName),
				AppVersion:  utils.ToString(app.AppDockerImageTag),
			})
		}

		namespace, workbenchName := s.getWorkspaceName(workbench.WorkspaceID), s.getWorkbenchName(workbench.ID)

		err = s.client.UpdateWorkbench(namespace, workbenchName, clientApps)
		if err != nil {
			logger.TechLog.Error(ctx, "unable to update workbench", zap.Error(err), zap.Uint64("workbenchID", workbench.ID))
		}
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

func (s *WorkbenchService) getWorkspaceName(id uint64) string {
	return fmt.Sprintf("workspace%v", id)
}
func (s *WorkbenchService) getWorkbenchName(id uint64) string {
	return fmt.Sprintf("workbench%v", id)
}
