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

type stubOrganizationer struct{}

func (stubOrganizationer) ListOrganizations(_ context.Context, _ service.ListOrganizationsReq) ([]*model.Organization, *common_model.PaginationResult, error) {
	return nil, &common_model.PaginationResult{}, nil
}
func (stubOrganizationer) GetOrganization(_ context.Context, _ service.GetOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizationer) GetOrganizationLogo(_ context.Context, _ service.GetOrganizationLogoReq) ([]byte, *string, error) {
	return nil, nil, nil
}
func (stubOrganizationer) CreateOrganization(_ context.Context, _ service.CreateOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizationer) UpdateOrganization(_ context.Context, _ service.UpdateOrganizationReq) (*model.Organization, error) {
	return &model.Organization{}, nil
}
func (stubOrganizationer) DeleteOrganization(_ context.Context, _ service.DeleteOrganizationReq) error {
	return nil
}

func newTestValidationOrganizationer() service.Organizationer {
	return Validation(chorus_validation.NewValidator())(stubOrganizationer{})
}

func ptr[T any](v T) *T { return &v }

func TestValidation_CreateOrganization_ValidRequestPasses(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "CHUV",
		Description:     ptr("A description"),
		Logo:            []byte{0x89, 0x50, 0x4E, 0x47},
		LogoContentType: ptr("image/png"),
		Country:         ptr("CH"),
		City:            ptr("Lausanne"),
		WebsiteURL:      ptr("https://www.chuv.ch/"),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_NameRequired(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_DescriptionOver250CharsRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:    1,
		Name:        "CHUV",
		Description: ptr(strings.Repeat("a", 251)),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_DescriptionAt250CharsAccepted(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:    1,
		Name:        "CHUV",
		Description: ptr(strings.Repeat("a", 250)),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_InvalidCountryCodeRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	for _, country := range []string{"ZZ", "ch", "SUI", "1A"} {
		_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
			TenantID: 1,
			Name:     "CHUV",
			Country:  ptr(country),
		})
		require.Errorf(t, err, "country %q should have been rejected", country)
	}
}

func TestValidation_CreateOrganization_ValidCountryCodeAccepted(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	for _, country := range []string{"CH", "FR", "US"} {
		_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
			TenantID: 1,
			Name:     "CHUV",
			Country:  ptr(country),
		})
		require.NoErrorf(t, err, "country %q should have been accepted", country)
	}
}

func TestValidation_CreateOrganization_CityOver100CharsRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID: 1,
		Name:     "CHUV",
		City:     ptr(strings.Repeat("a", 101)),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_WebsiteURLOver2048CharsRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	longURL := "https://example.com/" + strings.Repeat("a", 2048)
	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:   1,
		Name:       "CHUV",
		WebsiteURL: ptr(longURL),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_MalformedWebsiteURLRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:   1,
		Name:       "CHUV",
		WebsiteURL: ptr("not a url"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoOver512KBRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "CHUV",
		Logo:            make([]byte, 512*1024+1),
		LogoContentType: ptr("image/png"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoAt512KBAccepted(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "CHUV",
		Logo:            make([]byte, 512*1024),
		LogoContentType: ptr("image/png"),
	})

	require.NoError(t, err)
}

func TestValidation_CreateOrganization_UnsupportedLogoContentTypeRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "CHUV",
		Logo:            []byte{0x01},
		LogoContentType: ptr("application/pdf"),
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_LogoWithoutContentTypeRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID: 1,
		Name:     "CHUV",
		Logo:     []byte{0x01},
	})

	require.Error(t, err)
}

func TestValidation_CreateOrganization_ContentTypeWithoutLogoRejected(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.CreateOrganization(context.Background(), service.CreateOrganizationReq{
		TenantID:        1,
		Name:            "CHUV",
		LogoContentType: ptr("image/png"),
	})

	require.Error(t, err)
}

func TestValidation_UpdateOrganization_RequiresID(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.UpdateOrganization(context.Background(), service.UpdateOrganizationReq{
		Name: "CHUV",
	})

	require.Error(t, err)
}

func TestValidation_UpdateOrganization_OmittedLogoPassesValidation(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.UpdateOrganization(context.Background(), service.UpdateOrganizationReq{
		TenantID: 1,
		ID:       1,
		Name:     "CHUV",
	})

	require.NoError(t, err)
}

func TestValidation_GetOrganization_RequiresID(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, err := organizationer.GetOrganization(context.Background(), service.GetOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_DeleteOrganization_RequiresID(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	err := organizationer.DeleteOrganization(context.Background(), service.DeleteOrganizationReq{TenantID: 1})

	require.Error(t, err)
}

func TestValidation_ListOrganizations_NoFieldsRequired(t *testing.T) {
	organizationer := newTestValidationOrganizationer()

	_, _, err := organizationer.ListOrganizations(context.Background(), service.ListOrganizationsReq{TenantID: 1})

	assert.NoError(t, err)
}
