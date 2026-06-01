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

	// ApproverIDs lists the user IDs allowed to approve each arm of the
	// request, keyed by arm name (see ApprovalRequestArm* constants).
	// A user who appears in every arm can approve the whole request in one
	// step; otherwise each arm must be approved separately by a user
	// authorized for that arm.
	ApproverIDs map[string][]uint64

	// ArmApprovals records the per-arm approval decisions made so far.
	// The key is the arm name. A request is fully approved once every arm
	// it requires has an Approve=true entry; a single Approve=false entry
	// rejects the whole request.
	ArmApprovals map[string]ArmApproval

	ApprovedByID    *uint64
	AutoApproved    bool
	ApprovalMessage string

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ApprovedAt *time.Time
}

// ArmApproval records a single per-arm approval decision.
type ArmApproval struct {
	ApproverID uint64    `json:"approver_id"`
	ApprovedAt time.Time `json:"approved_at"`
	Approve    bool      `json:"approve"`
}

// Arm names used to partition the set of approvers for a request.
const (
	ApprovalRequestArmDownload = "download"
	ApprovalRequestArmUpload   = "upload"
)

// ArmsForType returns the ordered list of arm names that must be approved
// for a request of the given type.
func ArmsForType(t ApprovalRequestType) []string {
	switch t {
	case ApprovalRequestTypeDataExtraction:
		return []string{ApprovalRequestArmDownload}
	case ApprovalRequestTypeDataTransfer:
		return []string{ApprovalRequestArmDownload, ApprovalRequestArmUpload}
	default:
		return nil
	}
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
	return len(r.ArmsToApprove(userID)) > 0
}

// ArmsToApprove returns the names of the arms the given user is authorized
// to approve and that have not yet been decided. The result is empty if the
// user has nothing to approve on this request.
func (r *ApprovalRequest) ArmsToApprove(userID uint64) []string {
	if r.IsFinalState() {
		return nil
	}

	arms := ArmsForType(r.Type)
	// If no arms are declared (e.g. legacy/unknown type) but approver IDs
	// are set, treat the union of all approver IDs as a single implicit arm.
	if len(arms) == 0 {
		if len(r.ApproverIDs) == 0 {
			return nil
		}
		for arm := range r.ApproverIDs {
			arms = append(arms, arm)
		}
	}

	var pending []string
	for _, arm := range arms {
		if _, decided := r.ArmApprovals[arm]; decided {
			continue
		}
		if r.userIsApproverOf(userID, arm) {
			pending = append(pending, arm)
		}
	}
	return pending
}

func (r *ApprovalRequest) userIsApproverOf(userID uint64, arm string) bool {
	approvers, ok := r.ApproverIDs[arm]
	if !ok {
		return false
	}
	// An empty approver list for an arm allows anyone to approve that arm
	// (matches the legacy behaviour where an empty list was permissive).
	if len(approvers) == 0 {
		return true
	}
	for _, id := range approvers {
		if id == userID {
			return true
		}
	}
	return false
}

// IsFullyApproved reports whether every required arm has been approved.
func (r *ApprovalRequest) IsFullyApproved() bool {
	arms := ArmsForType(r.Type)
	if len(arms) == 0 {
		return false
	}
	for _, arm := range arms {
		decision, ok := r.ArmApprovals[arm]
		if !ok || !decision.Approve {
			return false
		}
	}
	return true
}

// HasArmRejection reports whether any arm has been explicitly rejected.
func (r *ApprovalRequest) HasArmRejection() bool {
	for _, decision := range r.ArmApprovals {
		if !decision.Approve {
			return true
		}
	}
	return false
}

// AllApproverIDs returns the deduplicated union of every approver across all arms.
func (r *ApprovalRequest) AllApproverIDs() []uint64 {
	seen := make(map[uint64]struct{})
	var ids []uint64
	for _, approvers := range r.ApproverIDs {
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
