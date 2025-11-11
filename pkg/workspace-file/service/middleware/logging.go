package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"

	"go.uber.org/zap"
)

type workspaceServiceLogging struct {
	logger *logger.ContextLogger
	next   service.WorkspaceFiler
}

func Logging(logger *logger.ContextLogger) func(service.WorkspaceFiler) service.WorkspaceFiler {
	return func(next service.WorkspaceFiler) service.WorkspaceFiler {
		return &workspaceServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workspaceServiceLogging) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	now := time.Now()

	res, err := c.next.GetWorkspaceFile(ctx, workspaceID, filePath)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to get workspace file: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workspaceServiceLogging) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	now := time.Now()

	files, err := c.next.ListWorkspaceFiles(ctx, workspaceID, filePath)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to list workspace files: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Int("num_files", len(files)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return files, nil
}

func (c workspaceServiceLogging) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	now := time.Now()

	newFile, err := c.next.CreateWorkspaceFile(ctx, workspaceID, file)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create workspace file: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return newFile, nil
}

func (c workspaceServiceLogging) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	now := time.Now()

	updatedFile, err := c.next.UpdateWorkspaceFile(ctx, workspaceID, oldPath, file)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to update workspace file: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return updatedFile, nil
}

func (c workspaceServiceLogging) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	now := time.Now()

	err := c.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete workspace file: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}
