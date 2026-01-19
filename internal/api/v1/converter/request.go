package converter

import (
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/model"
)

func RequestFromBusiness(request *model.Request) (*chorus.Request, error) {
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

	var files []*chorus.RequestFile
	for _, f := range request.Files {
		files = append(files, &chorus.RequestFile{
			SourcePath:      f.SourcePath,
			DestinationPath: f.DestinationPath,
			Size:            f.Size,
		})
	}

	protoRequest := &chorus.Request{
		Id:                request.ID,
		TenantId:          request.TenantID,
		RequesterId:       request.RequesterID,
		SourceWorkspaceId: request.SourceWorkspaceID,
		Type:              RequestTypeFromBusiness(request.Type),
		Status:            RequestStatusFromBusiness(request.Status),
		Title:             request.Title,
		Description:       request.Description,
		Files:             files,
		ApproverIds:       request.ApproverIDs,
		CreatedAt:         ca,
		UpdatedAt:         ua,
		ApprovedAt:        aa,
	}

	if request.DestinationWorkspaceID != nil {
		protoRequest.DestinationWorkspaceId = request.DestinationWorkspaceID
	}
	if request.ApprovedByID != nil {
		protoRequest.ApprovedById = request.ApprovedByID
	}

	return protoRequest, nil
}

func RequestToBusiness(request *chorus.Request) (*model.Request, error) {
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

	var files []model.RequestFile
	for _, f := range request.Files {
		files = append(files, model.RequestFile{
			SourcePath:      f.SourcePath,
			DestinationPath: f.DestinationPath,
			Size:            f.Size,
		})
	}

	return &model.Request{
		ID:                     request.Id,
		TenantID:               request.TenantId,
		RequesterID:            request.RequesterId,
		SourceWorkspaceID:      request.SourceWorkspaceId,
		DestinationWorkspaceID: request.DestinationWorkspaceId,
		Type:                   RequestTypeToBusiness(request.Type),
		Status:                 RequestStatusToBusiness(request.Status),
		Title:                  request.Title,
		Description:            request.Description,
		Files:                  files,
		ApproverIDs:            request.ApproverIds,
		ApprovedByID:           request.ApprovedById,
		CreatedAt:              ca,
		UpdatedAt:              ua,
		ApprovedAt:             approvedAt,
	}, nil
}

func RequestTypeFromBusiness(t model.RequestType) chorus.RequestType {
	switch t {
	case model.RequestTypeDownload:
		return chorus.RequestType_REQUEST_TYPE_DOWNLOAD
	case model.RequestTypeCopyToWorkspace:
		return chorus.RequestType_REQUEST_TYPE_COPY_TO_WORKSPACE
	default:
		return chorus.RequestType_REQUEST_TYPE_UNSPECIFIED
	}
}

func RequestTypeToBusiness(t chorus.RequestType) model.RequestType {
	switch t {
	case chorus.RequestType_REQUEST_TYPE_DOWNLOAD:
		return model.RequestTypeDownload
	case chorus.RequestType_REQUEST_TYPE_COPY_TO_WORKSPACE:
		return model.RequestTypeCopyToWorkspace
	default:
		return model.RequestTypeUnspecified
	}
}

func RequestStatusFromBusiness(s model.RequestStatus) chorus.RequestStatus {
	switch s {
	case model.RequestStatusPending:
		return chorus.RequestStatus_REQUEST_STATUS_PENDING
	case model.RequestStatusApproved:
		return chorus.RequestStatus_REQUEST_STATUS_APPROVED
	case model.RequestStatusRejected:
		return chorus.RequestStatus_REQUEST_STATUS_REJECTED
	case model.RequestStatusCancelled:
		return chorus.RequestStatus_REQUEST_STATUS_CANCELLED
	default:
		return chorus.RequestStatus_REQUEST_STATUS_UNSPECIFIED
	}
}

func RequestStatusToBusiness(s chorus.RequestStatus) model.RequestStatus {
	switch s {
	case chorus.RequestStatus_REQUEST_STATUS_PENDING:
		return model.RequestStatusPending
	case chorus.RequestStatus_REQUEST_STATUS_APPROVED:
		return model.RequestStatusApproved
	case chorus.RequestStatus_REQUEST_STATUS_REJECTED:
		return model.RequestStatusRejected
	case chorus.RequestStatus_REQUEST_STATUS_CANCELLED:
		return model.RequestStatusCancelled
	default:
		return model.RequestStatusUnspecified
	}
}
