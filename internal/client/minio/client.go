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
	StatObject(objectKey string) (*workspace_model.WorkspaceFile, error)
	GetObject(objectKey string) (*workspace_model.WorkspaceFile, error)
	ListObjects(objectKey string) ([]*workspace_model.WorkspaceFile, error)
	PutObject(objectKey string, content []byte) error
	DeleteObject(objectKey string) error
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

func (c *client) StatObject(objectKey string) (*workspace_model.WorkspaceFile, error) {
	objectInfo, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	file, err := c.ObjectToWorkspaceFile(objectInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to convert object to workspace file: %w", err)
	}

	return &file, nil
}

func (c *client) GetObject(objectKey string) (*workspace_model.WorkspaceFile, error) {
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

	return &file, nil
}

func (c *client) ListObjects(objectKey string) ([]*workspace_model.WorkspaceFile, error) {
	files := []*workspace_model.WorkspaceFile{}
	objectCh := c.minioClient.ListObjects(context.Background(), c.minioClientCfg.BucketName, minio.ListObjectsOptions{
		Prefix: objectKey,
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
	return files, nil
}

func (c *client) PutObject(objectKey string, content []byte) error {
	n, err := c.minioClient.PutObject(context.Background(), c.minioClientCfg.BucketName, objectKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to put object at %s: %w", objectKey, err)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully uploaded %s of size %d\n", objectKey, n.Size))

	return nil
}

func (c *client) DeleteObject(objectKey string) error {
	err := c.minioClient.RemoveObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", objectKey, err)
	}

	logger.TechLog.Info(context.Background(), fmt.Sprintf("Successfully deleted %s\n", objectKey))

	return nil
}
