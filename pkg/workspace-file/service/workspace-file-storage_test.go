package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
)

func TestFileLifecycle(t *testing.T) {
	unit.InitTestLogger()

	client := miniorawclient.NewTestClient()
	storage, err := minio.NewMinioFileStorage(client)
	if err != nil {
		t.Fatal(err)
	}
	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	// Define test parameters
	workspaceID := uint64(1)
	globalPath := "/test-client/testfile.txt"
	storePath := storagePathManager.ToStorePath(workspaceID, globalPath)
	content := []byte("Hello, Minio!")

	// Create file
	createdFile, err := storage.CreateFile(context.Background(), &model.File{
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

	client := miniorawclient.NewTestClient()
	storage, err := minio.NewMinioFileStorage(client)
	if err != nil {
		t.Fatal(err)
	}

	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	workspaceID := uint64(1)
	globalPath := "/test-client/existingfile.txt"
	storePath := storagePathManager.ToStorePath(workspaceID, globalPath)
	content := []byte("Existing file content")

	// Create the file first time
	_, err = storage.CreateFile(context.Background(), &model.File{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "initial file creation should not error: %v", err)

	// Attempt to create the same file again
	_, err = storage.CreateFile(context.Background(), &model.File{
		Path:    storePath,
		Content: content,
	})
	assert.Error(t, err, "creating a file that already exists should error")
}

func TestDirectoryLifeCycle(t *testing.T) {
	unit.InitTestLogger()

	client := miniorawclient.NewTestClient()
	storage, err := minio.NewMinioFileStorage(client)
	if err != nil {
		t.Fatal(err)
	}

	storagePathManager, err := NewMinioFileStoragePathManager("test", client, testClientPrefix)
	if err != nil {
		t.Fatal(err)
	}

	workspaceID := uint64(1)
	globalDirPath := "/test-client/mydir/"
	storeDirPath := storagePathManager.ToStorePath(workspaceID, globalDirPath)
	// Create directory
	_, err = storage.CreateFile(context.Background(), &model.File{
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
	_, err = storage.CreateFile(context.Background(), &model.File{
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
