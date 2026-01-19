package converter

import (
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
)

func ApprovalRequestFromBusiness(request *model.ApprovalRequest) (*chorus.ApprovalRequest, error) {
	if request == nil {
		return nil, nil
	}

	ca, err := ToProtoTimestamp(request.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(request.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}
	aa, err := PointerToProtoTimestamp(request.ApprovedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert approvedAt timestamp: %w", err)
	}

	var files []*chorus.ApprovalRequestFile
	for _, f := range request.Files {
		files = append(files, &chorus.ApprovalRequestFile{
			SourcePath:      f.SourcePath,
			DestinationPath: f.DestinationPath,
			Size:            f.Size,
		})
	}

	protoRequest := &chorus.ApprovalRequest{
		Id:          request.ID,
		TenantId:    request.TenantID,
		RequesterId: request.RequesterID,
		Type:        ApprovalRequestTypeFromBusiness(request.Type),
		Status:      ApprovalRequestStatusFromBusiness(request.Status),
		Title:       request.Title,
		Description: request.Description,
		Files:       files,
		ApproverIds: request.ApproverIDs,
		CreatedAt:   ca,
		UpdatedAt:   ua,
		ApprovedAt:  aa,
	}

	if request.ApprovedByID != nil {
		protoRequest.ApprovedById = request.ApprovedByID
	}

	switch details := request.Details.(type) {
	case model.DataExtractionDetails:
		protoRequest.Details = &chorus.ApprovalRequest_DataExtraction{
			DataExtraction: &chorus.DataExtractionDetails{
				SourceWorkspaceId: details.SourceWorkspaceID,
			},
		}
	case model.DataTransferDetails:
		protoRequest.Details = &chorus.ApprovalRequest_DataTransfer{
			DataTransfer: &chorus.DataTransferDetails{
				SourceWorkspaceId:      details.SourceWorkspaceID,
				DestinationWorkspaceId: details.DestinationWorkspaceID,
			},
		}
	}

	return protoRequest, nil
}

func ApprovalRequestToBusiness(request *chorus.ApprovalRequest) (*model.ApprovalRequest, error) {
	if request == nil {
		return nil, nil
	}

	ca, err := FromProtoTimestamp(request.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(request.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	var approvedAt *time.Time
	if request.ApprovedAt != nil {
		aa, err := FromProtoTimestamp(request.ApprovedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to convert approvedAt timestamp: %w", err)
		}
		approvedAt = &aa
	}

	var files []model.ApprovalRequestFile
	for _, f := range request.Files {
		files = append(files, model.ApprovalRequestFile{
			SourcePath:      f.SourcePath,
			DestinationPath: f.DestinationPath,
			Size:            f.Size,
		})
	}

	result := &model.ApprovalRequest{
		ID:           request.Id,
		TenantID:     request.TenantId,
		RequesterID:  request.RequesterId,
		Type:         ApprovalRequestTypeToBusiness(request.Type),
		Status:       ApprovalRequestStatusToBusiness(request.Status),
		Title:        request.Title,
		Description:  request.Description,
		Files:        files,
		ApproverIDs:  request.ApproverIds,
		ApprovedByID: request.ApprovedById,
		CreatedAt:    ca,
		UpdatedAt:    ua,
		ApprovedAt:   approvedAt,
	}

	switch d := request.Details.(type) {
	case *chorus.ApprovalRequest_DataExtraction:
		if d.DataExtraction != nil {
			result.Details = model.DataExtractionDetails{
				SourceWorkspaceID: d.DataExtraction.SourceWorkspaceId,
			}
		}
	case *chorus.ApprovalRequest_DataTransfer:
		if d.DataTransfer != nil {
			result.Details = model.DataTransferDetails{
				SourceWorkspaceID:      d.DataTransfer.SourceWorkspaceId,
				DestinationWorkspaceID: d.DataTransfer.DestinationWorkspaceId,
			}
		}
	}

	return result, nil
}

func ApprovalRequestTypeFromBusiness(t model.ApprovalRequestType) chorus.ApprovalRequestType {
	switch t {
	case model.ApprovalRequestTypeDataExtraction:
		return chorus.ApprovalRequestType_APPROVAL_REQUEST_TYPE_DATA_EXTRACTION
	case model.ApprovalRequestTypeDataTransfer:
		return chorus.ApprovalRequestType_APPROVAL_REQUEST_TYPE_DATA_TRANSFER
	default:
		return chorus.ApprovalRequestType_APPROVAL_REQUEST_TYPE_UNSPECIFIED
	}
}

func ApprovalRequestTypeToBusiness(t chorus.ApprovalRequestType) model.ApprovalRequestType {
	switch t {
	case chorus.ApprovalRequestType_APPROVAL_REQUEST_TYPE_DATA_EXTRACTION:
		return model.ApprovalRequestTypeDataExtraction
	case chorus.ApprovalRequestType_APPROVAL_REQUEST_TYPE_DATA_TRANSFER:
		return model.ApprovalRequestTypeDataTransfer
	default:
		return model.ApprovalRequestTypeUnspecified
	}
}

func ApprovalRequestStatusFromBusiness(s model.ApprovalRequestStatus) chorus.ApprovalRequestStatus {
	switch s {
	case model.ApprovalRequestStatusPending:
		return chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_PENDING
	case model.ApprovalRequestStatusApproved:
		return chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_APPROVED
	case model.ApprovalRequestStatusRejected:
		return chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_REJECTED
	case model.ApprovalRequestStatusCancelled:
		return chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_CANCELLED
	default:
		return chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_UNSPECIFIED
	}
}

func ApprovalRequestStatusToBusiness(s chorus.ApprovalRequestStatus) model.ApprovalRequestStatus {
	switch s {
	case chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_PENDING:
		return model.ApprovalRequestStatusPending
	case chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_APPROVED:
		return model.ApprovalRequestStatusApproved
	case chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_REJECTED:
		return model.ApprovalRequestStatusRejected
	case chorus.ApprovalRequestStatus_APPROVAL_REQUEST_STATUS_CANCELLED:
		return model.ApprovalRequestStatusCancelled
	default:
		return model.ApprovalRequestStatusUnspecified
	}
}
