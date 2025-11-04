package minio

import (
	"fmt"

	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

var _ MinioClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	return &testClient{}
}

func (c *testClient) StatWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) ListWorkspaceObjects(workspaceID uint64, path string) ([]*workspace_model.WorkspaceFile, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) GetWorkspaceObject(workspaceID uint64, path string) (*workspace_model.WorkspaceFile, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) PutWorkspaceObject(workspaceID uint64, file *workspace_model.WorkspaceFile) (*workspace_model.WorkspaceFile, error) {
	return nil, fmt.Errorf("Minio Test client not yet implemented")
}

func (c *testClient) DeleteWorkspaceObject(workspaceID uint64, path string) error {
	return fmt.Errorf("Minio Test client not yet implemented")
}
