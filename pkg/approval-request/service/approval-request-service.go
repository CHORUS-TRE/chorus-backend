package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	workspace_file_service "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
	"go.uber.org/zap"
)

var _ ApprovalRequester = (*ApprovalRequestService)(nil)

type ApprovalRequestFilter struct {
	StatusesIn        *[]model.ApprovalRequestStatus
	TypesIn           *[]model.ApprovalRequestType
	WorkspaceID *uint64
	PendingApproval   *bool
	ApproverID        *uint64
	RequesterID       *uint64
}

type ApprovalRequester interface {
	GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error)
	ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error)
	CountMyApprovalRequests(ctx context.Context, tenantID, userID uint64) (*model.ApprovalRequestCounts, error)
	CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error)
	CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error)
	ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error)
	DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error
	DownloadApprovalRequestFile(ctx context.Context, tenantID, requestID uint64, filePath string) (*model.ApprovalRequestFile, []byte, error)
}

type ApprovalRequestStore interface {
	GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error)
	ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error)
	CountMyApprovalRequests(ctx context.Context, tenantID, userID uint64) (*model.ApprovalRequestCounts, error)
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

func (s *ApprovalRequestService) CountMyApprovalRequests(ctx context.Context, tenantID, userID uint64) (*model.ApprovalRequestCounts, error) {
	return s.store.CountMyApprovalRequests(ctx, tenantID, userID)
}

// CreateDataExtractionRequest creates an approval request to download files from a workspace.
//
// Flow:
//  1. Determine approvers and whether the requester can self-approve.
//  2. Persist the request in the database (status: pending or approved).
//  3. Copy the requested files from the source workspace into an immutable
//     staging area so auditors can review the exact content.
//  4. Update the request with the file metadata (staging paths + sizes).
//  5. If auto-approved, the files are immediately available for download
//     from the staging area. Otherwise, notify the approvers.
func (s *ApprovalRequestService) CreateDataExtractionRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Type = model.ApprovalRequestTypeDataExtraction

	details := request.Details.DataExtractionDetails
	if details == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Invalid details type for data extraction request")
	}

	approversByStep, canAutoApprove, err := s.findApproversForDataExtractionRequest(ctx, request.TenantID, request.RequesterID, details.SourceWorkspaceID)
	if err != nil {
		return nil, err
	}

	if canAutoApprove {
		request.Status = model.ApprovalRequestStatusApproved
		request.AutoApproved = true
		request.ApprovalMessage = "Auto-approved: requester has permission to download files from workspace"
		now := time.Now()
		request.ApprovedAt = &now
	} else {
		request.Status = model.ApprovalRequestStatusPending
		request.ApproverIDsByStep = approversByStep
	}

	createdRequest, err := s.store.CreateApprovalRequest(ctx, request.TenantID, request)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to create approval request")
	}

	requestFiles, err := s.copyFilesToRequestStorage(ctx, createdRequest.ID, details.SourceWorkspaceID, filePaths)
	if err != nil {
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, err
	}

	createdDetails := createdRequest.Details.DataExtractionDetails
	if createdDetails == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Invalid details type for data extraction request")
	}
	createdDetails.Files = requestFiles

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, request.TenantID, createdRequest)
	if err != nil {
		_ = s.cleanupRequestStorage(ctx, createdRequest.ID)
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, cerr.WrapStoreError(err, "Unable to update approval request with files")
	}

	for _, approverID := range uniqueApproverIDs(approversByStep) {
		err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
			TenantID: request.TenantID,
			UserID:   approverID,
			Message:  fmt.Sprintf("Approval request '%s' has been created and is pending your approval.", request.Title),
			Content: notification_model.NotificationContent{
				Type: "ApprovalRequestNotification",
				ApprovalRequest: &notification_model.ApprovalRequestNotification{
					ApprovalRequestID: updatedRequest.ID,
					Autoapproved:      canAutoApprove,
				},
			},
		}, []uint64{approverID})
		if err != nil {
			logger.TechLog.Error(ctx, "Unable to create notification", zap.Uint64("tenant_id", request.TenantID), zap.Uint64("request_id", request.ID), zap.Uint64("user_id", approverID))
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) findApproversForDataExtractionRequest(ctx context.Context, tenantID, requesterID, workspaceID uint64) (map[model.ApprovalStep][]uint64, bool, error) {
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
		return nil, false, err
	}

	if len(approvers) == 0 {
		if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
			return nil, false, cerr.ErrInvalidRequest.WithMessage("No workspace data manager found for this workspace; please assign a data manager before creating approval requests")
		}
		return nil, false, cerr.ErrInvalidRequest.WithMessage("No users with approval permission found for this workspace")
	}

	requesterCanApprove := containsID(approvers, requesterID)

	return map[model.ApprovalStep][]uint64{
		model.StepDownload: approvers,
	}, requesterCanApprove, nil
}

