package service

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/miniofilestore"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/miniofilestore/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
)

const (
	testStoreName       = "test-client"
	testStoreName2      = "test-client-2"
	testWorkspacePrefix = "workspaces/%s"
)

func createTestService() *WorkspaceFileService {
	client := miniorawclient.NewTestClient()
	fileStore, _ := miniofilestore.NewMinioFileStorage(client)

	cfg := config.Config{}
	cfg.Services.WorkspaceFileService.Stores = map[string]config.WorkspaceFileStore{
		testStoreName: {
			FileStoreName:   testStoreName,
			WorkspacePrefix: testWorkspacePrefix,
		},
	}
	cfg.Storage.FileStores = map[string]config.FileStore{
		testStoreName: {
			Type: "minio",
			MinioConfig: config.FileStoreMinioConfig{
				Enabled: true,
			},
		},
	}

	fileStores := map[string]filestore.FileStore{
		testStoreName: fileStore,
	}

	service, _ := NewWorkspaceFileService(cfg, fileStores)
	return service
}

func createTestServiceWithTwoStores() *WorkspaceFileService {
	client1 := miniorawclient.NewTestClient()
	fileStore1, _ := miniofilestore.NewMinioFileStorage(client1)

	client2 := miniorawclient.NewTestClient()
	fileStore2, _ := miniofilestore.NewMinioFileStorage(client2)

	cfg := config.Config{}
	cfg.Services.WorkspaceFileService.Stores = map[string]config.WorkspaceFileStore{
		testStoreName: {
			FileStoreName:   testStoreName,
			WorkspacePrefix: testWorkspacePrefix,
		},
		testStoreName2: {
			FileStoreName:   testStoreName2,
			WorkspacePrefix: testWorkspacePrefix,
		},
	}
	cfg.Storage.FileStores = map[string]config.FileStore{
		testStoreName: {
			Type:        "minio",
			MinioConfig: config.FileStoreMinioConfig{Enabled: true},
		},
		testStoreName2: {
			Type:        "minio",
			MinioConfig: config.FileStoreMinioConfig{Enabled: true},
		},
	}

	service, _ := NewWorkspaceFileService(cfg, map[string]filestore.FileStore{
		testStoreName:  fileStore1,
		testStoreName2: fileStore2,
	})
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
			name:        "strips store name and adds workspace scope",
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
		{
			name:        "strips double slash after store name",
			workspaceID: 1,
			globalPath:  "/test-client//nested/file.txt",
			expected:    "workspaces/workspace1/nested/file.txt",
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

	storage := s.stores[testStoreName].store

	// Create file
	createdFile, err := storage.CreateFile(context.Background(), &filestore.File{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "file creation should not error: %v", err)
	assert.Equal(t, storePath, createdFile.Path, "created file path should match store path")

	// Get file metadata
	metadata, err := storage.StatFile(context.Background(), storePath)
	assert.NoError(t, err, "getting file metadata should not error: %v", err)
	assert.Equal(t, uint64(len(content)), metadata.Size, "file size should match content length")

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
	storage := s.stores[testStoreName].store

	// Create the file first time
	_, err := storage.CreateFile(context.Background(), &filestore.File{
		Path:    storePath,
		Content: content,
	})
	assert.NoError(t, err, "initial file creation should not error: %v", err)

	// Attempt to create the same file again
	_, err = storage.CreateFile(context.Background(), &filestore.File{
		Path:    storePath,
		Content: content,
	})
	assert.Error(t, err, "creating a file that already exists should error")
}

func TestDirectoryLifeCycle(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()

	storage := s.stores[testStoreName].store
	workspaceID := uint64(1)
	globalDirPath := "/test-client/mydir/"
	storeDirPath := s.toStorePath(testStoreName, workspaceID, globalDirPath)

	// Create directory
	_, err := storage.CreateDirectory(context.Background(), &filestore.File{
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
	_, err = storage.CreateFile(context.Background(), &filestore.File{
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

	storage := s.stores[testStoreName].store

	workspaceID := uint64(1)
	globalFilePath := "/test-client/conflict"
	storeFilePath := s.toStorePath(testStoreName, workspaceID, globalFilePath)

	// Create a file first
	_, err := storage.CreateFile(context.Background(), &filestore.File{
		Path:    storeFilePath,
		Content: []byte("This is a file"),
	})
	assert.NoError(t, err, "initial file creation should not error: %v", err)

	// Attempt to create a directory with the same name
	_, err = storage.CreateFile(context.Background(), &filestore.File{
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
	storage := s.stores[testStoreName].store

	// Create a directory first
	_, err := storage.CreateDirectory(context.Background(), &filestore.File{
		Path:        storeDirPath,
		IsDirectory: true,
	})
	assert.NoError(t, err, "initial directory creation should not error: %v", err)

	// Attempt to create a file with the same name
	_, err = storage.CreateFile(context.Background(), &filestore.File{
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
	storage := s.stores[testStoreName].store

	// Initiate multipart upload
	uploadInfo, err := storage.InitiateMultipartUpload(context.Background(), &filestore.File{
		Path:        storePath,
		IsDirectory: false,
		Size:        fileSize,
	})
	assert.NoError(t, err, "initiating multipart upload should not error: %v", err)
	assert.NotEmpty(t, uploadInfo.UploadID, "upload ID should not be empty")

	// Upload parts
	partSize := uint64(uploadInfo.PartSize)
	var parts []*filestore.FilePart
	for partNumber := uint64(1); partNumber <= uploadInfo.TotalParts; partNumber++ {
		partData := make([]byte, partSize)
		if partNumber == uploadInfo.TotalParts {
			lastPartSize := int(fileSize - (partNumber-1)*partSize)
			partData = make([]byte, lastPartSize)
		}
		part, err := storage.UploadPart(context.Background(), storePath, uploadInfo.UploadID, &filestore.FilePart{
			PartNumber: partNumber,
			Data:       partData,
		})
		assert.NoError(t, err, "uploading part %d should not error: %v", partNumber, err)
		assert.NotEmpty(t, part.ETag, "part should have an ETag")
		parts = append(parts, part)
	}

	// Complete multipart upload
	uploadedFile, err := storage.CompleteMultipartUpload(context.Background(), storePath, uploadInfo.UploadID, parts)
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
	storage := s.stores[testStoreName].store

	// Initiate multipart upload
	uploadInfo, err := storage.InitiateMultipartUpload(context.Background(), &filestore.File{
		Path:        storePath,
		IsDirectory: false,
		Size:        5 * 1024 * 1024, // 5 MB file
	})
	assert.NoError(t, err, "initiating multipart upload should not error: %v", err)
	assert.NotEmpty(t, uploadInfo.UploadID, "upload ID should not be empty")

	// Abort multipart upload
	err = storage.AbortMultipartUpload(context.Background(), storePath, uploadInfo.UploadID)
	assert.NoError(t, err, "aborting multipart upload should not error: %v", err)
}

func TestGetFileStream(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()
	storage := s.stores[testStoreName].store
	storePath := s.toStorePath(testStoreName, 1, "/test-client/stream.txt")
	content := []byte("streaming content for test")

	_, err := storage.CreateFile(context.Background(), &filestore.File{Path: storePath, Content: content})
	require.NoError(t, err)

	reader, meta, err := storage.GetFileStream(context.Background(), storePath)
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, content, data)
	assert.Equal(t, uint64(len(content)), meta.Size)
}

func TestUpdateWorkspaceFile_SameStoreMove(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()
	storage := s.stores[testStoreName].store
	workspaceID := uint64(1)
	sourcePath := "/test-client/move-src.txt"
	destPath := "/test-client/move-dst.txt"

	_, err := storage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName, workspaceID, sourcePath),
		Content: []byte("move me"),
	})
	require.NoError(t, err)

	result, err := s.UpdateWorkspaceFile(context.Background(), workspaceID, sourcePath, &filestore.File{
		Path: destPath,
		Name: "move-dst.txt",
	}, false)
	require.NoError(t, err)
	assert.Equal(t, destPath, result.Path)

	_, err = storage.StatFile(context.Background(), s.toStorePath(testStoreName, workspaceID, destPath))
	assert.NoError(t, err, "destination should exist after move")

	_, err = storage.StatFile(context.Background(), s.toStorePath(testStoreName, workspaceID, sourcePath))
	assert.Error(t, err, "source should be gone after move")
}

func TestUpdateWorkspaceFile_SameStoreCopy(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()
	storage := s.stores[testStoreName].store
	workspaceID := uint64(1)
	sourcePath := "/test-client/copy-src.txt"
	destPath := "/test-client/copy-dst.txt"
	content := []byte("copy me")

	_, err := storage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName, workspaceID, sourcePath),
		Content: content,
	})
	require.NoError(t, err)

	result, err := s.UpdateWorkspaceFile(context.Background(), workspaceID, sourcePath, &filestore.File{
		Path: destPath,
		Name: "copy-dst.txt",
	}, true)
	require.NoError(t, err)
	assert.Equal(t, destPath, result.Path)

	_, err = storage.StatFile(context.Background(), s.toStorePath(testStoreName, workspaceID, sourcePath))
	assert.NoError(t, err, "source should be preserved after copy")

	dstFile, err := storage.GetFile(context.Background(), s.toStorePath(testStoreName, workspaceID, destPath))
	require.NoError(t, err)
	assert.Equal(t, content, dstFile.Content, "destination content should match source")
}

func TestUpdateWorkspaceFile_CrossStoreMove(t *testing.T) {
	unit.InitTestLogger()

	s := createTestServiceWithTwoStores()
	srcStorage := s.stores[testStoreName].store
	dstStorage := s.stores[testStoreName2].store
	workspaceID := uint64(1)
	sourcePath := "/" + testStoreName + "/xmove-src.txt"
	destPath := "/" + testStoreName2 + "/xmove-dst.txt"
	content := []byte("cross-store move content")

	_, err := srcStorage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName, workspaceID, sourcePath),
		Content: content,
	})
	require.NoError(t, err)

	result, err := s.UpdateWorkspaceFile(context.Background(), workspaceID, sourcePath, &filestore.File{
		Path: destPath,
		Name: "xmove-dst.txt",
	}, false)
	require.NoError(t, err)
	assert.Equal(t, destPath, result.Path)

	_, err = srcStorage.StatFile(context.Background(), s.toStorePath(testStoreName, workspaceID, sourcePath))
	assert.Error(t, err, "source should be gone after cross-store move")

	dstFile, err := dstStorage.GetFile(context.Background(), s.toStorePath(testStoreName2, workspaceID, destPath))
	require.NoError(t, err)
	assert.Equal(t, content, dstFile.Content, "destination content should match source")
}

