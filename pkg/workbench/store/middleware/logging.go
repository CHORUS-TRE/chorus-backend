package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"go.uber.org/zap"
)

type workbenchStorageLogging struct {
	logger *logger.ContextLogger
	next   service.WorkbenchStore
}

func Logging(logger *logger.ContextLogger) func(service.WorkbenchStore) service.WorkbenchStore {
	return func(next service.WorkbenchStore) service.WorkbenchStore {
		return &workbenchStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workbenchStorageLogging) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) ListWorkbenchAppInstances(ctx context.Context, workbenchID uint64) ([]*model.AppInstance, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListWorkbenchAppInstances(ctx, workbenchID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) ListAllWorkbenches(ctx context.Context) ([]*model.Workbench, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListAllWorkbenches(ctx)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) SaveBatchProxyHit(ctx context.Context, proxyHitCountMap map[uint64]uint64, proxyHitDateMap map[uint64]time.Time) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	err := c.next.SaveBatchProxyHit(ctx, proxyHitCountMap, proxyHitDateMap)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) GetWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) (*model.Workbench, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbenchID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbenchID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateWorkbench(ctx, tenantID, workbench)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbench.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbench.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	workbenchId, err := c.next.CreateWorkbench(ctx, tenantID, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return workbenchId, nil
}

func (c workbenchStorageLogging) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateAppInstance(ctx, tenantID, appInstance)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstance.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstance.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) UpdateAppInstances(ctx context.Context, tenantID uint64, appInstances []*model.AppInstance) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateAppInstances(ctx, tenantID, appInstances)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	appInstanceId, err := c.next.CreateAppInstance(ctx, tenantID, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return appInstanceId, nil
}
