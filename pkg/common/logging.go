package common

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"go.uber.org/zap"
)

func LogErrorIfAny(err error, ctx context.Context, now time.Time, log *logger.ContextLogger) error {
	elapsed := zap.Int64(logger.LoggerKeyElapsedMs, time.Since(now).Milliseconds())
	if err != nil {
		log.Error(ctx, logger.LoggerMessageRequestFailed, zap.Error(err), elapsed)
		return err
	}

	log.Debug(ctx, "request completed", elapsed)
	return nil
}
