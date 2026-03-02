package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
)

func AuditEntryFromBusiness(auditEntry *model.AuditEntry) (*chorus.AuditEntry, error) {
	createdAt, err := ToProtoTimestamp(auditEntry.CreatedAt)
	if err != nil {
		return nil, err
	}

	details := make(map[string]string, len(auditEntry.Details))
	for k, v := range auditEntry.Details {
		details[k] = fmt.Sprintf("%v", v)
	}

	entry := &chorus.AuditEntry{
		Id: auditEntry.ID,

		ActorId:       auditEntry.ActorID,
		ActorUsername: auditEntry.ActorUsername,
		CorrelationId: auditEntry.CorrelationID,

		Action: string(auditEntry.Action),

		Description: auditEntry.Description,
		Details:     details,

		CreatedAt: createdAt,
	}

	if auditEntry.WorkspaceID != 0 {
		entry.WorkspaceId = &auditEntry.WorkspaceID
	}
	if auditEntry.WorkbenchID != 0 {
		entry.WorkbenchId = &auditEntry.WorkbenchID
	}
	if auditEntry.UserID != 0 {
		entry.UserId = &auditEntry.UserID
	}

	return entry, nil
}

func AuditFilterToBusiness(filter *chorus.AuditFilter) (*model.AuditFilter, error) {
	if filter == nil {
		return nil, nil
	}

	fromTime, err := FromProtoTimestamp(filter.FromTime)
	if err != nil {
		return nil, fmt.Errorf("unable to convert fromTime timestamp: %w", err)
	}

	toTime, err := FromProtoTimestamp(filter.ToTime)
	if err != nil {
		return nil, fmt.Errorf("unable to convert toTime timestamp: %w", err)
	}

	return &model.AuditFilter{
		ActorID: filter.ActorId,

		Action: model.AuditAction(filter.Action),

		WorkspaceID: filter.WorkspaceId,
		WorkbenchID: filter.WorkbenchId,
		UserID:      filter.UserId,

		FromTime: fromTime,
		ToTime:   toTime,
	}, nil
}
