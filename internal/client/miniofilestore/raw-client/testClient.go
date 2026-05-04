package miniorawclient

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"
)

var _ MinioClienter = &testClient{}

type testClient struct {
	objects   map[string]*MinioObject
	uploads   map[string][]*MinioObjectPartInfo
	partData  map[string]map[int][]byte
	mutex     sync.RWMutex
}

func NewTestClient() *testClient {
	return &testClient{
		objects:  make(map[string]*MinioObject),
		uploads:  make(map[string][]*MinioObjectPartInfo),
		partData: make(map[string]map[int][]byte),
	}
}

func (c *testClient) GetClientConfig() MinioClientConfig {
	return MinioClientConfig{
		Name:                   "test-client",
		Endpoint:               "test-endpoint",
		AccessKeyID:            "test-access-key",
		SecretAccessKey:        "test-secret-key",
		UseSSL:                 false,
		BucketName:             "test-bucket",
		MultipartMinPartSize:   5 * 1024 * 1024,        // 5MB
		MultipartMaxPartSize:   5 * 1024 * 1024 * 1024, // 5GB
		MultipartMaxTotalParts: 10000,
	}
}

func (c *testClient) Ping() error {
	return nil
}

func (c *testClient) GetObjectStream(objectKey string) (io.ReadCloser, *MinioObjectInfo, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	object, ok := c.objects[objectKey]
	if !ok {
		return nil, nil, fmt.Errorf("object not found: %s", objectKey)
	}
	return io.NopCloser(bytes.NewReader(object.Content)), &object.MinioObjectInfo, nil
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
					if len(parts) == 1 {
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
	object.Size = uint64(len(object.Content))
	c.objects[objectKey] = object
	return &object.MinioObjectInfo, nil
}

func (c *testClient) MoveObject(oldObjectKey string, newObjectKey string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	object, ok := c.objects[oldObjectKey]
	if !ok {
		return fmt.Errorf("object not found: %s", oldObjectKey)
	}
	object.Key = newObjectKey
	c.objects[newObjectKey] = object
	delete(c.objects, oldObjectKey)
	return nil
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

func (c *testClient) NewMultipartUpload(objectKey string, objectSize uint64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	uploadID := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	c.uploads[uploadID] = []*MinioObjectPartInfo{}
	c.partData[uploadID] = make(map[int][]byte)
	return uploadID, nil
}

func (c *testClient) PutObjectPart(objectKey string, uploadId string, partNumber int, data []byte) (*MinioObjectPartInfo, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	parts, ok := c.uploads[uploadId]
	if !ok {
		return nil, fmt.Errorf("upload ID not found: %s", uploadId)
	}

	partInfo := &MinioObjectPartInfo{
		PartNumber: partNumber,
		ETag:       fmt.Sprintf("etag-%s-part-%d", uploadId, partNumber),
	}
	c.uploads[uploadId] = append(parts, partInfo)

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	c.partData[uploadId][partNumber] = dataCopy

	return &MinioObjectPartInfo{
		PartNumber: partInfo.PartNumber,
		ETag:       partInfo.ETag,
	}, nil
}

func (c *testClient) CompleteMultipartUpload(objectKey string, uploadId string, parts []*MinioObjectPartInfo) (*MinioObject, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.uploads[uploadId]
	if !ok {
		return nil, fmt.Errorf("upload ID not found: %s", uploadId)
	}

	var content []byte
	for _, part := range parts {
		content = append(content, c.partData[uploadId][part.PartNumber]...)
	}

	object := &MinioObject{
		MinioObjectInfo: MinioObjectInfo{
			Key:          objectKey,
			Size:         uint64(len(content)),
			LastModified: time.Now(),
		},
		Content: content,
	}

	c.objects[object.Key] = object
	delete(c.uploads, uploadId)
	delete(c.partData, uploadId)
	return object, nil
}

func (c *testClient) AbortMultipartUpload(objectKey string, uploadId string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.uploads[uploadId]
	if !ok {
		return fmt.Errorf("upload ID not found: %s", uploadId)
	}
	delete(c.uploads, uploadId)
	delete(c.partData, uploadId)
	return nil
}
