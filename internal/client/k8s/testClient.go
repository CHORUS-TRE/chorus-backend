package k8s

var _ K8sClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	c := &testClient{}
	return c
}

func (c *testClient) WatchOnNewWorkbench(func(namespace, workbenchName string, tenantID uint64, apps []AppInstance) error) error {
	return nil
}

func (c *testClient) WatchOnUpdateWorkbench(func(namespace, workbenchName string, tenantID uint64, apps []AppInstance) error) error {
	return nil
}

func (c *testClient) WatchOnDeleteWorkbench(func(namespace, workbenchName string, tenantID uint64, apps []AppInstance) error) error {
	return nil
}

func (c *testClient) CreateWorkspace(tenantID uint64, namespace string) error {
	return nil
}

func (c *testClient) DeleteWorkspace(namespace string) error {
	return nil
}

func (c *testClient) CreateWorkbench(tenantID uint64, req MakeWorkbenchRequest) error {
	return nil
}

func (c *testClient) UpdateWorkbench(tenantID uint64, req MakeWorkbenchRequest) error {
	return nil
}

func (c *testClient) CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error) {
	return 0, nil, nil
}

func (c *testClient) CreateAppInstance(namespace, workbenchName string, app AppInstance) error {
	return nil
}

func (c *testClient) DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	return nil
}

func (c *testClient) DeleteWorkbench(namespace, workbenchName string) error {
	return nil
}
