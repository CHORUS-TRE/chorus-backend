package minio

import workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"

var _ MinioClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	return &testClient{}
}

func (c *testClient) StatObject(path string) (*workspace_model.WorkspaceFile, error) {
	return nil, nil
}

func (c *testClient) ListObjects(path string) ([]*workspace_model.WorkspaceFile, error) {
	return nil, nil
}

func (c *testClient) GetObject(objectKey string) (*workspace_model.WorkspaceFile, error) {
	return nil, nil
}

func (c *testClient) PutObject(objectKey string, content []byte) error {
	return nil
}

func (c *testClient) DeleteObject(objectKey string) error {
	return nil
}
