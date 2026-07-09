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

		CreatedAt: ca,
		UpdatedAt: ua,
	}

	if organization.Description != nil {
		proto.Description = *organization.Description
	}
	if organization.Country != nil {
		proto.Country = *organization.Country
	}
	if organization.City != nil {
		proto.City = *organization.City
	}
	if organization.ContactUserID != nil {
		proto.ContactUserId = *organization.ContactUserID
	}
	if organization.WebsiteURL != nil {
		proto.WebsiteUrl = *organization.WebsiteURL
	}

	return proto, nil
}
