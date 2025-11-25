package miniorawclient

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
	StatObject(objectKey string) (*MinioObjectInfo, error)
	GetObject(objectKey string) (*MinioObject, error)
	ListObjects(objectKey string, recursive bool) ([]*MinioObjectInfo, error)
	PutObject(objectKey string, object *MinioObject) (*MinioObjectInfo, error)
	MoveObject(oldObjectKey string, newObjectKey string) error
	DeleteObject(objectKey string) error
	NewMultipartUpload(objectKey string, objectSize uint64) (string, error)
	PutObjectPart(uploadId string, partNumber int, data []byte) (*MinioObjectPartInfo, error)
	CompleteMultipartUpload(uploadId string, parts []*MinioObjectPartInfo) (*MinioObject, error)
	AbortMultipartUpload(uploadId string) error
}

type client struct {
	cfg            config.Config
	minioClientCfg MinioClientConfig

	minioClient *minio.Client
	minioCore   *minio.Core
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

	minioCore, err := minio.NewCore(clientCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(clientCfg.AccessKeyID, clientCfg.SecretAccessKey, ""),
		Secure: clientCfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating minio core: %w", err)
	}

	return &client{
		cfg:            cfg,
		minioClientCfg: clientCfg,
		minioClient:    minioClient,
		minioCore:      minioCore,
	}, nil
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

func (c *client) ListObjects(objectKey string, recursive bool) ([]*MinioObjectInfo, error) {
	objects := []*MinioObjectInfo{}
	objectCh := c.minioClient.ListObjects(context.Background(), c.minioClientCfg.BucketName, minio.ListObjectsOptions{
		Prefix:       objectKey,
		WithMetadata: true,
		Recursive:    recursive,
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

func (c *client) MoveObject(oldObjectKey string, newObjectKey string) error {
	src := minio.CopySrcOptions{
		Bucket: c.minioClientCfg.BucketName,
		Object: oldObjectKey,
	}

	dst := minio.CopyDestOptions{
		Bucket: c.minioClientCfg.BucketName,
		Object: newObjectKey,
	}

	_, err := c.minioClient.CopyObject(context.Background(), dst, src)
	if err != nil {
		return fmt.Errorf("unable to move object from %s to %s: %w", oldObjectKey, newObjectKey, err)
	}

	err = c.minioClient.RemoveObject(context.Background(), c.minioClientCfg.BucketName, oldObjectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete old object at %s after move: %w", oldObjectKey, err)
	}

	return nil
}

func (c *client) DeleteObject(objectKey string) error {
	err := c.minioClient.RemoveObject(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete object at %s: %w", objectKey, err)
	}

	return nil
}

func (c *client) NewMultipartUpload(objectKey string, objectSize uint64) (string, error) {
	uploadId, err := c.minioCore.NewMultipartUpload(context.Background(), c.minioClientCfg.BucketName, objectKey, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to initiate multipart upload for object %s: %w", objectKey, err)
	}

	return uploadId, nil
}

func (c *client) PutObjectPart(uploadId string, partNumber int, data []byte) (*MinioObjectPartInfo, error) {
	objectPart, err := c.minioCore.PutObjectPart(context.Background(), c.minioClientCfg.BucketName, "", uploadId, partNumber, bytes.NewReader(data), int64(len(data)), minio.PutObjectPartOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to upload part %d for upload %s: %w", partNumber, uploadId, err)
	}

	return &MinioObjectPartInfo{
		PartNumber: objectPart.PartNumber,
		ETag:       objectPart.ETag,
	}, nil
}

func (c *client) CompleteMultipartUpload(uploadId string, parts []*MinioObjectPartInfo) (*MinioObject, error) {
	var completeParts []minio.CompletePart
	for _, part := range parts {
		completeParts = append(completeParts, minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		})
	}

	uploadInfo, err := c.minioCore.CompleteMultipartUpload(context.Background(), c.minioClientCfg.BucketName, "", uploadId, completeParts, minio.PutObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to complete multipart upload %s: %w", uploadId, err)
	}

	return &MinioObject{
		MinioObjectInfo: MinioObjectInfo{
			Key:          uploadInfo.Key,
			Size:         uploadInfo.Size,
			LastModified: uploadInfo.LastModified,
		},
	}, nil
}

func (c *client) AbortMultipartUpload(uploadId string) error {
	err := c.minioCore.AbortMultipartUpload(context.Background(), c.minioClientCfg.BucketName, "", uploadId)
	if err != nil {
		return fmt.Errorf("unable to abort multipart upload %s: %w", uploadId, err)
	}

	return nil
}
