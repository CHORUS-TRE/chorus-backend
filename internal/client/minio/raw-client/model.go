package miniorawclient

import (
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
