package model

import (
	"fmt"
	"time"
)

type ApprovalRequest struct {
	ID          uint64
	TenantID    uint64
	RequesterID uint64

	Type        ApprovalRequestType
	Status      ApprovalRequestStatus
	Title       string
	Description string

	Details ApprovalRequestDetails

	ApproverIDs     []uint64
	ApprovedByID    *uint64
	AutoApproved    bool
	ApprovalMessage string

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ApprovedAt *time.Time
}

type ApprovalRequestCounts struct {
	Total          uint64
	TotalApprover  uint64
	TotalRequester uint64
	CountByStatus  map[string]uint64
	CountByType    map[string]uint64
}

type ApprovalRequestDetails struct {
	DataExtractionDetails *DataExtractionDetails `json:"data_extraction_details,omitempty"`
	DataTransferDetails   *DataTransferDetails   `json:"data_transfer_details,omitempty"`
}

type DataExtractionDetails struct {
	SourceWorkspaceID uint64                `json:"source_workspace_id"`
	Files             []ApprovalRequestFile `json:"files"`
}

type DataTransferDetails struct {
	SourceWorkspaceID      uint64                `json:"source_workspace_id"`
	DestinationWorkspaceID uint64                `json:"destination_workspace_id"`
	Files                  []ApprovalRequestFile `json:"files"`
}

func (r *ApprovalRequest) IsFinalState() bool {
	return r.Status == ApprovalRequestStatusApproved ||
		r.Status == ApprovalRequestStatusRejected ||
		r.Status == ApprovalRequestStatusCancelled
}

func (r *ApprovalRequest) CanBeDeletedBy(userID uint64) bool {
	return r.RequesterID == userID && !r.IsFinalState()
}

func (r *ApprovalRequest) CanBeApprovedBy(userID uint64) bool {
	if r.IsFinalState() {
		return false
	}
	if len(r.ApproverIDs) == 0 {
		return true
	}
	for _, id := range r.ApproverIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (r *ApprovalRequest) GetSourceWorkspaceID() uint64 {
	if r.Details.DataExtractionDetails != nil {
		return r.Details.DataExtractionDetails.SourceWorkspaceID
	}
	if r.Details.DataTransferDetails != nil {
		return r.Details.DataTransferDetails.SourceWorkspaceID
	}
	return 0
}

func GetApprovalRequestStoragePath(requestID uint64) string {
	return fmt.Sprintf("approval-request-%v", requestID)
}

// ApprovalRequestFile tracks a file associated with an approval request.
//
// When a request is created, files are copied from the source workspace into an
// immutable staging area so that auditors can review the exact content that was
// (or will be) transferred. The two path fields reflect this:
//   - SourcePath:      the original path inside the source workspace (e.g. "data/results.csv").
//   - DestinationPath: the path inside the staging area (e.g. "approval-request-42/data/results.csv").
//
// For data-transfer requests, once approved the files are copied from staging
// (DestinationPath) into the destination workspace using SourcePath to preserve
// the original directory structure.
type ApprovalRequestFile struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	Size            uint64 `json:"size"`
}

type ApprovalRequestType string

const (
	ApprovalRequestTypeUnspecified    ApprovalRequestType = ""
	ApprovalRequestTypeDataExtraction ApprovalRequestType = "data_extraction"
	ApprovalRequestTypeDataTransfer   ApprovalRequestType = "data_transfer"
)

func ApprovalRequestTypes() []ApprovalRequestType {
	return []ApprovalRequestType{
		ApprovalRequestTypeDataExtraction,
		ApprovalRequestTypeDataTransfer,
	}
}

func (t ApprovalRequestType) String() string {
	return string(t)
}

func ToApprovalRequestType(s string) (ApprovalRequestType, error) {
	switch s {
	case string(ApprovalRequestTypeDataExtraction), "APPROVAL_REQUEST_TYPE_DATA_EXTRACTION":
		return ApprovalRequestTypeDataExtraction, nil
	case string(ApprovalRequestTypeDataTransfer), "APPROVAL_REQUEST_TYPE_DATA_TRANSFER":
		return ApprovalRequestTypeDataTransfer, nil
	case "", "APPROVAL_REQUEST_TYPE_UNSPECIFIED":
		return ApprovalRequestTypeUnspecified, nil
	default:
		return ApprovalRequestTypeUnspecified, fmt.Errorf("unexpected ApprovalRequestType: %s", s)
	}
}

type ApprovalRequestStatus string

const (
	ApprovalRequestStatusUnspecified ApprovalRequestStatus = ""
	ApprovalRequestStatusPending     ApprovalRequestStatus = "pending"
	ApprovalRequestStatusApproved    ApprovalRequestStatus = "approved"
	ApprovalRequestStatusRejected    ApprovalRequestStatus = "rejected"
	ApprovalRequestStatusCancelled   ApprovalRequestStatus = "cancelled"
)

func ApprovalRequestStatuses() []ApprovalRequestStatus {
	return []ApprovalRequestStatus{
		ApprovalRequestStatusPending,
		ApprovalRequestStatusApproved,
		ApprovalRequestStatusRejected,
		ApprovalRequestStatusCancelled,
	}
}

func (s ApprovalRequestStatus) String() string {
	return string(s)
}

func ToApprovalRequestStatus(s string) (ApprovalRequestStatus, error) {
	switch s {
	case string(ApprovalRequestStatusPending), "APPROVAL_REQUEST_STATUS_PENDING":
		return ApprovalRequestStatusPending, nil
	case string(ApprovalRequestStatusApproved), "APPROVAL_REQUEST_STATUS_APPROVED":
		return ApprovalRequestStatusApproved, nil
	case string(ApprovalRequestStatusRejected), "APPROVAL_REQUEST_STATUS_REJECTED":
		return ApprovalRequestStatusRejected, nil
	case string(ApprovalRequestStatusCancelled), "APPROVAL_REQUEST_STATUS_CANCELLED":
		return ApprovalRequestStatusCancelled, nil
	case "", "APPROVAL_REQUEST_STATUS_UNSPECIFIED":
		return ApprovalRequestStatusUnspecified, nil
	default:
		return ApprovalRequestStatusUnspecified, fmt.Errorf("unexpected ApprovalRequestStatus: %s", s)
	}
}

func (ApprovalRequest) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":         true,
		"type":       true,
		"status":     true,
		"title":      true,
		"createdat":  true,
		"updatedat":  true,
		"approvedat": true,
	}
	return validSortTypes[sortType]
}
