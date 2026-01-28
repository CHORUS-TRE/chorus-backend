package k8s

var _ K8sClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	c := &testClient{}
	return c
}

// Workspace (Namespace) operations
func (c *testClient) CreateWorkspace(tenantID uint64, namespace string) error {
	return nil
}

func (c *testClient) DeleteWorkspace(namespace string) error {
	return nil
}

// Workbench operations
func (c *testClient) CreateWorkbench(workbench Workbench) error {
	return nil
}

func (c *testClient) UpdateWorkbench(workbench Workbench) error {
	return nil
}

func (c *testClient) DeleteWorkbench(namespace, workbenchName string) error {
	return nil
}

// AppInstance operations
func (c *testClient) CreateAppInstance(namespace, workbenchName string, app AppInstance) error {
	return nil
}

func (c *testClient) UpdateAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	return nil
}

func (c *testClient) DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	return nil
}

// Utility operations
func (c *testClient) PrePullImageOnAllNodes(image string) {}

func (c *testClient) CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error) {
	return 0, nil, nil
}

// Watchers registration methods
func (c *testClient) RegisterOnNewWorkbenchHandler(func(workbench Workbench) error) error {
	return nil
}

func (c *testClient) RegisterOnUpdateWorkbenchHandler(func(workbench Workbench) error) error {
	return nil
}

func (c *testClient) RegisterOnDeleteWorkbenchHandler(func(workbench Workbench) error) error {
	return nil
}
