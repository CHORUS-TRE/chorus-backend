package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"

	"go.uber.org/zap"
)

type approvalRequestStorageLogging struct {
	logger *logger.ContextLogger
	next   service.ApprovalRequestStore
}

func Logging(log *logger.ContextLogger) func(service.ApprovalRequestStore) service.ApprovalRequestStore {
	l := logger.With(log, zap.String("layer", "store"))
	return func(next service.ApprovalRequestStore) service.ApprovalRequestStore {
		return &approvalRequestStorageLogging{
			logger: l,
			next:   next,
		}
	}
}

func (s *approvalRequestStorageLogging) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	log := logger.With(s.logger,
		zap.String("method", "GetApprovalRequest"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("request_id", requestID),
	)
	log.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	log.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (s *approvalRequestStorageLogging) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter service.ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	log := logger.With(s.logger,
		zap.String("method", "ListApprovalRequests"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("user_id", userID),
	)
	log.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, paginationRes, err := s.next.ListApprovalRequests(ctx, tenantID, userID, pagination, filter)
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	log.Debug(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (s *approvalRequestStorageLogging) CreateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error) {
	log := logger.With(s.logger,
		zap.String("method", "CreateApprovalRequest"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("requester_id", request.RequesterID),
		zap.String("type", string(request.Type)),
	)
	log.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.CreateApprovalRequest(ctx, tenantID, request)
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	log.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("request_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (s *approvalRequestStorageLogging) UpdateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error) {
	log := logger.With(s.logger,
		zap.String("method", "UpdateApprovalRequest"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("request_id", request.ID),
	)
	log.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.UpdateApprovalRequest(ctx, tenantID, request)
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	log.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (s *approvalRequestStorageLogging) DeleteApprovalRequest(ctx context.Context, tenantID, requestID uint64) error {
	log := logger.With(s.logger,
		zap.String("method", "DeleteApprovalRequest"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("request_id", requestID),
	)
	log.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.DeleteApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	log.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
