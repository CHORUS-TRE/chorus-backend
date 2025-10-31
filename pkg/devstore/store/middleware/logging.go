package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"
	"go.uber.org/zap"
)

type devstoreStorageLogging struct {
	logger *logger.ContextLogger
	next   service.DevstoreStore
}

func Logging(logger *logger.ContextLogger) func(service.DevstoreStore) service.DevstoreStore {
	return func(next service.DevstoreStore) service.DevstoreStore {
		return &devstoreStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c *devstoreStorageLogging) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	res, err := c.next.ListEntries(ctx, tenantID, scope, scopeID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.String("scope", string(scope)),
			zap.Uint64("scopeID", scopeID),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithCountField(len(res)),
		zap.String("scope", string(scope)),
		zap.Uint64("scopeID", scopeID),
	)
	return res, nil
}

func (c *devstoreStorageLogging) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	res, err := c.next.GetEntry(ctx, tenantID, scope, scopeID, key)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.String("scope", string(scope)),
			zap.Uint64("scopeID", scopeID),
			zap.String("key", key),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("scope", string(scope)),
		zap.Uint64("scopeID", scopeID),
		zap.String("key", key),
	)
	return res, nil
}

func (c *devstoreStorageLogging) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	res, err := c.next.PutEntry(ctx, tenantID, scope, scopeID, key, value)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.String("scope", string(scope)),
			zap.Uint64("scopeID", scopeID),
			zap.String("key", key),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("scope", string(scope)),
		zap.Uint64("scopeID", scopeID),
		zap.String("key", key),
	)
	return res, nil
}

func (c *devstoreStorageLogging) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	err := c.next.DeleteEntry(ctx, tenantID, scope, scopeID, key)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.String("scope", string(scope)),
			zap.Uint64("scopeID", scopeID),
			zap.String("key", key),
		)
		return err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("scope", string(scope)),
		zap.Uint64("scopeID", scopeID),
		zap.String("key", key),
	)
	return nil
}
