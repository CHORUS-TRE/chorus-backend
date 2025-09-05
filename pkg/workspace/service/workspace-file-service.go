package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

func getSampleWorkspaceFiles() map[uint64]map[string]*model.WorkspaceFile {
	files := make(map[uint64]map[string]*model.WorkspaceFile)

	now := time.Now()

	// Workspace 1 sample files
	files[1] = make(map[string]*model.WorkspaceFile)
	files[1]["/"] = &model.WorkspaceFile{
		Name:        "workspace1",
		Path:        "/",
		IsDirectory: true,
		Size:        0,
		MimeType:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     nil,
	}
	files[1]["/README.md"] = &model.WorkspaceFile{
		Name:        "README.md",
		Path:        "/README.md",
		IsDirectory: false,
		Size:        245,
		MimeType:    "text/markdown",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("# My Project\n\nThis is a sample project for testing.\n\n## Features\n- File management\n- Preview functionality\n- Simple API"),
	}
	files[1]["/src"] = &model.WorkspaceFile{
		Name:        "src",
		Path:        "/src",
		IsDirectory: true,
		Size:        0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	files[1]["/src/main.go"] = &model.WorkspaceFile{
		Name:        "main.go",
		Path:        "/src/main.go",
		IsDirectory: false,
		Size:        156,
		MimeType:    "text/x-go",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"),
	}
	files[1]["/src/utils"] = &model.WorkspaceFile{
		Name:        "utils",
		Path:        "/src/utils",
		IsDirectory: true,
		Size:        0,
		MimeType:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     nil,
	}
	files[1]["/src/utils/helper.go"] = &model.WorkspaceFile{
		Name:        "helper.go",
		Path:        "/src/utils/helper.go",
		IsDirectory: false,
		Size:        78,
		MimeType:    "text/x-go",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("package utils\n\nfunc Helper() string {\n\treturn \"utility function\"\n}"),
	}

	// Workspace 2 sample files
	files[2] = make(map[string]*model.WorkspaceFile)
	files[2]["/"] = &model.WorkspaceFile{
		Name:        "workspace2",
		Path:        "/",
		IsDirectory: true,
		Size:        0,
		MimeType:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     nil,
	}
	files[2]["/config.yaml"] = &model.WorkspaceFile{
		Name:        "config.yaml",
		Path:        "/config.yaml",
		IsDirectory: false,
		Size:        98,
		MimeType:    "application/x-yaml",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("app:\n  name: chorus-backend\n  version: 1.0.0\nserver:\n  port: 8080\n  host: localhost"),
	}
	files[2]["/scripts"] = &model.WorkspaceFile{
		Name:        "scripts",
		Path:        "/scripts",
		IsDirectory: true,
		Size:        0,
		MimeType:    "",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     nil,
	}
	files[2]["/scripts/deploy.sh"] = &model.WorkspaceFile{
		Name:        "deploy.sh",
		Path:        "/scripts/deploy.sh",
		IsDirectory: false,
		Size:        67,
		MimeType:    "application/x-sh",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("#!/bin/bash\necho \"Deploying application...\"\nkubectl apply -f deploy/"),
	}
	files[2]["/scripts/test.sh"] = &model.WorkspaceFile{
		Name:        "test.sh",
		Path:        "/scripts/test.sh",
		IsDirectory: false,
		Size:        45,
		MimeType:    "application/x-sh",
		CreatedAt:   now,
		UpdatedAt:   now,
		Content:     []byte("#!/bin/bash\necho \"Running tests...\"\ngo test ./..."),
	}

	return files
}

func (s *WorkspaceService) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	s.filesMu.RLock()
	defer s.filesMu.RUnlock()

	filePath = normalizePath(filePath)

	workspaceFiles, exists := s.files[workspaceID]
	if !exists {
		return nil, fmt.Errorf("workspace not found: %d", workspaceID)
	}

	file, exists := workspaceFiles[filePath]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Return a copy to prevent external mutation
	fileCopy := *file
	return &fileCopy, nil
}

func (s *WorkspaceService) GetWorkspaceFileChildren(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	s.filesMu.RLock()
	defer s.filesMu.RUnlock()

	filePath = normalizePath(filePath)

	workspaceFiles, exists := s.files[workspaceID]
	if !exists {
		return []*model.WorkspaceFile{}, nil
	}

	children := s.getChildren(workspaceFiles, filePath)

	result := make([]*model.WorkspaceFile, len(children))
	for i, child := range children {
		childCopy := *child
		result[i] = &childCopy
	}

	return result, nil
}

func (s *WorkspaceService) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if file.Path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	file.Path = normalizePath(file.Path)

	if _, exists := s.files[workspaceID]; !exists {
		s.files[workspaceID] = make(map[string]*model.WorkspaceFile)
	}

	if _, exists := s.files[workspaceID][file.Path]; exists {
		return nil, fmt.Errorf("file already exists: %s", file.Path)
	}

	newFile := *file
	now := time.Now()
	newFile.CreatedAt = now
	newFile.UpdatedAt = now

	s.files[workspaceID][file.Path] = &newFile

	// Return a copy
	result := newFile
	return &result, nil
}

func (s *WorkspaceService) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if file.Path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}
	if file.Path == "/" {
		return nil, fmt.Errorf("cannot update root workspace directory")
	}

	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	file.Path = normalizePath(file.Path)

	workspaceFiles, exists := s.files[workspaceID]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", file.Path)
	}

	existingFile, exists := workspaceFiles[file.Path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", file.Path)
	}

	delete(workspaceFiles, existingFile.Path)

	updatedFile := *file
	updatedFile.CreatedAt = existingFile.CreatedAt
	updatedFile.UpdatedAt = time.Now()
	s.files[workspaceID][updatedFile.Path] = &updatedFile

	result := updatedFile
	return &result, nil
}

func (s *WorkspaceService) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if filePath == "/" {
		return fmt.Errorf("cannot delete root workspace directory")
	}

	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	filePath = normalizePath(filePath)

	workspaceFiles, exists := s.files[workspaceID]
	if !exists {
		return fmt.Errorf("file not found: %s", filePath)
	}

	if _, exists := workspaceFiles[filePath]; !exists {
		return fmt.Errorf("file not found: %s", filePath)
	}

	delete(workspaceFiles, filePath)

	return nil
}

func (s *WorkspaceService) getChildren(workspaceFiles map[string]*model.WorkspaceFile, parentPath string) []*model.WorkspaceFile {
	var children []*model.WorkspaceFile

	for _, file := range workspaceFiles {
		// Skip the parent itself
		if file.Path == parentPath {
			continue
		}

		// For root directory
		if parentPath == "/" {
			// Get files that are direct children of root (no additional slashes after the first)
			if strings.HasPrefix(file.Path, "/") && file.Path != "/" {
				relativePath := strings.TrimPrefix(file.Path, "/")
				if !strings.Contains(relativePath, "/") {
					children = append(children, file)
				}
			}
		} else {
			// For subdirectories, check if file is a direct child
			if strings.HasPrefix(file.Path, parentPath+"/") {
				relativePath := strings.TrimPrefix(file.Path, parentPath+"/")
				// Check if it's a direct child (no more slashes in relative path)
				if !strings.Contains(relativePath, "/") {
					children = append(children, file)
				}
			}
		}
	}
	return children
}

func normalizePath(path string) string {
	// Handle empty path or root
	if path == "" || path == "/" {
		return "/"
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Remove trailing slash (except for root)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}
