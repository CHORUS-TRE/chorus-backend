package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"
)

var _ service.OrganizationStore = (*OrganizationStorage)(nil)

// organizationColumns lists every column returned for a full Organization row,
// deliberately excluding logo/logocontenttype: the logo is served separately
// through GetOrganizationLogo to avoid pulling large blobs into list/get responses.
const organizationColumns = `
	id, tenantid, name, description, country, city, contactuserid, websiteurl,
	createdat, updatedat
`

const getOrganizationQuery = `
	SELECT ` + organizationColumns + `
	FROM organizations
	WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
`

const listOrganizationsQuery = `
	SELECT ` + organizationColumns + `
	FROM organizations
	WHERE tenantid = $1 AND deletedat IS NULL
`

const createOrganizationQuery = `
	INSERT INTO organizations (tenantid, name, description, logo, logocontenttype, country, city, contactuserid, websiteurl, createdat, updatedat)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	RETURNING ` + organizationColumns + `;
`

// OrganizationStorage is the handler through which a PostgresDB backend can be queried.
type OrganizationStorage struct {
	db *sqlx.DB
}

// NewOrganizationStorage returns a fresh organization storage instance.
func NewOrganizationStorage(db *sqlx.DB) *OrganizationStorage {
	return &OrganizationStorage{db: db}
}

func (s *OrganizationStorage) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	var organization model.Organization
	if err := s.db.GetContext(ctx, &organization, getOrganizationQuery, tenantID, id); err != nil {
		return nil, err
	}

	return &organization, nil
}

func (s *OrganizationStorage) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	const query = `
		SELECT logo, logocontenttype
		FROM organizations
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var logo model.OrganizationLogo
	if err := s.db.GetContext(ctx, &logo, query, tenantID, id); err != nil {
		return nil, err
	}
	if logo.LogoContentType == "" {
		return nil, nil
	}

	return &logo, nil
}

func (s *OrganizationStorage) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Organization, *common_model.PaginationResult, error) {
	const countQuery = `SELECT COUNT(*) FROM organizations WHERE tenantid = $1 AND deletedat IS NULL`

	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, tenantID); err != nil {
		return nil, nil, err
	}

	clause, validatedPagination := storage.BuildPaginationClause(pagination, model.Organization{})
	query := listOrganizationsQuery + clause

	paginationRes := &common_model.PaginationResult{Total: uint64(totalCount)}
	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var organizations []*model.Organization
	if err := s.db.SelectContext(ctx, &organizations, query, tenantID); err != nil {
		return nil, nil, err
	}

	return organizations, paginationRes, nil
}

func (s *OrganizationStorage) CreateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	logo, logoContentType := organization.Logo.Unwrap()

	var created model.Organization
	err := s.db.GetContext(ctx, &created, createOrganizationQuery,
		tenantID, organization.Name, organization.Description, logo, logoContentType,
		organization.Country, organization.City, organization.ContactUserID, organization.WebsiteURL,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create organization: %w", err)
	}

	return &created, nil
}

func (s *OrganizationStorage) UpdateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	setLogoClause := ""
	args := []interface{}{
		tenantID, organization.ID, organization.Name, organization.Description,
		organization.Country, organization.City, organization.ContactUserID, organization.WebsiteURL,
	}
	// A nil Logo means "not provided, leave the existing logo untouched" (see
	// model.Organization.Logo's doc comment).
	if organization.Logo != nil {
		logo, logoContentType := organization.Logo.Unwrap()
		setLogoClause = ", logo = $9, logocontenttype = $10"
		args = append(args, logo, logoContentType)
	}

	query := fmt.Sprintf(`
		UPDATE organizations
		SET name = $3, description = $4, country = $5, city = $6, contactuserid = $7, websiteurl = $8,
		    updatedat = NOW()%s
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
		RETURNING %s;
	`, setLogoClause, organizationColumns)

	var updated model.Organization
	if err := s.db.GetContext(ctx, &updated, query, args...); err != nil {
		return nil, fmt.Errorf("unable to update organization: %w", err)
	}

	return &updated, nil
}

func (s *OrganizationStorage) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	const query = `
		UPDATE organizations
		SET deletedat = NOW(), updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	result, err := s.db.ExecContext(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("unable to delete organization: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to get rows affected: %w", err)
	}
	if affected == 0 {
		return cerr.ErrNoRowsDeleted
	}

	return nil
}
