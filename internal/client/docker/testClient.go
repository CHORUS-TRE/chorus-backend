package docker

var _ DockerClienter = &testClient{}

type testClient struct{}

func NewTestClient() *testClient {
	return &testClient{}
}

// VerifyImageExists always returns true for testing purposes
func (c *testClient) ImageExists(imageRef string, username string, password string) (bool, error) {
	// In test mode, we assume all images exist
	return true, nil
}
