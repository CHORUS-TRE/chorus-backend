package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
)

// OrganizationSummaryFromBusiness converts a business Organization into its wire
// representation without the logo, used for every read path (list, get, and the
// replies of create/update) - see Organization's doc comment in organization.proto.
func OrganizationSummaryFromBusiness(organization *model.Organization) (*chorus.OrganizationSummary, error) {
	ca, err := ToProtoTimestamp(organization.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(organization.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	proto := &chorus.OrganizationSummary{
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

// OrganizationToBusiness converts the wire Organization (the only message that
// carries the logo) into its business representation, for use as the input to
// CreateOrganization/UpdateOrganization. The caller is responsible for setting
// TenantID from the authenticated context - it is never trusted from the client.
func OrganizationToBusiness(organization *chorus.Organization) *model.Organization {
	business := &model.Organization{
		ID:   organization.Id,
		Name: organization.Name,

		Description: nonEmptyString(organization.Description),
		Country:     nonEmptyString(organization.Country),
		City:        nonEmptyString(organization.City),

		ContactUserID: nonZeroUint64(organization.ContactUserId),
		WebsiteURL:    nonEmptyString(organization.WebsiteUrl),
	}

	if len(organization.Logo) > 0 || organization.LogoContentType != "" {
		business.Logo = &model.OrganizationLogo{
			Logo:            organization.Logo,
			LogoContentType: organization.LogoContentType,
		}
	}

	return business
}
