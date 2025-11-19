package miniorawclient

import (
	"strings"
	"time"
)

type MinioObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	MimeType     string
}

type MinioObject struct {
	MinioObjectInfo
	Content []byte
}

func NormalizePrefix(prefix string) string {
	normalizedPrefix := "/" + strings.TrimPrefix(prefix, "/")
	normalizedPrefix = strings.TrimSuffix(normalizedPrefix, "/") + "/"

	return normalizedPrefix
}
