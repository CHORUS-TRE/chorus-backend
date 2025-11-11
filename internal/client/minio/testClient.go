package minio

import (
	"fmt"
)

var _ MinioClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	return &testClient{}
}

func (c *testClient) GetClientName() string {
	return "test-minio-client"
}

func (c *testClient) GetClientPrefix() string {
	return "/test-client/"
}

func (c *testClient) GetObject(objectKey string) (*MinioObject, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) StatObject(objectKey string) (*MinioObjectInfo, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) ListObjects(objectKey string) ([]*MinioObjectInfo, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) PutObject(objectKey string, object *MinioObject) (*MinioObjectInfo, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) DeleteObject(objectKey string) error {
	return fmt.Errorf("Minio Test client not yet implemented")
}
