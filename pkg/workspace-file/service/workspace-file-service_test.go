package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio/model"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
)

const (
	testStoreName       = "test"
	testStorePrefix     = "/test-client/"
	testWorkspacePrefix = "workspaces/%s"
)

func createTestService() *WorkspaceFileService {
	client := miniorawclient.NewTestClient()
	fileStore, _ := minio.NewMinioFileStorage(client)

	storeConfigs := map[string]config.WorkspaceFileStore{
		"test": {
			ClientName:      "test",
			StorePrefix:     testStorePrefix,
			WorkspacePrefix: testWorkspacePrefix,
		},
	}

	fileStores := map[string]WorkspaceFileStore{
		"test": fileStore,
	}

	service, _ := NewWorkspaceFileService(fileStores, storeConfigs)
	return service
}

func TestToStorePath(t *testing.T) {
	s := createTestService()
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
			result := s.toStorePath(testStoreName, tt.workspaceID, tt.globalPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromStorePath(t *testing.T) {
	s := createTestService()
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
			result := s.fromStorePath(testStoreName, tt.workspaceID, tt.storePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathRoundTrip(t *testing.T) {
	s := createTestService()

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
			storePath := s.toStorePath(testStoreName, 1, tt.globalPath)
			globalPath := s.fromStorePath(testStoreName, 1, storePath)

			assert.Equal(t, tt.globalPath, globalPath, "round trip conversion should preserve path")
		})
	}
}

func TestFileLifecycle(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	// Define test parameters
	workspaceID := uint64(1)
	globalPath := "/test-client/testfile.txt"
	storePath := s.toStorePath(testStoreName, workspaceID, globalPath)
	content := []byte("Hello, Minio!")

	storage := s.fileStores[testStoreName]

	// Create file
	createdFile, err := storage.CreateFile(context.Background(), &model.File{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "file creation should not error: %v", err)
	assert.Equal(t, storePath, createdFile.Path, "created file path should match store path")

	// Get file metadata
	metadata, err := storage.StatFile(context.Background(), storePath)
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
	_, err = storage.StatFile(context.Background(), storePath)
	assert.Error(t, err, "getting metadata of deleted file should error")
}

func TestCreateFileAlreadyExists(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	workspaceID := uint64(1)
	globalPath := "/test-client/existingfile.txt"
	storePath := s.toStorePath(testStoreName, workspaceID, globalPath)
	content := []byte("Existing file content")
	storage := s.fileStores[testStoreName]

	// Create the file first time
	_, err := storage.CreateFile(context.Background(), &model.File{
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

	s := createTestService()

	storage := s.fileStores[testStoreName]
	workspaceID := uint64(1)
	globalDirPath := "/test-client/mydir/"
	storeDirPath := s.toStorePath(testStoreName, workspaceID, globalDirPath)

	// Create directory
	_, err := storage.CreateDirectory(context.Background(), &model.File{
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
	err = storage.DeleteDirectory(context.Background(), storeDirPath)
	assert.NoError(t, err, "deleting directory should not error: %v", err)

	// Verify deletion of directory
	_, err = storage.StatFile(context.Background(), storeDirPath)
	assert.Error(t, err, "getting metadata of deleted directory should error")

	// Verify deletion of file inside directory
	_, err = storage.StatFile(context.Background(), fileInDirPath)
	assert.Error(t, err, "getting metadata of file inside deleted directory should error")
}

func TestCreateConflictingDirectory(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	storage := s.fileStores[testStoreName]

	workspaceID := uint64(1)
	globalFilePath := "/test-client/conflict"
	storeFilePath := s.toStorePath(testStoreName, workspaceID, globalFilePath)

	// Create a file first
	_, err := storage.CreateFile(context.Background(), &model.File{
		Path:    storeFilePath,
		Content: []byte("This is a file"),
	})
	assert.NoError(t, err, "initial file creation should not error: %v", err)

	// Attempt to create a directory with the same name
	_, err = storage.CreateFile(context.Background(), &model.File{
		Path:        storeFilePath,
		IsDirectory: true,
	})
	assert.Error(t, err, "creating a directory that conflicts with existing file should error")
}

func TestCreateConflictingFile(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	workspaceID := uint64(1)
	globalDirPath := "/test-client/conflictdir/"
	storeDirPath := s.toStorePath(testStoreName, workspaceID, globalDirPath)
	storage := s.fileStores[testStoreName]

	// Create a directory first
	_, err := storage.CreateDirectory(context.Background(), &model.File{
		Path:        storeDirPath,
		IsDirectory: true,
	})
	assert.NoError(t, err, "initial directory creation should not error: %v", err)

	// Attempt to create a file with the same name
	_, err = storage.CreateFile(context.Background(), &model.File{
		Path:        storeDirPath,
		IsDirectory: false,
		Content:     []byte("This is a file"),
	})
	assert.Error(t, err, "creating a file that conflicts with existing directory should error")
}

func TestFileUpload(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	workspaceID := uint64(1)
	globalPath := "/test-client/largefile.txt"
	storePath := s.toStorePath(testStoreName, workspaceID, globalPath)
	fileSize := uint64(10 * 1024 * 1024) // 10 MB
	storage := s.fileStores[testStoreName]

	// Initiate multipart upload
	uploadInfo, err := storage.InitiateMultipartUpload(context.Background(), &model.File{
		Path:        storePath,
		IsDirectory: false,
		Size:        fileSize,
	})
	assert.NoError(t, err, "initiating multipart upload should not error: %v", err)
	assert.NotEmpty(t, uploadInfo.UploadID, "upload ID should not be empty")

	// Upload parts
	partSize := uint64(uploadInfo.PartSize)
	var parts []*model.FilePart
	for partNumber := uint64(1); partNumber <= uploadInfo.TotalParts; partNumber++ {
		partData := make([]byte, partSize)
		if partNumber == uploadInfo.TotalParts {
			lastPartSize := int(fileSize - (partNumber-1)*partSize)
			partData = make([]byte, lastPartSize)
		}
		part, err := storage.UploadPart(context.Background(), uploadInfo.UploadID, &model.FilePart{
			PartNumber: partNumber,
			Data:       partData,
		})
		assert.NoError(t, err, "uploading part %d should not error: %v", partNumber, err)
		assert.NotEmpty(t, part.ETag, "part should have an ETag")
		parts = append(parts, part)
	}

	// Complete multipart upload
	uploadedFile, err := storage.CompleteMultipartUpload(context.Background(), &model.File{
		Path:        storePath,
		IsDirectory: false,
		Size:        fileSize,
	}, uploadInfo.UploadID, parts)
	assert.NoError(t, err, "completing multipart upload should not error: %v", err)
	assert.Equal(t, storePath, uploadedFile.Path, "uploaded file path should match")

	// Retrieve and verify uploaded file
	_, err = storage.StatFile(context.Background(), storePath)
	assert.NoError(t, err, "retrieving uploaded file should not error: %v", err)
}

func TestAbortFileUpload(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	workspaceID := uint64(1)
	globalPath := "/test-client/abortfile.txt"
	storePath := s.toStorePath(testStoreName, workspaceID, globalPath)
	storage := s.fileStores[testStoreName]

	// Initiate multipart upload
	uploadInfo, err := storage.InitiateMultipartUpload(context.Background(), &model.File{
		Path:        storePath,
		IsDirectory: false,
		Size:        5 * 1024 * 1024, // 5 MB file
	})
	assert.NoError(t, err, "initiating multipart upload should not error: %v", err)
	assert.NotEmpty(t, uploadInfo.UploadID, "upload ID should not be empty")

	// Abort multipart upload
	err = storage.AbortMultipartUpload(context.Background(), uploadInfo.UploadID)
	assert.NoError(t, err, "aborting multipart upload should not error: %v", err)
}
