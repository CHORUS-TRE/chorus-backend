package service

import (
	"testing"

	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/stretchr/testify/assert"
)

const testClientPrefix = "/test-client/"

func TestNormalizePath(t *testing.T) {
	client := miniorawclient.NewTestClient()
	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path with leading slash",
			input:    "/archive/file.txt",
			expected: "/archive/file.txt",
		},
		{
			name:     "path without leading slash",
			input:    "archive/file.txt",
			expected: "/archive/file.txt",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "/",
		},
		{
			name:     "nested path",
			input:    "/folder/subfolder/file.txt",
			expected: "/folder/subfolder/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storagePathManager.NormalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToStorePath(t *testing.T) {
	client := miniorawclient.NewTestClient()
	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		workspaceID uint64
		globalPath  string
		expected    string
	}{
		{
			name:        "strips prefix and adds workspace scope",
			workspaceID: 1,
			globalPath:  "/test-client/file.txt",
			expected:    "workspaces/workspace1/file.txt",
		},
		{
			name:        "handles nested paths",
			workspaceID: 1,
			globalPath:  "/test-client/folder/file.txt",
			expected:    "workspaces/workspace1/folder/file.txt",
		},
		{
			name:        "handles path without leading slash",
			workspaceID: 1,
			globalPath:  "test-client/file.txt",
			expected:    "workspaces/workspace1/file.txt",
		},
		{
			name:        "handles root of store",
			workspaceID: 1,
			globalPath:  "/test-client/",
			expected:    "workspaces/workspace1/",
		},
		{
			name:        "different workspace ID",
			workspaceID: 42,
			globalPath:  "/test-client/data.json",
			expected:    "workspaces/workspace42/data.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storagePathManager.ToStorePath(tt.workspaceID, tt.globalPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromStorePath(t *testing.T) {
	client := miniorawclient.NewTestClient()
	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		workspaceID uint64
		storePath   string
		expected    string
	}{
		{
			name:        "removes workspace scope and adds prefix",
			workspaceID: 1,
			storePath:   "workspaces/workspace1/file.txt",
			expected:    "/test-client/file.txt",
		},
		{
			name:        "handles nested paths with workspace scope",
			workspaceID: 1,
			storePath:   "workspaces/workspace1/folder/file.txt",
			expected:    "/test-client/folder/file.txt",
		},
		{
			name:        "handles workspace root",
			workspaceID: 1,
			storePath:   "workspaces/workspace1/",
			expected:    "/test-client/",
		},
		{
			name:        "handles different workspace ID",
			workspaceID: 42,
			storePath:   "workspaces/workspace42/data.json",
			expected:    "/test-client/data.json",
		},
		{
			name:        "handles path without workspace scope",
			workspaceID: 1,
			storePath:   "file.txt",
			expected:    "/test-client/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storagePathManager.FromStorePath(tt.workspaceID, tt.storePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathRoundTrip(t *testing.T) {
	client := miniorawclient.NewTestClient()
	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		globalPath string
	}{
		{
			name:       "simple file",
			globalPath: "/test-client/file.txt",
		},
		{
			name:       "nested file",
			globalPath: "/test-client/folder/subfolder/file.txt",
		},
		{
			name:       "directory",
			globalPath: "/test-client/folder/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to store path and back
			storePath := storagePathManager.ToStorePath(1, tt.globalPath)
			globalPath := storagePathManager.FromStorePath(1, storePath)

			// Normalize both for comparison (handle trailing slashes)
			expected := storagePathManager.NormalizePath(tt.globalPath)
			actual := storagePathManager.NormalizePath(globalPath)

			assert.Equal(t, expected, actual, "round trip conversion should preserve path")
		})
	}
}
