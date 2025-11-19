package minio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
)

func TestNormalizePath(t *testing.T) {
	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
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
			result := storage.NormalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToStorePath(t *testing.T) {
	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
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
			result := storage.ToStorePath(tt.workspaceID, tt.globalPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromStorePath(t *testing.T) {
	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
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
			result := storage.FromStorePath(tt.workspaceID, tt.storePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathRoundTrip(t *testing.T) {
	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
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
			storePath := storage.ToStorePath(1, tt.globalPath)
			globalPath := storage.FromStorePath(1, storePath)

			// Normalize both for comparison (handle trailing slashes)
			expected := storage.NormalizePath(tt.globalPath)
			actual := storage.NormalizePath(globalPath)

			assert.Equal(t, expected, actual, "round trip conversion should preserve path")
		})
	}
}

func TestFileLifecycle(t *testing.T) {
	unit.InitTestLogger()

	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
	if err != nil {
		t.Fatal(err)
	}

	// Define test parameters
	workspaceID := uint64(1)
	globalPath := "/test-client/testfile.txt"
	storePath := storage.ToStorePath(workspaceID, globalPath)
	content := []byte("Hello, Minio!")

	// Create file
	createdFile, err := storage.CreateFile(context.Background(), &model.WorkspaceFile{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "file creation should not error: %v", err)
	assert.Equal(t, storePath, createdFile.Path, "created file path should match store path")

	// Get file metadata
	metadata, err := storage.GetFileMetadata(context.Background(), storePath)
	assert.NoError(t, err, "getting file metadata should not error: %v", err)
	assert.Equal(t, int64(len(content)), metadata.Size, "file size should match content length")

	// Get file content
	retrievedFile, err := storage.GetFile(context.Background(), storePath)
	assert.NoError(t, err, "getting file content should not error: %v", err)
	assert.Equal(t, content, retrievedFile.Content, "file content should match original content")

	// List files in directory
	files, err := storage.ListFiles(context.Background(), "workspaces/workspace1/")
	assert.NoError(t, err, "listing files should not error: %v", err)
	found := false
	for _, f := range files {
		if f.Path == storePath {
			found = true
			break
		}
	}
	assert.True(t, found, "created file should be listed in directory")

	// Delete file
	err = storage.DeleteFile(context.Background(), storePath)
	assert.NoError(t, err, "deleting file should not error: %v", err)

	// Verify deletion
	_, err = storage.GetFileMetadata(context.Background(), storePath)
	assert.Error(t, err, "getting metadata of deleted file should error")
}

func TestCreateFileAlreadyExists(t *testing.T) {
	unit.InitTestLogger()

	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
	if err != nil {
		t.Fatal(err)
	}

	workspaceID := uint64(1)
	globalPath := "/test-client/existingfile.txt"
	storePath := storage.ToStorePath(workspaceID, globalPath)
	content := []byte("Existing file content")

	// Create the file first time
	_, err = storage.CreateFile(context.Background(), &model.WorkspaceFile{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "initial file creation should not error: %v", err)

	// Attempt to create the same file again
	_, err = storage.CreateFile(context.Background(), &model.WorkspaceFile{
		Path:    storePath,
		Content: content,
	})
	assert.Error(t, err, "creating a file that already exists should error")
}

func TestDirectoryLifeCycle(t *testing.T) {
	unit.InitTestLogger()

	client := minio.NewTestClient()
	storage, err := NewMinioFileStorage("test", client)
	if err != nil {
		t.Fatal(err)
	}

	workspaceID := uint64(1)
	globalDirPath := "/test-client/mydir/"
	storeDirPath := storage.ToStorePath(workspaceID, globalDirPath)

	// Create directory
	_, err = storage.CreateFile(context.Background(), &model.WorkspaceFile{
		Path:        storeDirPath,
		IsDirectory: true,
	})
	assert.NoError(t, err, "directory creation should not error: %v", err)

	// List files in root to verify directory exists
	files, err := storage.ListFiles(context.Background(), "workspaces/workspace1/")
	assert.NoError(t, err, "listing files should not error: %v", err)

	found := false
	for _, f := range files {
		if f.Path == storeDirPath && f.IsDirectory {
			found = true
			break
		}
	}
	assert.True(t, found, "created directory should be listed in root")

	// Create a file inside the directory
	fileInDirPath := storeDirPath + "file.txt"
	_, err = storage.CreateFile(context.Background(), &model.WorkspaceFile{
		Path:    fileInDirPath,
		Content: []byte("File inside directory"),
	})

	assert.NoError(t, err, "file creation inside directory should not error: %v", err)

	// List files in the directory
	dirFiles, err := storage.ListFiles(context.Background(), storeDirPath)
	assert.NoError(t, err, "listing files in directory should not error: %v", err)

	foundFile := false
	for _, f := range dirFiles {
		if f.Path == fileInDirPath {
			foundFile = true
			break
		}
	}
	assert.True(t, foundFile, "file inside directory should be listed")

	// Delete the directory
	err = storage.DeleteFile(context.Background(), storeDirPath)
	assert.NoError(t, err, "deleting directory should not error: %v", err)

	// Verify deletion of directory
	_, err = storage.GetFileMetadata(context.Background(), storeDirPath)
	assert.Error(t, err, "getting metadata of deleted directory should error")

	// Verify deletion of file inside directory
	_, err = storage.GetFileMetadata(context.Background(), fileInDirPath)
	assert.Error(t, err, "getting metadata of file inside deleted directory should error")
}