func TestUpdateWorkspaceFile_CrossStoreCopy(t *testing.T) {
	unit.InitTestLogger()

	s := createTestServiceWithTwoStores()
	srcStorage := s.stores[testStoreName].store
	dstStorage := s.stores[testStoreName2].store
	workspaceID := uint64(1)
	sourcePath := "/" + testStoreName + "/xcopy-src.txt"
	destPath := "/" + testStoreName2 + "/xcopy-dst.txt"
	content := []byte("cross-store copy content")

	_, err := srcStorage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName, workspaceID, sourcePath),
		Content: content,
	})
	require.NoError(t, err)

	result, err := s.UpdateWorkspaceFile(context.Background(), workspaceID, sourcePath, &filestore.File{
		Path: destPath,
		Name: "xcopy-dst.txt",
	}, true)
	require.NoError(t, err)
	assert.Equal(t, destPath, result.Path)

	_, err = srcStorage.StatFile(context.Background(), s.toStorePath(testStoreName, workspaceID, sourcePath))
	assert.NoError(t, err, "source should be preserved after cross-store copy")

	dstFile, err := dstStorage.GetFile(context.Background(), s.toStorePath(testStoreName2, workspaceID, destPath))
	require.NoError(t, err)
	assert.Equal(t, content, dstFile.Content, "destination content should match source")
}

