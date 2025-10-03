package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ MinioClienter = &client{}

type MinioClienter interface {
	// Workspace object operations
	StatWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error)
	GetWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error)
	ListWorkspaceObjects(workspaceID uint64, path string) ([]*workspace_model.WorkspaceFile, error)
	PutWorkspaceObject(workspaceID uint64, path string, content []byte, contentType string) (*workspace_model.WorkspaceFile, error)
	DeleteWorkspaceObject(workspaceID uint64, path string) error
}

type client struct {
	cfg            config.Config
	minioClientCfg MinioClientConfig

	minioClient *minio.Client
}

func NewClient(cfg config.Config) (*client, error) {
	clientCfg, err := getMinioClientConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting minio config: %w", err)
	}

	minioClient, err := minio.New(clientCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(clientCfg.AccessKeyID, clientCfg.SecretAccessKey, ""),
		Secure: clientCfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating minio client: %w", err)
	}

	return &client{
		cfg:            cfg,
		minioClientCfg: clientCfg,
		minioClient:    minioClient,
	}, nil
}

func (c *client) StatWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error) {
	objectKey := WorkspacePathToObjectKey(workspaceID, path)

	objectInfo, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	file, err := c.ObjectToWorkspaceFile(objectInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to convert object to workspace file: %w", err)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully retrieved %s metadata from workspace %d\n", objectKey, workspaceID))

	return &file, nil
}

func (c *client) GetWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error) {
	objectKey := WorkspacePathToObjectKey(workspaceID, path)

	reader, err := c.minioClient.GetObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get object %s: %w", objectKey, err)
	}

	stat, err := reader.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}
	defer reader.Close()

	file, err := c.ObjectToWorkspaceFile(stat)
	if err != nil {
		return nil, fmt.Errorf("unable to convert object to workspace file: %w", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read object content: %w", err)
	}
	file.Content = content

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully downloaded %s from workspace %d\n", objectKey, workspaceID))

	return &file, nil
}

func (c *client) ListWorkspaceObjects(workspaceID uint64, path string) ([]*workspace_model.WorkspaceFile, error) {
	objectKey := WorkspacePathToObjectKey(workspaceID, path)

	files := []*workspace_model.WorkspaceFile{}
	objectCh := c.minioClient.ListObjects(context.Background(), c.minioClientCfg.BucketName, minio.ListObjectsOptions{
		Prefix:       objectKey,
		WithMetadata: true,
		// Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error while streaming the response from the object: %w", object.Err)
		}
		file, err := c.ObjectToWorkspaceFile(object)
		if err != nil {
			return nil, fmt.Errorf("unable to convert object to workspace file: %w", err)
		}
		files = append(files, &file)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully listed objects under %s in workspace %d\n", objectKey, workspaceID))

	return files, nil
}

func (c *client) PutWorkspaceObject(workspaceID uint64, path string, content []byte, contentType string) (*workspace_model.WorkspaceFile, error) {
	objectKey := WorkspacePathToObjectKey(workspaceID, path)

	_, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return nil, fmt.Errorf("object at %s already exists in workspace %d", objectKey, workspaceID)
	}

	_, err = c.minioClient.PutObject(context.Background(), c.minioClientCfg.BucketName, objectKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", objectKey, err)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully uploaded %s in workspace %d\n", objectKey, workspaceID))

	objectInfo, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to verify uploaded object %s: %w", objectKey, err)
	}

	file, err := c.ObjectToWorkspaceFile(objectInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to convert object to workspace file: %w", err)
	}

	return &file, nil
}

func (c *client) DeleteWorkspaceObject(workspaceID uint64, path string) error {
	objectKey := WorkspacePathToObjectKey(workspaceID, path)

	err := c.minioClient.RemoveObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", objectKey, err)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully deleted %s from workspace %d\n", objectKey, workspaceID))

	return nil
}
