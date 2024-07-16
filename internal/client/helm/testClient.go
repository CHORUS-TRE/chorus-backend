package helm

var _ HelmClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	c := &testClient{}
	return c
}

func (c *testClient) CreateWorkbench(namespace, workbenchName string) error {
	return nil
}

func (c *testClient) CreateAppInstance(namespace, workbenchName, appName, appImage string) error {
	return nil
}

func (c *testClient) DeleteApp(namespace, workbenchName, appName string) error {
	return nil
}

func (c *testClient) DeleteWorkbench(namespace, workbenchName string) error {
	return nil
}