// CreateDataTransferRequest creates an approval request to transfer files between workspaces.
//
// Flow:
//  1. Determine approvers and whether the requester can self-approve.
//  2. Persist the request in the database (status: pending or approved).
//  3. Copy the requested files from the source workspace into an immutable
//     staging area so auditors can review the exact content.
//  4. Update the request with the file metadata (staging paths + sizes).
//  5. If auto-approved, immediately copy the files from staging into the
//     destination workspace. Otherwise, notify the approvers.
func (s *ApprovalRequestService) CreateDataTransferRequest(ctx context.Context, request *model.ApprovalRequest, filePaths []string) (*model.ApprovalRequest, error) {
	request.Type = model.ApprovalRequestTypeDataTransfer

	details := request.Details.DataTransferDetails
	if details == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Invalid details type for data transfer request")
	}

	if details.DestinationWorkspaceID == 0 {
		return nil, cerr.ErrInvalidRequest.WithMessage("Destination workspace ID is required for data transfer requests")
	}

	approversByStep, canAutoApprove, err := s.findApproversForDataTransferRequest(ctx, request.TenantID, request.RequesterID, details.SourceWorkspaceID, details.DestinationWorkspaceID)
	if err != nil {
		return nil, err
	}

	if canAutoApprove {
		request.Status = model.ApprovalRequestStatusApproved
		request.AutoApproved = true
		request.ApprovalMessage = "Auto-approved: requester has data transfer permission"
		now := time.Now()
		request.ApprovedAt = &now
	} else {
		request.Status = model.ApprovalRequestStatusPending
		request.ApproverIDsByStep = approversByStep
	}

	createdRequest, err := s.store.CreateApprovalRequest(ctx, request.TenantID, request)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to create approval request")
	}

	requestFiles, err := s.copyFilesToRequestStorage(ctx, createdRequest.ID, details.SourceWorkspaceID, filePaths)
	if err != nil {
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, err
	}

	createdDetails := createdRequest.Details.DataTransferDetails
	if createdDetails == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Invalid details type for data transfer request")
	}
	createdDetails.Files = requestFiles

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, request.TenantID, createdRequest)
	if err != nil {
		_ = s.cleanupRequestStorage(ctx, createdRequest.ID)
		_ = s.store.DeleteApprovalRequest(ctx, request.TenantID, createdRequest.ID)
		return nil, cerr.WrapStoreError(err, "Unable to update approval request with files")
	}

	if canAutoApprove {
		if err := s.executeApprovedRequest(ctx, updatedRequest); err != nil {
			return nil, err
		}
	}

	for _, approverID := range uniqueApproverIDs(approversByStep) {
		err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
			TenantID: request.TenantID,
			UserID:   approverID,
			Message:  fmt.Sprintf("Approval request '%s' has been created and is pending your approval.", request.Title),
			Content: notification_model.NotificationContent{
				Type: "ApprovalRequestNotification",
				ApprovalRequest: &notification_model.ApprovalRequestNotification{
					ApprovalRequestID: updatedRequest.ID,
					Autoapproved:      canAutoApprove,
				},
			},
		}, []uint64{approverID})
		if err != nil {
			logger.TechLog.Error(ctx, "Unable to create notification", zap.Uint64("tenant_id", request.TenantID), zap.Uint64("request_id", request.ID), zap.Uint64("user_id", approverID))
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) findApproversForDataTransferRequest(ctx context.Context, tenantID, requesterID, sourceWorkspaceID, targetWorkspaceID uint64) (map[model.ApprovalStep][]uint64, bool, error) {
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
		return nil, false, err
	}

	filterUpload := authorization_model.FindUsersWithPermissionFilter{
		PermissionName:          authorization_model.PermissionUploadFilesToWorkspace,
		Context:                 uploadWorkspaceContext,
		PreferExactContextMatch: true,
	}
	if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
		filterUpload.ViaRoles = []authorization_model.RoleName{authorization_model.RoleWorkspaceDataManager}
	}

	uploadApprovers, err := s.userPermissionFinder.FindUsersWithPermission(ctx, tenantID, filterUpload)
	if err != nil {
		return nil, false, err
	}

	if len(downloadApprovers) == 0 {
		if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
			return nil, false, cerr.ErrInvalidRequest.WithMessage("No workspace data manager found for the source workspace; please assign a data manager before creating approval requests")
		}
		return nil, false, cerr.ErrInvalidRequest.WithMessage("No users with download approval permission found for the source workspace")
	}
	if len(uploadApprovers) == 0 {
		if s.cfg.Services.ApprovalRequestService.RequireDataManagerApproval {
			return nil, false, cerr.ErrInvalidRequest.WithMessage("No workspace data manager found for the destination workspace; please assign a data manager before creating approval requests")
		}
		return nil, false, cerr.ErrInvalidRequest.WithMessage("No users with upload approval permission found for the destination workspace")
	}

	// The requester may self-approve only if they are authorized for every step.
	requesterCanApprove := containsID(downloadApprovers, requesterID) && containsID(uploadApprovers, requesterID)

	return map[model.ApprovalStep][]uint64{
		model.StepDownload: downloadApprovers,
		model.StepUpload:   uploadApprovers,
	}, requesterCanApprove, nil
}

