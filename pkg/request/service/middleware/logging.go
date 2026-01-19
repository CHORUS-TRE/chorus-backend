package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/service"

	"go.uber.org/zap"
)

type requestServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Requester
}

func Logging(logger *logger.ContextLogger) func(service.Requester) service.Requester {
	return func(next service.Requester) service.Requester {
		return &requestServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c requestServiceLogging) GetRequest(ctx context.Context, tenantID, requestID uint64) (*model.Request, error) {
	now := time.Now()

	res, err := c.next.GetRequest(ctx, tenantID, requestID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to get request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c requestServiceLogging) ListRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter service.RequestFilter) ([]*model.Request, *common_model.PaginationResult, error) {
	now := time.Now()

	res, paginationRes, err := c.next.ListRequests(ctx, tenantID, userID, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to list requests: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_requests", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c requestServiceLogging) CreateRequest(ctx context.Context, request *model.Request, filePaths []string) (*model.Request, error) {
	now := time.Now()

	res, err := c.next.CreateRequest(ctx, request, filePaths)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c requestServiceLogging) ApproveRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.Request, error) {
	now := time.Now()

	res, err := c.next.ApproveRequest(ctx, tenantID, requestID, userID, approve)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Bool("approve", approve),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to approve request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Bool("approve", approve),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c requestServiceLogging) DeleteRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	now := time.Now()

	err := c.next.DeleteRequest(ctx, tenantID, requestID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
