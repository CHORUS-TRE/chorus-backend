package service

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	workspace_file_service "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
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

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *notification_model.Notification, userIDs []uint64) error
}

type UserPermissionFinder interface {
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter authorization_model.FindUsersWithPermissionFilter) ([]uint64, error)
}

type ApprovalRequestService struct {
	store                ApprovalRequestStore
	workspaceFileStore   workspace_file_service.WorkspaceFiler
	stagingFileStore     filestore.FileStore
	notificationStore    NotificationStore
	userPermissionFinder UserPermissionFinder
	cfg                  config.Config
}

func NewApprovalRequestService(
	store ApprovalRequestStore,
	workspaceFileStore workspace_file_service.WorkspaceFiler,
	stagingFileStore filestore.FileStore,
	notificationStore NotificationStore,
	userPermissionFinder UserPermissionFinder,
	cfg config.Config,
) *ApprovalRequestService {
	return &ApprovalRequestService{
		store:                store,
		workspaceFileStore:   workspaceFileStore,
		stagingFileStore:     stagingFileStore,
		notificationStore:    notificationStore,
		userPermissionFinder: userPermissionFinder,
		cfg:                  cfg,
	}
}

func (s *ApprovalRequestService) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	return s.store.GetApprovalRequest(ctx, tenantID, requestID)
}

func (s *ApprovalRequestService) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	return s.store.ListApprovalRequests(ctx, tenantID, userID, pagination, filter)
}

func (s *ApprovalRequestService) CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Type = model.ApprovalRequestTypeDataExtraction

	details := request.Details.DataExtractionDetails
	if details == nil {
		return nil, fmt.Errorf("invalid details type for data extraction request")
	}

	approvers, canAutoApprove, err := s.findApproversForDataExtractionRequest(ctx, request.TenantID, request.RequesterID, details.SourceWorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("unable to find approvers: %w", err)
	}

	if canAutoApprove {
		request.Status = model.ApprovalRequestStatusApproved
		request.AutoApproved = true
		request.ApprovalMessage = "Auto-approved: requester has permission to download files from workspace"
		request.ApprovedByID = &request.RequesterID
		now := time.Now()
		request.ApprovedAt = &now
	} else {
		request.Status = model.ApprovalRequestStatusPending
		request.ApproverIDs = approvers
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

	if !canAutoApprove {
		for _, approverID := range approvers {
			_ = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
				TenantID: request.TenantID,
				UserID:   approverID,
				Message:  fmt.Sprintf("Approval request '%s' has been created and is pending your approval.", request.Title),
				Content: notification_model.NotificationContent{
					Type: "ApprovalRequestNotification",
					ApprovalRequest: &notification_model.ApprovalRequestNotification{
						ApprovalRequestID: updatedRequest.ID,
					},
				},
			}, []uint64{approverID})
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) findApproversForDataExtractionRequest(ctx context.Context, tenantID, requesterID, workspaceID uint64) ([]uint64, bool, error) {
	workspaceContext := authorization_model.NewContext(authorization_model.WithWorkspace(workspaceID))

	filter := authorization_model.FindUsersWithPermissionFilter{
		PermissionName:          authorization_model.PermissionDownloadFilesFromWorkspace,
		Context:                 workspaceContext,
		PreferExactContextMatch: true,
	}

	if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
		filter.ViaRoles = []authorization_model.RoleName{authorization_model.RoleWorkspaceDataManager}
	}

	approvers, err := s.userPermissionFinder.FindUsersWithPermission(ctx, tenantID, filter)
	if err != nil {
		return nil, false, fmt.Errorf("unable to find approvers: %w", err)
	}

	requesterCanApprove := false
	for _, approverID := range approvers {
		if approverID == requesterID {
			requesterCanApprove = true
			break
		}
	}

	if len(approvers) == 0 {
		if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
			return nil, false, fmt.Errorf("no workspace data manager found for this workspace; please assign a data manager before creating approval requests")
		}
		return nil, false, fmt.Errorf("no users with approval permission found for this workspace")
	}

	return approvers, requesterCanApprove, nil
}

