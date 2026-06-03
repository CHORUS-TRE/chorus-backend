package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
)

func TermsOfUseVersionFromBusiness(v *model.TermsOfUseVersion) (*chorus.TermsOfUseVersion, error) {
	ca, err := ToProtoTimestamp(v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.TermsOfUseVersion{
		Id:       v.ID,
		TenantId: v.TenantID,
		Content:  v.Content,
		Status:   TermsOfUseVersionStatusFromBusiness(v.Status),
		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func TermsOfUseAcceptanceFromBusiness(a *model.TermsOfUseAcceptance) (*chorus.TermsOfUseAcceptance, error) {
	aa, err := ToProtoTimestamp(a.AcceptedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert acceptedAt timestamp: %w", err)
	}

	return &chorus.TermsOfUseAcceptance{
		Id:                  a.ID,
		TenantId:            a.TenantID,
		UserId:              a.UserID,
		TermsOfUseVersionId: a.TermsOfUseVersionID,
		AcceptedAt:          aa,
	}, nil
}

func TermsOfUseVersionStatusFromBusiness(s model.TermsOfUseVersionStatus) chorus.TermsOfUseVersionStatus {
	switch s {
	case model.TermsOfUseVersionStatusPublished:
		return chorus.TermsOfUseVersionStatus_TERMS_OF_USE_VERSION_STATUS_PUBLISHED
	case model.TermsOfUseVersionStatusArchived:
		return chorus.TermsOfUseVersionStatus_TERMS_OF_USE_VERSION_STATUS_ARCHIVED
	default:
		return chorus.TermsOfUseVersionStatus_TERMS_OF_USE_VERSION_STATUS_DRAFT
	}
}
