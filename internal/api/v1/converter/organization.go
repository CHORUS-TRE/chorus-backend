package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
)

func OrganizationFromBusiness(organization *model.Organization) (*chorus.Organization, error) {
	ca, err := ToProtoTimestamp(organization.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(organization.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	proto := &chorus.Organization{
		Id:       organization.ID,
		TenantId: organization.TenantID,
		Name:     organization.Name,

		Description:   derefOrZero(organization.Description),
		Country:       derefOrZero(organization.Country),
		City:          derefOrZero(organization.City),
		ContactUserId: derefOrZero(organization.ContactUserID),
		WebsiteUrl:    derefOrZero(organization.WebsiteURL),

		CreatedAt: ca,
		UpdatedAt: ua,
	}

	return proto, nil
}
