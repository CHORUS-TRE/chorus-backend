package model

import (
	"time"
)

type WorkspaceFile struct {
	Path string

	Name        string
	IsDirectory bool
	Size        int64
	MimeType    string

	UpdatedAt time.Time

	Content []byte
}
