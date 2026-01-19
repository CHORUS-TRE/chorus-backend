package service

import (
	"context"
	"fmt"
	"path"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/model"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

var _ Requester = (*RequestService)(nil)

type RequestFilter struct {
	StatusesIn        *[]model.RequestStatus
	TypesIn           *[]model.RequestType
	SourceWorkspaceID *uint64
	PendingApproval   *bool
}

type Requester interface {
	GetRequest(ctx context.Context, tenantID, requestID uint64) (*model.Request, error)
	ListRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter RequestFilter) ([]*model.Request, *common_model.PaginationResult, error)
	CreateRequest(ctx context.Context, request *model.Request, filePaths []string) (*model.Request, error)
	ApproveRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.Request, error)
	DeleteRequest(ctx context.Context, tenantID, requestID, userID uint64) error
}

type RequestStore interface {
	GetRequest(ctx context.Context, tenantID, requestID uint64) (*model.Request, error)
	ListRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter RequestFilter) ([]*model.Request, *common_model.PaginationResult, error)
	CreateRequest(ctx context.Context, tenantID uint64, request *model.Request) (*model.Request, error)
	UpdateRequest(ctx context.Context, tenantID uint64, request *model.Request) (*model.Request, error)
	DeleteRequest(ctx context.Context, tenantID, requestID uint64) error
}

type RequestService struct {
	store            RequestStore
	sourceFileStore  filestore.FileStore
	requestFileStore filestore.FileStore
	workspacePrefix  string
}

func NewRequestService(store RequestStore, sourceFileStore filestore.FileStore, requestFileStore filestore.FileStore, workspacePrefix string) *RequestService {
	return &RequestService{
		store:            store,
		sourceFileStore:  sourceFileStore,
		requestFileStore: requestFileStore,
		workspacePrefix:  workspacePrefix,
	}
}

func (s *RequestService) GetRequest(ctx context.Context, tenantID, requestID uint64) (*model.Request, error) {
	return s.store.GetRequest(ctx, tenantID, requestID)
}

func (s *RequestService) ListRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter RequestFilter) ([]*model.Request, *common_model.PaginationResult, error) {
	return s.store.ListRequests(ctx, tenantID, userID, pagination, filter)
}

func (s *RequestService) CreateRequest(ctx context.Context, request *model.Request, filePaths []string) (*model.Request, error) {
	request.Status = model.RequestStatusPending

	if request.Type == model.RequestTypeCopyToWorkspace && request.DestinationWorkspaceID == nil {
		return nil, fmt.Errorf("destination workspace ID is required for copy to workspace requests")
	}

	createdRequest, err := s.store.CreateRequest(ctx, request.TenantID, request)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	requestFiles, err := s.copyFilesToRequestStorage(ctx, createdRequest.ID, request.SourceWorkspaceID, filePaths)
	if err != nil {
		_ = s.store.DeleteRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to copy files to request storage: %w", err)
	}

	createdRequest.Files = requestFiles
	updatedRequest, err := s.store.UpdateRequest(ctx, request.TenantID, createdRequest)
	if err != nil {
		_ = s.cleanupRequestStorage(ctx, createdRequest.ID)
		_ = s.store.DeleteRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to update request with files: %w", err)
	}

	return updatedRequest, nil
}

func (s *RequestService) ApproveRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.Request, error) {
	request, err := s.store.GetRequest(ctx, tenantID, requestID)
	if err != nil {
		return nil, fmt.Errorf("unable to get request: %w", err)
	}

	if !request.CanBeApprovedBy(userID) {
		return nil, fmt.Errorf("user is not authorized to approve this request")
	}

	if approve {
		request.Status = model.RequestStatusApproved
		request.ApprovedByID = &userID

		if request.Type == model.RequestTypeCopyToWorkspace {
			err = s.copyFilesToDestinationWorkspace(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("unable to copy files to destination workspace: %w", err)
			}
		}
	} else {
		request.Status = model.RequestStatusRejected
	}

	updatedRequest, err := s.store.UpdateRequest(ctx, tenantID, request)
	if err != nil {
		return nil, fmt.Errorf("unable to update request: %w", err)
	}

	return updatedRequest, nil
}

func (s *RequestService) DeleteRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	request, err := s.store.GetRequest(ctx, tenantID, requestID)
	if err != nil {
		return fmt.Errorf("unable to get request: %w", err)
	}

	if !request.CanBeDeletedBy(userID) {
		return fmt.Errorf("request cannot be deleted: either not owned by user or in final state")
	}

	err = s.cleanupRequestStorage(ctx, requestID)
	if err != nil {
		return fmt.Errorf("unable to cleanup request storage: %w", err)
	}

	return s.store.DeleteRequest(ctx, tenantID, requestID)
}

func (s *RequestService) copyFilesToRequestStorage(ctx context.Context, requestID, sourceWorkspaceID uint64, filePaths []string) ([]model.RequestFile, error) {
	var requestFiles []model.RequestFile

	workspaceDir := fmt.Sprintf(s.workspacePrefix, workspace_model.GetWorkspaceClusterName(sourceWorkspaceID))
	requestDir := model.GetRequestStoragePath(requestID)

	for _, filePath := range filePaths {
		sourcePath := path.Join(workspaceDir, filePath)
		destPath := path.Join(requestDir, filePath)

		file, err := s.sourceFileStore.GetFile(ctx, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("unable to get source file %s: %w", filePath, err)
		}

		if file.IsDirectory {
			return nil, fmt.Errorf("directories are not supported: %s", filePath)
		}

		destFile := &filestore.File{
			Path:    destPath,
			Name:    file.Name,
			Content: file.Content,
		}

		_, err = s.requestFileStore.CreateFile(ctx, destFile)
		if err != nil {
			return nil, fmt.Errorf("unable to copy file %s to request storage: %w", filePath, err)
		}

		requestFiles = append(requestFiles, model.RequestFile{
			SourcePath:      filePath,
			DestinationPath: destPath,
			Size:            file.Size,
		})
	}

	return requestFiles, nil
}

func (s *RequestService) copyFilesToDestinationWorkspace(ctx context.Context, request *model.Request) error {
	if request.DestinationWorkspaceID == nil {
		return fmt.Errorf("destination workspace ID is required")
	}

	destWorkspaceDir := fmt.Sprintf(s.workspacePrefix, workspace_model.GetWorkspaceClusterName(*request.DestinationWorkspaceID))

	for _, reqFile := range request.Files {
		file, err := s.requestFileStore.GetFile(ctx, reqFile.DestinationPath)
		if err != nil {
			return fmt.Errorf("unable to get file from request storage %s: %w", reqFile.DestinationPath, err)
		}

		destPath := path.Join(destWorkspaceDir, reqFile.SourcePath)
		destFile := &filestore.File{
			Path:    destPath,
			Name:    file.Name,
			Content: file.Content,
		}

		_, err = s.sourceFileStore.CreateFile(ctx, destFile)
		if err != nil {
			return fmt.Errorf("unable to copy file to destination workspace: %w", err)
		}
	}

	return nil
}

func (s *RequestService) cleanupRequestStorage(ctx context.Context, requestID uint64) error {
	requestDir := model.GetRequestStoragePath(requestID)
	return s.requestFileStore.DeleteDirectory(ctx, requestDir)
}
