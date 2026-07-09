//go:build unit

package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	chorus_validation "github.com/CHORUS-TRE/chorus-backend/internal/utils/validation"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"
)

type stubOrganizer struct{}

func (stubOrganizer) ListOrganizations(_ context.Context, _ service.ListOrganizationsReq) ([]*model.Organization, *common_model.PaginationResult, error) {
	return nil, &common_model.PaginationResult{}, nil
}
func (stubOrganizer) GetOrganization(_ context.Context, _ service.GetOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizer) GetOrganizationLogo(_ context.Context, _ service.GetOrganizationLogoReq) ([]byte, *string, error) {
	return nil, nil, nil
}
func (stubOrganizer) CreateOrganization(_ context.Context, _ service.CreateOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizer) UpdateOrganization(_ context.Context, _ service.UpdateOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizer) DeleteOrganization(_ context.Context, _ service.DeleteOrganizationReq) error {
	return nil
}

func newTestValidationOrganizer() service.Organizer {
	return Validation(chorus_validation.NewValidator())(stubOrganizer{})
}

func ptr[T any](v T) *T { return &v }

func TestValidation_CreateOrganization_ValidRequestPasses(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "ACME",
		Description:     ptr("A description"),
		Logo:            []byte{0x89, 0x50, 0x4E, 0x47},
		LogoContentType: ptr("image/png"),
		Country:         ptr("CH"),
		City:            ptr("Lausanne"),
		WebsiteURL:      ptr("https://acme.example.com"),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_NameRequired(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_DescriptionOver250CharsRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:    1,
		Name:        "ACME",
		Description: ptr(strings.Repeat("a", 251)),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_DescriptionAt250CharsAccepted(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:    1,
		Name:        "ACME",
		Description: ptr(strings.Repeat("a", 250)),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_InvalidCountryCodeRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	for _, country := range []string{"ZZ", "ch", "SUI", "1A"} {
		_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
			TenantID: 1,
			Name:     "ACME",
			Country:  ptr(country),
		})
		require.Errorf(t, err, "country %q should have been rejected", country)
	}
}

func TestValidation_CreateOrganization_ValidCountryCodeAccepted(t *testing.T) {
	organizer := newTestValidationOrganizer()

	for _, country := range []string{"CH", "FR", "US"} {
		_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
			TenantID: 1,
			Name:     "ACME",
			Country:  ptr(country),
		})
		require.NoErrorf(t, err, "country %q should have been accepted", country)
	}
}

func TestValidation_CreateOrganization_CityOver100CharsRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID: 1,
		Name:     "ACME",
		City:     ptr(strings.Repeat("a", 101)),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_WebsiteURLOver2048CharsRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	longURL := "https://example.com/" + strings.Repeat("a", 2048)
	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:   1,
		Name:       "ACME",
		WebsiteURL: ptr(longURL),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_MalformedWebsiteURLRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:   1,
		Name:       "ACME",
		WebsiteURL: ptr("not a url"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoOver512KBRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "ACME",
		Logo:            make([]byte, 512*1024+1),
		LogoContentType: ptr("image/png"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoAt512KBAccepted(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "ACME",
		Logo:            make([]byte, 512*1024),
		LogoContentType: ptr("image/png"),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_UnsupportedLogoContentTypeRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "ACME",
		Logo:            []byte{0x01},
		LogoContentType: ptr("application/pdf"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoWithoutContentTypeRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID: 1,
		Name:     "ACME",
		Logo:     []byte{0x01},
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_ContentTypeWithoutLogoRejected(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "ACME",
		LogoContentType: ptr("image/png"),
	})

	require.Error(t, err)
}

func TestValidation_UpdateOrganization_RequiresID(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.UpdateOrganization(context.Background(), service.UpdateOrganizationReq{
		Name: "ACME",
	})

	require.Error(t, err)
}

func TestValidation_UpdateOrganization_OmittedLogoPassesValidation(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.UpdateOrganization(context.Background(), service.UpdateOrganizationReq{
		TenantID: 1,
		ID:       1,
		Name:     "ACME",
	})

	require.NoError(t, err)
}

func TestValidation_GetOrganization_RequiresID(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, err := organizer.GetOrganization(context.Background(), service.GetOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_DeleteOrganization_RequiresID(t *testing.T) {
	organizer := newTestValidationOrganizer()

	err := organizer.DeleteOrganization(context.Background(), service.DeleteOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_ListOrganizations_NoFieldsRequired(t *testing.T) {
	organizer := newTestValidationOrganizer()

	_, _, err := organizer.ListOrganizations(context.Background(), service.ListOrganizationsReq{TenantID: 1})

	assert.NoError(t, err)
}
