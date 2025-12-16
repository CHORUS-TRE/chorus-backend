package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/blockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
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

func (c workspaceServiceLogging) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*blockstore.File, error) {
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

func (c workspaceServiceLogging) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*blockstore.File, error) {
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

func (c workspaceServiceLogging) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *blockstore.File) (*blockstore.File, error) {
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

func (c workspaceServiceLogging) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *blockstore.File) (*blockstore.File, error) {
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

func (c workspaceServiceLogging) InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *blockstore.File) (*blockstore.FileUploadInfo, error) {
	now := time.Now()

	uploadInfo, err := c.next.InitiateWorkspaceFileUpload(ctx, workspaceID, filePath, file)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to initiate workspace file upload: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return uploadInfo, nil
}

func (c workspaceServiceLogging) UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *blockstore.FilePart) (*blockstore.FilePart, error) {
	now := time.Now()

	uploadedPart, err := c.next.UploadWorkspaceFilePart(ctx, workspaceID, filePath, uploadID, part)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to upload workspace file part: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return uploadedPart, nil
}

func (c workspaceServiceLogging) CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*blockstore.FilePart) (*blockstore.File, error) {
	now := time.Now()

	completedFile, err := c.next.CompleteWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID, parts)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to complete workspace file upload: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return completedFile, nil
}

func (c workspaceServiceLogging) AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error {
	now := time.Now()

	err := c.next.AbortWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to abort workspace file upload: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}
