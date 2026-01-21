package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	"go.uber.org/zap"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
)

type auditStorageLogging struct {
	logger *logger.ContextLogger
	next   service.AuditStore
}

func Logging(logger *logger.ContextLogger) func(service.AuditStore) service.AuditStore {
	return func(next service.AuditStore) service.AuditStore {
		return &auditStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c auditStorageLogging) Record(ctx context.Context, entry *model.AuditEntry) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	err := c.next.Record(ctx, entry)
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

func (c auditStorageLogging) RecordBatch(ctx context.Context, entries []*model.AuditEntry) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	err := c.next.RecordBatch(ctx, entries)
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

func (c auditStorageLogging) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, paginationRes, err := c.next.List(ctx, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c auditStorageLogging) Count(ctx context.Context, filter *model.AuditFilter) (int64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	count, err := c.next.Count(ctx, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(int(count)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return count, nil
}
