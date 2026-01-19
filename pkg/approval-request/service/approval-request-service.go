package service

import (
	"context"
	"fmt"
	"path"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

var _ ApprovalRequester = (*ApprovalRequestService)(nil)

type ApprovalRequestFilter struct {
	StatusesIn        *[]model.ApprovalRequestStatus
	TypesIn           *[]model.ApprovalRequestType
	SourceWorkspaceID *uint64
	PendingApproval   *bool
}

type ApprovalRequester interface {
	GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error)
	ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error)
	CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error)
	CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error)
	ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error)
	DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error
}

type ApprovalRequestStore interface {
	GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error)
	ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error)
	CreateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error)
	UpdateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error)
	DeleteApprovalRequest(ctx context.Context, tenantID, requestID uint64) error
}

type ApprovalRequestService struct {
	store            ApprovalRequestStore
	sourceFileStore  filestore.FileStore
	requestFileStore filestore.FileStore
	workspacePrefix  string
}

func NewApprovalRequestService(store ApprovalRequestStore, sourceFileStore filestore.FileStore, requestFileStore filestore.FileStore, workspacePrefix string) *ApprovalRequestService {
	return &ApprovalRequestService{
		store:            store,
		sourceFileStore:  sourceFileStore,
		requestFileStore: requestFileStore,
		workspacePrefix:  workspacePrefix,
	}
}

func (s *ApprovalRequestService) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	return s.store.GetApprovalRequest(ctx, tenantID, requestID)
}

func (s *ApprovalRequestService) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	return s.store.ListApprovalRequests(ctx, tenantID, userID, pagination, filter)
}

func (s *ApprovalRequestService) CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Status = model.ApprovalRequestStatusPending
	request.Type = model.ApprovalRequestTypeDataExtraction

	details := request.Details.DataExtractionDetails
	if details == nil {
		return nil, fmt.Errorf("invalid details type for data extraction request")
	}

	createdRequest, err := s.store.CreateApprovalRequest(ctx, request.TenantID, request)
	if err != nil {
		return nil, fmt.Errorf("unable to create approval request: %w", err)
	}

	requestFiles, err := s.copyFilesToRequestStorage(ctx, createdRequest.ID, details.SourceWorkspaceID, filePaths)
	if err != nil {
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to copy files to request storage: %w", err)
	}

	createdDetails := createdRequest.Details.DataExtractionDetails
	if createdDetails == nil {
		return nil, fmt.Errorf("invalid details type for data extraction request")
	}
	createdDetails.Files = requestFiles

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, request.TenantID, createdRequest)
	if err != nil {
		_ = s.cleanupRequestStorage(ctx, createdRequest.ID)
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to update request with files: %w", err)
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Status = model.ApprovalRequestStatusPending
	request.Type = model.ApprovalRequestTypeDataTransfer

	details := request.Details.DataTransferDetails
	if details == nil {
		return nil, fmt.Errorf("invalid details type for data transfer request")
	}

	if details.DestinationWorkspaceID == 0 {
		return nil, fmt.Errorf("destination workspace ID is required for data transfer requests")
	}

	createdRequest, err := s.store.CreateApprovalRequest(ctx, request.TenantID, request)
	if err != nil {
		return nil, fmt.Errorf("unable to create approval request: %w", err)
	}

	requestFiles, err := s.copyFilesToRequestStorage(ctx, createdRequest.ID, details.SourceWorkspaceID, filePaths)
	if err != nil {
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to copy files to request storage: %w", err)
	}

	createdDetails := createdRequest.Details.DataTransferDetails
	if createdDetails == nil {
		return nil, fmt.Errorf("invalid details type for data transfer request")
	}
	createdDetails.Files = requestFiles

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, request.TenantID, createdRequest)
	if err != nil {
		_ = s.cleanupRequestStorage(ctx, createdRequest.ID)
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, fmt.Errorf("unable to update request with files: %w", err)
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error) {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return nil, fmt.Errorf("unable to get request: %w", err)
	}

	if !request.CanBeApprovedBy(userID) {
		return nil, fmt.Errorf("user is not authorized to approve this request")
	}

	if approve {
		request.Status = model.ApprovalRequestStatusApproved
	} else {
		request.Status = model.ApprovalRequestStatusRejected
	}
	request.ApprovedByID = &userID

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, tenantID, request)
	if err != nil {
		return nil, fmt.Errorf("unable to update request: %w", err)
	}

	if approve {
		if err := s.executeApprovedRequest(ctx, updatedRequest); err != nil {
			return nil, fmt.Errorf("unable to execute approved request: %w", err)
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return fmt.Errorf("unable to get request: %w", err)
	}

	if !request.CanBeDeletedBy(userID) {
		return fmt.Errorf("user is not authorized to delete this request")
	}

	if err := s.cleanupRequestStorage(ctx, requestID); err != nil {
		return fmt.Errorf("unable to cleanup request storage: %w", err)
	}

	return s.store.DeleteApprovalRequest(ctx, tenantID, requestID)
}

func (s *ApprovalRequestService) copyFilesToRequestStorage(ctx context.Context, requestID, sourceWorkspaceID uint64, filePaths []string) ([]model.ApprovalRequestFile, error) {
	var requestFiles []model.ApprovalRequestFile

	workspaceDir := fmt.Sprintf(s.workspacePrefix, workspace_model.GetWorkspaceClusterName(sourceWorkspaceID))
	requestDir := model.GetApprovalRequestStoragePath(requestID)

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

		requestFiles = append(requestFiles, model.ApprovalRequestFile{
			SourcePath:      filePath,
			DestinationPath: destPath,
			Size:            file.Size,
		})
	}

	return requestFiles, nil
}

func (s *ApprovalRequestService) cleanupRequestStorage(ctx context.Context, requestID uint64) error {
	requestDir := model.GetApprovalRequestStoragePath(requestID)
	return s.requestFileStore.DeleteDirectory(ctx, requestDir)
}

func (s *ApprovalRequestService) executeApprovedRequest(ctx context.Context, request *model.ApprovalRequest) error {
	switch request.Type {
	case model.ApprovalRequestTypeDataExtraction:
		return nil
	case model.ApprovalRequestTypeDataTransfer:
		details := request.Details.DataTransferDetails
		if details == nil {
			return fmt.Errorf("invalid details type for data transfer request")
		}
		return s.copyFilesToDestinationWorkspace(ctx, *details)
	default:
		return fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

func (s *ApprovalRequestService) copyFilesToDestinationWorkspace(ctx context.Context, details model.DataTransferDetails) error {
	destWorkspaceDir := fmt.Sprintf(s.workspacePrefix, workspace_model.GetWorkspaceClusterName(details.DestinationWorkspaceID))

	for _, reqFile := range details.Files {
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