func TestUpdateWorkspaceFile_FailsIfDestinationExists(t *testing.T) {
	unit.InitTestLogger()

	s := createTestServiceWithTwoStores()
	srcStorage := s.stores[testStoreName].store
	dstStorage := s.stores[testStoreName2].store
	workspaceID := uint64(1)
	sourcePath := "/" + testStoreName + "/overwrite-src.txt"
	destPath := "/" + testStoreName2 + "/overwrite-dst.txt"
	originalContent := []byte("old content")

	_, err := srcStorage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName, workspaceID, sourcePath),
		Content: []byte("new content"),
	})
	require.NoError(t, err)

	_, err = dstStorage.CreateFile(context.Background(), &filestore.File{
		Path:    s.toStorePath(testStoreName2, workspaceID, destPath),
		Content: originalContent,
	})
	require.NoError(t, err)

	_, err = s.UpdateWorkspaceFile(context.Background(), workspaceID, sourcePath, &filestore.File{
		Path: destPath,
		Name: "overwrite-dst.txt",
	}, true)
	assert.Error(t, err, "should fail when destination already exists")

	dstFile, err := dstStorage.GetFile(context.Background(), s.toStorePath(testStoreName2, workspaceID, destPath))
	require.NoError(t, err)
	assert.Equal(t, originalContent, dstFile.Content, "destination content should be unchanged")
}

