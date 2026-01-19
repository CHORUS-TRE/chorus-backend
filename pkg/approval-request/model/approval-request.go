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

	ApproverIDs  []uint64
	ApprovedByID *uint64

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ApprovedAt *time.Time
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

func GetApprovalRequestStoragePath(requestID uint64) string {
	return fmt.Sprintf("approval-request-%v", requestID)
}

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
