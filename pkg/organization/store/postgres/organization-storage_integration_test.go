//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	integration "github.com/CHORUS-TRE/chorus-backend/tests/integration/postgres"
)

const (
	orgTestTenantID = uint64(88300)
	orgTestUserID   = uint64(88301)
)

func setupOrganizationFixtures(t *testing.T, db *sqlx.DB) {
	t.Helper()
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO tenants (id, name, createdat, updatedat) VALUES ($1, 'org_test_tenant', NOW(), NOW())`, orgTestTenantID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, tenantid, firstname, lastname, username, status, createdat, updatedat)
		VALUES ($1, $2, 'contact', 'person', 'org_test_contact', 'active', NOW(), NOW())
	`, orgTestUserID, orgTestTenantID)
	require.NoError(t, err)
}

func ptr[T any](v T) *T { return &v }

func TestOrganizationStorage_CreateAndGetOrganization(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{
		Name:          "CHUV",
		Description:   ptr("A description"),
		Logo:          &model.OrganizationLogo{Logo: []byte{0x89, 0x50, 0x4E, 0x47}, LogoContentType: "image/png"},
		Country:       ptr("CH"),
		City:          ptr("Lausanne"),
		ContactUserID: ptr(orgTestUserID),
		WebsiteURL:    ptr("https://www.chuv.ch/"),
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "CHUV", created.Name)
	require.Equal(t, "CH", *created.Country)
	require.Equal(t, orgTestUserID, *created.ContactUserID)
	require.Nil(t, created.Logo, "Create must not return the logo bytes, consistent with Get/List")

	fetched, err := store.GetOrganization(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, "CHUV", fetched.Name)
	require.Nil(t, fetched.Logo, "GetOrganization must not return the logo bytes")

	logo, err := store.GetOrganizationLogo(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err)
	require.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, logo.Logo)
	require.Equal(t, "image/png", logo.LogoContentType)
}

func TestOrganizationStorage_GetOrganizationLogo_NullColumnScansCleanly(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: "CHUV"})
	require.NoError(t, err)

	logo, err := store.GetOrganizationLogo(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err, "a NULL logo column must scan cleanly into []byte, not error like it would for a plain string")
	require.Nil(t, logo, "no logo uploaded must report as nil, consistent with model.Organization.Logo's nil-means-absent convention")
}

func TestOrganizationStorage_GetOrganization_WrongTenantNotFound(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: "CHUV"})
	require.NoError(t, err)

	_, err = store.GetOrganization(ctx, orgTestTenantID+1, created.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestOrganizationStorage_ListOrganizations_ExcludesSoftDeleted(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	kept, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: "Kept Org"})
	require.NoError(t, err)
	deleted, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: "Deleted Org"})
	require.NoError(t, err)

	require.NoError(t, store.DeleteOrganization(ctx, orgTestTenantID, deleted.ID))

	organizations, pagination, err := store.ListOrganizations(ctx, orgTestTenantID, nil)
	require.NoError(t, err)
	require.Len(t, organizations, 1)
	require.Equal(t, kept.ID, organizations[0].ID)
	require.Equal(t, uint64(1), pagination.Total)

	_, err = store.GetOrganization(ctx, orgTestTenantID, deleted.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestOrganizationStorage_ListOrganizations_Pagination(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	for _, name := range []string{"Org A", "Org B", "Org C"} {
		_, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: name})
		require.NoError(t, err)
	}

	organizations, pagination, err := store.ListOrganizations(ctx, orgTestTenantID, &common_model.Pagination{Limit: 2, Offset: 0})
	require.NoError(t, err)
	require.Len(t, organizations, 2)
	require.Equal(t, uint64(3), pagination.Total)
	require.Equal(t, uint64(2), pagination.Limit)
}

