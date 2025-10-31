package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"
	"go.uber.org/zap"
)

type devstoreServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Devstorer
}

func Logging(logger *logger.ContextLogger) func(service.Devstorer) service.Devstorer {
	return func(next service.Devstorer) service.Devstorer {
		return &devstoreServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (s *devstoreServiceLogging) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error) {
	now := time.Now()

	entries, err := s.next.ListEntries(ctx, tenantID, scope, scopeID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to list devstore entries: %w", err)
	}

	s.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_entries", len(entries)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return entries, nil
}

func (s *devstoreServiceLogging) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error) {
	now := time.Now()

	entry, err := s.next.GetEntry(ctx, tenantID, scope, scopeID, key)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to get devstore entry: %w", err)
	}

	s.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("key", key),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return entry, nil
}

func (s *devstoreServiceLogging) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	now := time.Now()

	res, err := s.next.PutEntry(ctx, tenantID, scope, scopeID, key, value)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to put devstore entry: %w", err)
	}

	s.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("key", key),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (s *devstoreServiceLogging) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	now := time.Now()

	err := s.next.DeleteEntry(ctx, tenantID, scope, scopeID, key)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete devstore entry: %w", err)
	}

	s.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("key", key),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
