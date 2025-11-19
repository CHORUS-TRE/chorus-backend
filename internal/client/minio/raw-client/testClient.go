package miniorawclient

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"
)

var _ MinioClienter = &testClient{}

type testClient struct {
	objects map[string]*MinioObject
	mutex   sync.RWMutex
}

func NewTestClient() *testClient {
	return &testClient{
		objects: make(map[string]*MinioObject),
	}
}

func (c *testClient) GetClientName() string {
	return "test-minio-client"
}

func (c *testClient) GetObject(objectKey string) (*MinioObject, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if object, ok := c.objects[objectKey]; ok {
		return object, nil
	}
	return nil, fmt.Errorf("object not found: %s", objectKey)
}

func (c *testClient) StatObject(objectKey string) (*MinioObjectInfo, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if object, ok := c.objects[objectKey]; ok {
		return &object.MinioObjectInfo, nil
	}
	return nil, fmt.Errorf("object not found: %s", objectKey)
}

func (c *testClient) ListObjects(prefix string, recursive bool) ([]*MinioObjectInfo, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var results []*MinioObjectInfo
	seen := make(map[string]bool)

	for key, object := range c.objects {
		if strings.HasPrefix(key, prefix) {
			if recursive {
				results = append(results, &object.MinioObjectInfo)
			} else {
				subPath := strings.TrimPrefix(key, prefix)
				parts := strings.Split(subPath, "/")
				if len(parts) > 0 {
					// It's a file or a directory in the current level
					if len(parts) == 1 && parts[0] != "" {
						results = append(results, &object.MinioObjectInfo)
					} else if len(parts) > 1 {
						// It's a subdirectory, add it once
						dirPath := path.Join(prefix, parts[0]) + "/"
						if !seen[dirPath] {
							results = append(results, &MinioObjectInfo{
								Key:          dirPath,
								LastModified: object.LastModified, // Or some other logic
							})
							seen[dirPath] = true
						}
					}
				}
			}
		}
	}
	return results, nil
}

func (c *testClient) PutObject(objectKey string, object *MinioObject) (*MinioObjectInfo, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	object.LastModified = time.Now()
	object.Key = objectKey
	object.Size = int64(len(object.Content))
	c.objects[objectKey] = object
	return &object.MinioObjectInfo, nil
}

func (c *testClient) DeleteObject(objectKey string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.objects[objectKey]; !ok {
		return fmt.Errorf("object not found: %s", objectKey)
	}
	delete(c.objects, objectKey)
	return nil
}