// containsID reports whether ids contains target.
func containsID(ids []uint64, target uint64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

// uniqueApproverIDs returns the deduplicated union of approver ids across all steps.
func uniqueApproverIDs(approversByStep map[model.ApprovalStep][]uint64) []uint64 {
	seen := make(map[uint64]struct{})
	var ids []uint64
	for _, approvers := range approversByStep {
		for _, id := range approvers {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *ApprovalRequestService) ApproveApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64, approve bool) (*model.ApprovalRequest, error) {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to get approval request")
	}

	if request.Status != model.ApprovalRequestStatusPending {
		return nil, cerr.ErrInvalidRequest.WithMessage("Request is not pending approval")
	}

	// Determine which steps this user is entitled to decide on (and that have
	// not already been decided). If empty, the user cannot act on this request.
	stepsToDecide := request.StepsToApprove(userID)
	if len(stepsToDecide) == 0 {
		return nil, cerr.ErrPermissionDenied.WithMessage("User is not authorized to approve any pending step of this request")
	}

	if request.StepDecisions == nil {
		request.StepDecisions = make(map[model.ApprovalStep]model.ApprovalStepDecision)
	}
	now := time.Now()
	for _, step := range stepsToDecide {
		request.StepDecisions[step] = model.ApprovalStepDecision{
			ApproverID: userID,
			ApprovedAt: now,
			Approve:    approve,
		}
		if !approve {
			// A single rejection rejects the whole request; record only that step.
			break
		}
	}

	// Compute the resulting request status.
	switch {
	case !approve:
		request.Status = model.ApprovalRequestStatusRejected
		request.ApprovedAt = &now
	case request.IsFullyApproved():
		request.Status = model.ApprovalRequestStatusApproved
		request.ApprovedAt = &now
	default:
		// Still waiting on other steps; stay pending.
		request.Status = model.ApprovalRequestStatusPending
	}

	updatedRequest, err := s.store.UpdateApprovalRequest(ctx, tenantID, request)
	if err != nil {
		return nil, cerr.WrapStoreError(err, "Unable to update approval request")
	}

	if updatedRequest.Status == model.ApprovalRequestStatusApproved {
		if err := s.executeApprovedRequest(ctx, updatedRequest); err != nil {
			return nil, err
		}
	}

	// Only notify the requester once the request reaches a terminal state.
	if updatedRequest.IsFinalState() {
		notifyMessage := fmt.Sprintf("Approval request '%s' has been %s.", request.Title, map[bool]string{true: "approved", false: "rejected"}[approve])
		err = s.notificationStore.CreateNotification(ctx, &notification_model.Notification{
			TenantID: request.TenantID,
			UserID:   request.RequesterID,
			Message:  notifyMessage,
			Content: notification_model.NotificationContent{
				Type: "ApprovalRequestNotification",
				ApprovalRequest: &notification_model.ApprovalRequestNotification{
					ApprovalRequestID: updatedRequest.ID,
				},
			},
		}, []uint64{request.RequesterID})
		if err != nil {
			logger.TechLog.Error(ctx, "Unable to create notification", zap.Uint64("tenant_id", request.TenantID), zap.Uint64("request_id", request.ID), zap.Uint64("user_id", request.RequesterID))
		}
	}

	return updatedRequest, nil
}

func (s *ApprovalRequestService) DeleteApprovalRequest(ctx context.Context, tenantID, requestID, userID uint64) error {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return cerr.WrapStoreError(err, "Unable to get approval request")
	}

	if !request.CanBeDeletedBy(userID) {
		return cerr.ErrPermissionDenied.WithMessage("User is not authorized to delete this request")
	}

	if err := s.cleanupRequestStorage(ctx, requestID); err != nil {
		return err
	}

	if err := s.store.DeleteApprovalRequest(ctx, tenantID, requestID); err != nil {
		return cerr.WrapStoreError(err, "Unable to delete approval request")
	}
	return nil
}

