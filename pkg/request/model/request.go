package model

import (
	"fmt"
	"time"
)

type Request struct {
	ID                     uint64
	TenantID               uint64
	RequesterID            uint64
	SourceWorkspaceID      uint64
	DestinationWorkspaceID *uint64

	Type        RequestType
	Status      RequestStatus
	Title       string
	Description string

	Files        []RequestFile
	ApproverIDs  []uint64
	ApprovedByID *uint64

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ApprovedAt *time.Time
}

func (r *Request) IsFinalState() bool {
	return r.Status == RequestStatusApproved ||
		r.Status == RequestStatusRejected ||
		r.Status == RequestStatusCancelled
}

func (r *Request) CanBeDeletedBy(userID uint64) bool {
	return r.RequesterID == userID && !r.IsFinalState()
}

func (r *Request) CanBeApprovedBy(userID uint64) bool {
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

func GetRequestStoragePath(requestID uint64) string {
	return fmt.Sprintf("request%v", requestID)
}

type RequestFile struct {
	SourcePath      string
	DestinationPath string
	Size            uint64
}

type RequestType string

const (
	RequestTypeUnspecified     RequestType = ""
	RequestTypeDownload        RequestType = "download"
	RequestTypeCopyToWorkspace RequestType = "copy_to_workspace"
)

func (t RequestType) String() string {
	return string(t)
}

func ToRequestType(s string) (RequestType, error) {
	switch s {
	case string(RequestTypeDownload), "REQUEST_TYPE_DOWNLOAD":
		return RequestTypeDownload, nil
	case string(RequestTypeCopyToWorkspace), "REQUEST_TYPE_COPY_TO_WORKSPACE":
		return RequestTypeCopyToWorkspace, nil
	case "", "REQUEST_TYPE_UNSPECIFIED":
		return RequestTypeUnspecified, nil
	default:
		return RequestTypeUnspecified, fmt.Errorf("unexpected RequestType: %s", s)
	}
}

type RequestStatus string

const (
	RequestStatusUnspecified RequestStatus = ""
	RequestStatusPending     RequestStatus = "pending"
	RequestStatusApproved    RequestStatus = "approved"
	RequestStatusRejected    RequestStatus = "rejected"
	RequestStatusCancelled   RequestStatus = "cancelled"
)

func (s RequestStatus) String() string {
	return string(s)
}

func ToRequestStatus(status string) (RequestStatus, error) {
	switch status {
	case string(RequestStatusPending), "REQUEST_STATUS_PENDING":
		return RequestStatusPending, nil
	case string(RequestStatusApproved), "REQUEST_STATUS_APPROVED":
		return RequestStatusApproved, nil
	case string(RequestStatusRejected), "REQUEST_STATUS_REJECTED":
		return RequestStatusRejected, nil
	case string(RequestStatusCancelled), "REQUEST_STATUS_CANCELLED":
		return RequestStatusCancelled, nil
	case "", "REQUEST_STATUS_UNSPECIFIED":
		return RequestStatusUnspecified, nil
	default:
		return RequestStatusUnspecified, fmt.Errorf("unexpected RequestStatus: %s", status)
	}
}