func TestOrganizationStorage_UpdateOrganization_WithoutLogoPreservesExisting(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{
		Name: "CHUV",
		Logo: &model.OrganizationLogo{Logo: []byte{0x01, 0x02, 0x03}, LogoContentType: "image/png"},
	})
	require.NoError(t, err)

	updated, err := store.UpdateOrganization(ctx, orgTestTenantID, &model.Organization{
		ID:   created.ID,
		Name: "CHUV Renamed",
	})
	require.NoError(t, err)
	require.Equal(t, "CHUV Renamed", updated.Name)

	logo, err := store.GetOrganizationLogo(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0x02, 0x03}, logo.Logo, "logo must be preserved when the update omits Logo")
	require.Equal(t, "image/png", logo.LogoContentType)
}

// TestOrganizationStorage_UpdateOrganization_OmittedOptionalFieldsAreCleared documents
// that, unlike Logo, every other optional field follows full-replace PUT semantics:
// omitting a field on update clears it rather than preserving the existing value.
func TestOrganizationStorage_UpdateOrganization_OmittedOptionalFieldsAreCleared(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{
		Name:          "CHUV",
		Description:   ptr("A description"),
		Country:       ptr("CH"),
		City:          ptr("Lausanne"),
		ContactUserID: ptr(orgTestUserID),
		WebsiteURL:    ptr("https://www.chuv.ch/"),
	})
	require.NoError(t, err)

	updated, err := store.UpdateOrganization(ctx, orgTestTenantID, &model.Organization{
		ID:   created.ID,
		Name: "CHUV Renamed",
	})
	require.NoError(t, err)
	require.Equal(t, "CHUV Renamed", updated.Name)
	require.Nil(t, updated.Description, "omitted Description must be cleared, not preserved")
	require.Nil(t, updated.Country, "omitted Country must be cleared, not preserved")
	require.Nil(t, updated.City, "omitted City must be cleared, not preserved")
	require.Nil(t, updated.ContactUserID, "omitted ContactUserID must be cleared, not preserved")
	require.Nil(t, updated.WebsiteURL, "omitted WebsiteURL must be cleared, not preserved")

	fetched, err := store.GetOrganization(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err)
	require.Nil(t, fetched.Description)
	require.Nil(t, fetched.Country)
	require.Nil(t, fetched.City)
	require.Nil(t, fetched.ContactUserID)
	require.Nil(t, fetched.WebsiteURL)
}

func TestOrganizationStorage_UpdateOrganization_WithLogoReplacesExisting(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{
		Name: "CHUV",
		Logo: &model.OrganizationLogo{Logo: []byte{0x01, 0x02, 0x03}, LogoContentType: "image/png"},
	})
	require.NoError(t, err)

	_, err = store.UpdateOrganization(ctx, orgTestTenantID, &model.Organization{
		ID:   created.ID,
		Name: "CHUV",
		Logo: &model.OrganizationLogo{Logo: []byte{0xFF, 0xD8, 0xFF}, LogoContentType: "image/jpeg"},
	})
	require.NoError(t, err)

	logo, err := store.GetOrganizationLogo(ctx, orgTestTenantID, created.ID)
	require.NoError(t, err)
	require.Equal(t, []byte{0xFF, 0xD8, 0xFF}, logo.Logo)
	require.Equal(t, "image/jpeg", logo.LogoContentType)
}

func TestOrganizationStorage_DeleteOrganization_NotFoundReturnsErrNoRowsDeleted(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() { integration.CleanupTables(db) })

	setupOrganizationFixtures(t, db)
	store := NewOrganizationStorage(db)
	ctx := context.Background()

	created, err := store.CreateOrganization(ctx, orgTestTenantID, &model.Organization{Name: "CHUV"})
	require.NoError(t, err)

	require.NoError(t, store.DeleteOrganization(ctx, orgTestTenantID, created.ID))

	err = store.DeleteOrganization(ctx, orgTestTenantID, created.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, cerr.ErrNoRowsDeleted))
}