// copyFilesToRequestStorage copies files from the source workspace into the
// staging area. This creates an immutable audit trail: the staged files can be
// reviewed by approvers and are the ones ultimately delivered (for transfers)
// or downloaded (for extractions).
func (s *ApprovalRequestService) copyFilesToRequestStorage(ctx context.Context, requestID, sourceWorkspaceID uint64, filePaths []string) ([]model.ApprovalRequestFile, error) {
	var requestFiles []model.ApprovalRequestFile

	requestDir := model.GetApprovalRequestStoragePath(requestID)

	for _, filePath := range filePaths {
		destPath := path.Join(requestDir, filePath)

		file, err := s.workspaceFileStore.GetWorkspaceFileWithContent(ctx, sourceWorkspaceID, filePath)
		if err != nil {
			return nil, err
		}

		if file.IsDirectory {
			return nil, cerr.ErrInvalidRequest.WithMessage(fmt.Sprintf("Directories are not supported: %s", filePath))
		}

		destFile := &filestore.File{
			Path:    destPath,
			Name:    file.Name,
			Content: file.Content,
		}

		_, err = s.stagingFileStore.CreateFile(ctx, destFile)
		if err != nil {
			return nil, cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to copy file %s to request storage", filePath))
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
	if err := s.stagingFileStore.DeleteDirectory(ctx, requestDir); err != nil {
		return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to cleanup request storage for request %d", requestID))
	}
	return nil
}

func (s *ApprovalRequestService) executeApprovedRequest(ctx context.Context, request *model.ApprovalRequest) error {
	logger.TechLog.Debug(ctx, "Executing approved request", zap.Uint64("request_id", request.ID), zap.String("type", string(request.Type)))
	switch request.Type {
	case model.ApprovalRequestTypeDataExtraction:
		return nil
	case model.ApprovalRequestTypeDataTransfer:
		details := request.Details.DataTransferDetails
		if details == nil {
			return cerr.ErrInvalidRequest.WithMessage("Invalid details type for data transfer request")
		}
		return s.copyFilesToDestinationWorkspace(ctx, *details)
	default:
		return cerr.ErrInvalidRequest.WithMessage(fmt.Sprintf("Unsupported request type: %s", request.Type))
	}
}

// copyFilesToDestinationWorkspace reads each file from the staging area
// (DestinationPath) and writes it into the destination workspace, preserving
// the original directory structure (SourcePath).
func (s *ApprovalRequestService) copyFilesToDestinationWorkspace(ctx context.Context, details model.DataTransferDetails) error {
	logger.TechLog.Debug(ctx, "Copying approved files to destination workspace", zap.Uint64("destination_workspace_id", details.DestinationWorkspaceID), zap.Int("file_count", len(details.Files)))
	for _, reqFile := range details.Files {
		file, err := s.stagingFileStore.GetFile(ctx, reqFile.DestinationPath)
		if err != nil {
			return cerr.ErrInternal.Wrap(err, fmt.Sprintf("Unable to get file from request storage %s", reqFile.DestinationPath))
		}

		destFile := &filestore.File{
			Path:    reqFile.SourcePath,
			Name:    file.Name,
			Content: file.Content,
		}

		_, err = s.workspaceFileStore.CreateWorkspaceFile(ctx, details.DestinationWorkspaceID, destFile)
		if err != nil {
			var chorusErr *cerr.ChorusError
			if !errors.As(err, &chorusErr) || chorusErr.ChorusCode != cerr.ErrAlreadyExists.ChorusCode {
				return err
			}

			destFile.Path = appendUUIDToFilename(destFile.Path)
			destFile.Name = path.Base(destFile.Path)
			logger.TechLog.Info(ctx, "File already exists in destination workspace, retrying with unique name", zap.String("new_path", destFile.Path))

			_, err = s.workspaceFileStore.CreateWorkspaceFile(ctx, details.DestinationWorkspaceID, destFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// appendUUIDToFilename inserts an uuid suffix before the file extension.
// e.g. "workspace-archive/hello.txt" -> "workspace-archive/hello_<uuid>.txt"
func appendUUIDToFilename(filePath string) string {
	ext := path.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return fmt.Sprintf("%s_%s%s", base, uuid.Next(), ext)
}

func (s *ApprovalRequestService) DownloadApprovalRequestFile(ctx context.Context, tenantID, requestID uint64, filePath string) (*model.ApprovalRequestFile, []byte, error) {
	request, err := s.store.GetApprovalRequest(ctx, tenantID, requestID)
	if err != nil {
		return nil, nil, cerr.WrapStoreError(err, "Unable to get approval request")
	}

	if request.Type != model.ApprovalRequestTypeDataExtraction {
		return nil, nil, cerr.ErrInvalidRequest.WithMessage("Download is only available for data extraction requests")
	}

	details := request.Details.DataExtractionDetails
	if details == nil {
		return nil, nil, cerr.ErrInvalidRequest.WithMessage("Invalid request details")
	}

	var requestFile *model.ApprovalRequestFile
	for i := range details.Files {
		if details.Files[i].SourcePath == filePath {
			requestFile = &details.Files[i]
			break
		}
	}

	if requestFile == nil {
		return nil, nil, cerr.ErrNotFound.WithMessage("File not found in request")
	}

	file, err := s.stagingFileStore.GetFile(ctx, requestFile.DestinationPath)
	if err != nil {
		return nil, nil, cerr.ErrInternal.Wrap(err, "Unable to get file from staging storage")
	}

	return requestFile, file.Content, nil
}
