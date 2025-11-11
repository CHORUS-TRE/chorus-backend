package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var _ MinioClienter = &client{}

type MinioClienter interface {
	GetClientName() string
	GetClientPrefix() string

	StatObject(objectKey string) (*MinioObjectInfo, error)
	GetObject(objectKey string) (*MinioObject, error)
	ListObjects(objectKey string) ([]*MinioObjectInfo, error)
	PutObject(objectKey string, object *MinioObject) (*MinioObjectInfo, error)
	DeleteObject(objectKey string) error
}

type client struct {
	cfg            config.Config
	minioClientCfg MinioClientConfig

	minioClient *minio.Client
}

func NewClient(cfg config.Config, clientName string) (*client, error) {
	clientCfg, err := getMinioClientConfig(cfg, clientName)
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

func (c *client) GetClientName() string {
	return c.minioClientCfg.Name
}

func (c *client) GetClientPrefix() string {
	return c.minioClientCfg.Prefix
}

func (c *client) StatObject(objectKey string) (*MinioObjectInfo, error) {
	objectInfo, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}

	return &MinioObjectInfo{
		Key:          objectInfo.Key,
		Size:         objectInfo.Size,
		LastModified: objectInfo.LastModified,
		MimeType:     objectInfo.ContentType,
	}, nil
}

func (c *client) GetObject(objectKey string) (*MinioObject, error) {
	reader, err := c.minioClient.GetObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get object %s: %w", objectKey, err)
	}

	stat, err := reader.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to stat object %s: %w", objectKey, err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read object content: %w", err)
	}

	return &MinioObject{
		MinioObjectInfo: MinioObjectInfo{
			Key:          stat.Key,
			Size:         stat.Size,
			LastModified: stat.LastModified,
			MimeType:     stat.ContentType,
		},
		Content: content,
	}, nil
}

func (c *client) ListObjects(objectKey string) ([]*MinioObjectInfo, error) {
	objects := []*MinioObjectInfo{}
	objectCh := c.minioClient.ListObjects(context.Background(), c.minioClientCfg.BucketName, minio.ListObjectsOptions{
		Prefix:       objectKey,
		WithMetadata: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		objects = append(objects, &MinioObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			MimeType:     object.ContentType,
		})
	}

	return objects, nil
}

func (c *client) PutObject(objectKey string, object *MinioObject) (*MinioObjectInfo, error) {
	_, err := c.minioClient.PutObject(context.Background(), c.minioClientCfg.BucketName, objectKey, bytes.NewReader(object.Content), int64(len(object.Content)), minio.PutObjectOptions{ContentType: object.MimeType})
	if err != nil {
		return nil, fmt.Errorf("unable to put object at %s: %w", objectKey, err)
	}

	objectInfo, err := c.minioClient.StatObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to verify uploaded object %s: %w", objectKey, err)
	}

	return &MinioObjectInfo{
		Key:          objectInfo.Key,
		Size:         objectInfo.Size,
		LastModified: objectInfo.LastModified,
		MimeType:     objectInfo.ContentType,
	}, nil
}

func (c *client) DeleteObject(objectKey string) error {
	err := c.minioClient.RemoveObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", objectKey, err)
	}

	return nil
}
