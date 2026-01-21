package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"go.uber.org/zap"
)

type auditServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Auditer
}

func Logging(logger *logger.ContextLogger) func(service.Auditer) service.Auditer {
	return func(next service.Auditer) service.Auditer {
		return &auditServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c auditServiceLogging) Record(ctx context.Context, entry *model.AuditEntry) error {
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

func (c auditServiceLogging) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	entries, paginationRes, err := c.next.List(ctx, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Int("num_entries", len(entries)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return entries, paginationRes, nil
}