func (s *ApprovalRequestService) CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Type = model.ApprovalRequestTypeDataTransfer

	details := request.Details.DataTransferDetails
	if details == nil {
		return nil, fmt.Errorf("invalid details type for data transfer request")
	}

	if details.DestinationWorkspaceID == 0 {
		return nil, fmt.Errorf("destination workspace ID is required for data transfer requests")
	}

	approvers, canAutoApprove, err := s.findApproversForDataTransferRequest(ctx, request.TenantID, request.RequesterID, details.SourceWorkspaceID, details.DestinationWorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("unable to find approvers: %w", err)
	}

	if canAutoApprove {
		request.Status = model.ApprovalRequestStatusApproved
		request.AutoApproved = true
		request.ApprovalMessage = "Auto-approved: requester has data transfer permission"
		request.ApprovedByID = &request.RequesterID
		now := time.Now()
		request.ApprovedAt = &now
	} else {
		request.Status = model.ApprovalRequestStatusPending
		request.ApproverIDs = approvers
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

	if createdRequest.AutoApproved {
		if err := s.executeApprovedRequest(ctx, updatedRequest); err != nil {
			return nil, fmt.Errorf("unable to execute auto-approved request: %w", err)
		}
	} else {
		for _, approverID := range approvers {
			_ = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
				TenantID: request.TenantID,
				UserID:   approverID,
				Message:  fmt.Sprintf("Approval request '%s' has been created and is pending your approval.", request.Title),
				Content: notification_model.NotificationContent{
					Type: "ApprovalRequestNotification",
					ApprovalRequest: &notification_model.ApprovalRequestNotification{
						ApprovalRequestID: updatedRequest.ID,
					},
				},
			}, []uint64{approverID})
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) findApproversForDataTransferRequest(ctx context.Context, tenantID, requesterID, sourceWorkspaceID, targetWorkspaceID uint64) ([]uint64, bool, error) {
	downloadWorkspaceContext := authorization_model.NewContext(authorization_model.WithWorkspace(sourceWorkspaceID))
	uploadWorkspaceContext := authorization_model.NewContext(authorization_model.WithWorkspace(targetWorkspaceID))

	filterDownload := authorization_model.FindUsersWithPermissionFilter{
		PermissionName:          authorization_model.PermissionDownloadFilesFromWorkspace,
		Context:                 downloadWorkspaceContext,
		PreferExactContextMatch: true,
	}
	if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
		filterDownload.ViaRoles = []authorization_model.RoleName{authorization_model.RoleWorkspaceDataManager}
	}

	downloadApprovers, err := s.userPermissionFinder.FindUsersWithPermission(ctx, tenantID, filterDownload)
	if err != nil {
		return nil, false, fmt.Errorf("unable to find approvers: %w", err)
	}

	filterUpload := authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionUploadFilesToWorkspace,
		Context:        uploadWorkspaceContext,
	}

	uploadApprovers, err := s.userPermissionFinder.FindUsersWithPermission(ctx, tenantID, filterUpload)
	if err != nil {
		return nil, false, fmt.Errorf("unable to find approvers: %w", err)
	}

	requesterCanDownloadApprove := false
	for _, approverID := range downloadApprovers {
		if approverID == requesterID {
			requesterCanDownloadApprove = true
			break
		}
	}
	requesterCanUploadApprove := false
	for _, approverID := range uploadApprovers {
		if approverID == requesterID {
			requesterCanUploadApprove = true
			break
		}
	}
	requesterCanApprove := requesterCanDownloadApprove && requesterCanUploadApprove

	approvers := make([]uint64, 0)
	downloadApproversMap := make(map[uint64]struct{})
	for _, approverID := range downloadApprovers {
		downloadApproversMap[approverID] = struct{}{}
	}
	for _, approverID := range uploadApprovers {
		if _, ok := downloadApproversMap[approverID]; ok {
			approvers = append(approvers, approverID)
		}
	}

	if len(approvers) == 0 {
		if len(downloadApprovers) == 0 && s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
			return nil, false, fmt.Errorf("no workspace data manager found for this workspace; please assign a data manager before creating approval requests")
		}
		return nil, false, fmt.Errorf("no users with approval permission found for this workspace")
	}

	return approvers, requesterCanApprove, nil
}

func (s *ApprovalRequestService) ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error) {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return nil, fmt.Errorf("unable to get request: %w", err)
	}

	if request.Status != model.ApprovalRequestStatusPending {
		return nil, fmt.Errorf("request is not pending approval")
	}

	if approve {
		request.Status = model.ApprovalRequestStatusApproved
	} else {
		request.Status = model.ApprovalRequestStatusRejected
	}
	request.ApprovedByID = &userID
	now := time.Now()
	request.ApprovedAt = &now

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

	requestDir := model.GetApprovalRequestStoragePath(requestID)

	for _, filePath := range filePaths {
		destPath := path.Join(requestDir, filePath)

		file, err := s.workspaceFileStore.GetWorkspaceFileWithContent(ctx, sourceWorkspaceID, filePath)
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

		_, err = s.stagingFileStore.CreateFile(ctx, destFile)
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
	return s.stagingFileStore.DeleteDirectory(ctx, requestDir)
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
	for _, reqFile := range details.Files {
		file, err := s.stagingFileStore.GetFile(ctx, reqFile.DestinationPath)
		if err != nil {
			return fmt.Errorf("unable to get file from request storage %s: %w", reqFile.DestinationPath, err)
		}

		destFile := &filestore.File{
			Path:    reqFile.SourcePath,
			Name:    file.Name,
			Content: file.Content,
		}

		_, err = s.workspaceFileStore.CreateWorkspaceFile(ctx, details.DestinationWorkspaceID, destFile)
		if err != nil {
			return fmt.Errorf("unable to copy file to destination workspace: %w", err)
		}
	}

	return nil
}
