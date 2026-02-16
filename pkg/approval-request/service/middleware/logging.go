package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"go.uber.org/zap"
)

type approvalRequestServiceLogging struct {
	logger *logger.ContextLogger
	next   service.ApprovalRequester
}

func Logging(logger *logger.ContextLogger) func(service.ApprovalRequester) service.ApprovalRequester {
	return func(next service.ApprovalRequester) service.ApprovalRequester {
		return &approvalRequestServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c approvalRequestServiceLogging) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	now := time.Now()

	res, err := c.next.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to get approval request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c approvalRequestServiceLogging) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter service.ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	now := time.Now()

	res, paginationRes, err := c.next.ListApprovalRequests(ctx, tenantID, userID, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to list approval requests: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_requests", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c approvalRequestServiceLogging) CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	now := time.Now()

	res, err := c.next.CreateDataExtractionRequest(ctx, request, filePaths)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create data extraction request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c approvalRequestServiceLogging) CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	now := time.Now()

	res, err := c.next.CreateDataTransferRequest(ctx, request, filePaths)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create data transfer request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c approvalRequestServiceLogging) ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error) {
	now := time.Now()

	res, err := c.next.ApproveApprovalRequest(ctx, tenantID, requestID, userID, approve)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Bool("approve", approve),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to approve approval request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Bool("approve", approve),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c approvalRequestServiceLogging) DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	now := time.Now()

	err := c.next.DeleteApprovalRequest(ctx, tenantID, requestID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete approval request: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c approvalRequestServiceLogging) DownloadApprovalRequestFile(ctx context.Context, tenantID, requestID uint64, filePath string) (*model.ApprovalRequestFile, []byte, error) {
	now := time.Now()

	file, content, err := c.next.DownloadApprovalRequestFile(ctx, tenantID, requestID, filePath)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("request_id", requestID),
			zap.String("file_path", filePath),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to download approval request file: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", requestID),
		zap.String("file_path", filePath),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return file, content, nil
}