func TestUpdateWorkspaceFile_FailsOnMissingSource(t *testing.T) {
	unit.InitTestLogger()

	s := createTestServiceWithTwoStores()
	dstStorage := s.stores[testStoreName2].store
	workspaceID := uint64(1)

	_, err := s.UpdateWorkspaceFile(context.Background(), workspaceID,
		"/"+testStoreName+"/missing.txt",
		&filestore.File{Path: "/" + testStoreName2 + "/abort-dst.txt", Name: "abort-dst.txt"},
		false,
	)
	assert.Error(t, err)

	_, err = dstStorage.StatFile(context.Background(), s.toStorePath(testStoreName2, workspaceID, "/"+testStoreName2+"/abort-dst.txt"))
	assert.Error(t, err, "destination should not exist when source was not found")
}

func TestFileUploadPartSizeCalculation(t *testing.T) {
	unit.InitTestLogger()

	s := createTestService()
	storage := s.stores[testStoreName].store

	tests := []struct {
		name               string
		fileSize           uint64
		expectedPartSize   uint64
		expectedTotalParts uint64
		expectError        bool
	}{
		{
			name:        "zero file",
			fileSize:    0, // 0 bytes
			expectError: true,
		},
		{
			name:               "tiny file",
			fileSize:           1, // 1 byte
			expectedPartSize:   1,
			expectedTotalParts: 1,
		},
		{
			name:               "single part file",
			fileSize:           5 * 1024 * 1024, // 5 MB
			expectedPartSize:   5 * 1024 * 1024,
			expectedTotalParts: 1,
		},
		{
			name:               "slightly over single part file",
			fileSize:           5*1024*1024 + 1, // > 5 MB
			expectedPartSize:   5 * 1024 * 1024, // single part
			expectedTotalParts: 2,
		},
		{
			name:               "medium file",
			fileSize:           500 * 1024 * 1024, // 500 MB
			expectedPartSize:   5 * 1024 * 1024,
			expectedTotalParts: 100,
		},
		{
			name:               "large file (10GB)",
			fileSize:           10 * 1024 * 1024 * 1024, // 10 GB
			expectedPartSize:   5 * 1024 * 1024,
			expectedTotalParts: 2048,
		},
		{
			name:               "huge file (100GB)",
			fileSize:           100 * 1024 * 1024 * 1024, // 100 GB
			expectedPartSize:   10737419,                 // ~10.24 MB (100GB/10000)
			expectedTotalParts: 10000,
		},
		{
			name:        "exceeds max parts",
			fileSize:    60000 * 1024 * 1024 * 1024, // 60 TB
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploadInfo, err := storage.InitiateMultipartUpload(context.Background(), &filestore.File{
				Path:        "/test-client/testfile.txt",
				IsDirectory: false,
				Size:        tt.fileSize,
			})
			if tt.expectError {
				assert.Error(t, err, "expected error but got none")
				return
			} else {
				assert.Equal(t, tt.expectedPartSize, uploadInfo.PartSize, "calculated part size should match expected")
			}
		})
	}
}
