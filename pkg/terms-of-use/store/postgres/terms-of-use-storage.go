package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

var _ service.TermsOfUseStore = (*TermsOfUseStorage)(nil)

type TermsOfUseStorage struct {
	db *sqlx.DB
}

func NewTermsOfUseStorage(db *sqlx.DB) *TermsOfUseStorage {
	return &TermsOfUseStorage{db: db}
}

func (s *TermsOfUseStorage) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	const query = `
		INSERT INTO terms_of_use_versions (tenantid, content, status, createdat, updatedat)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, tenantid, content, status, createdat, updatedat
	`

	var result model.TermsOfUseVersion
	if err := s.db.GetContext(ctx, &result, query, tenantID, content, model.TermsOfUseVersionStatusDraft); err != nil {
		return nil, fmt.Errorf("unable to create terms of use version for tenant %d: %w", tenantID, err)
	}

	return &result, nil
}

func (s *TermsOfUseStorage) UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error) {
	const query = `
		UPDATE terms_of_use_versions
		SET content = $3, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND status = $4
		RETURNING id, tenantid, content, status, createdat, updatedat
	`

	var result model.TermsOfUseVersion
	if err := s.db.GetContext(ctx, &result, query, tenantID, versionID, content, model.TermsOfUseVersionStatusDraft); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("terms of use version %d not found or not in draft status: %w", versionID, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("unable to update terms of use version %d: %w", versionID, err)
	}

	return &result, nil
}

func (s *TermsOfUseStorage) PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	const archiveQuery = `
		UPDATE terms_of_use_versions
		SET status = $2, updatedat = NOW()
		WHERE tenantid = $1 AND status = $3
	`
	if _, err := tx.ExecContext(ctx, archiveQuery, tenantID, model.TermsOfUseVersionStatusArchived, model.TermsOfUseVersionStatusPublished); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("unable to archive published version for tenant %d: %w", tenantID, err)
	}

	const publishQuery = `
		UPDATE terms_of_use_versions
		SET status = $3, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2
		RETURNING id, tenantid, content, status, createdat, updatedat
	`
	var result model.TermsOfUseVersion
	if err := tx.GetContext(ctx, &result, publishQuery, tenantID, versionID, model.TermsOfUseVersionStatusPublished); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("terms of use version %d not found for tenant %d: %w", versionID, tenantID, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("unable to publish terms of use version %d: %w", versionID, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to commit publish transaction: %w", err)
	}

	return &result, nil
}

func (s *TermsOfUseStorage) GetTermsOfUseVersion(ctx context.Context, tenantID uint64, versionID uint64) (*model.TermsOfUseVersion, error) {
	const query = `
		SELECT id, tenantid, content, status, createdat, updatedat
		FROM terms_of_use_versions
		WHERE id = $1 AND tenantid = $2
	`

	var result model.TermsOfUseVersion
	if err := s.db.GetContext(ctx, &result, query, versionID, tenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("terms of use version %d not found: %w", versionID, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("unable to get terms of use version %d: %w", versionID, err)
	}

	return &result, nil
}

func (s *TermsOfUseStorage) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	const countQuery = `SELECT COUNT(*) FROM terms_of_use_versions WHERE tenantid = $1`

	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, tenantID); err != nil {
		return nil, nil, err
	}

	query := `
		SELECT id, tenantid, content, status, createdat, updatedat
		FROM terms_of_use_versions
		WHERE tenantid = $1
	`

	// Add pagination
	clause, validatedPagination := storage.BuildPaginationClause(pagination, model.TermsOfUseVersion{})
	query += clause

	// Build pagination result
	paginationRes := &common_model.PaginationResult{
		Total: uint64(totalCount),
	}

	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var results []*model.TermsOfUseVersion
	if err := s.db.SelectContext(ctx, &results, query, tenantID); err != nil {
		return nil, nil, fmt.Errorf("unable to list terms of use versions for tenant %d: %w", tenantID, err)
	}

	return results, paginationRes, nil
}

func (s *TermsOfUseStorage) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error) {
	const query = `
		SELECT id, tenantid, content, status, createdat, updatedat
		FROM terms_of_use_versions
		WHERE tenantid = $1 AND status = $2
	`

	var result model.TermsOfUseVersion
	if err := s.db.GetContext(ctx, &result, query, tenantID, model.TermsOfUseVersionStatusPublished); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no published terms of use version for tenant %d: %w", tenantID, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("unable to get current terms of use version for tenant %d: %w", tenantID, err)
	}

	return &result, nil
}

func (s *TermsOfUseStorage) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	const query = `
		SELECT id, tenantid, userid, termsofuseversionid, acceptedat
		FROM terms_of_use_acceptances
		WHERE tenantid = $1 AND userid = $2
		ORDER BY acceptedat DESC
	`

	var results []*model.TermsOfUseAcceptance
	if err := s.db.SelectContext(ctx, &results, query, tenantID, userID); err != nil {
		return nil, fmt.Errorf("unable to list terms of use acceptances for user %d: %w", userID, err)
	}

	return results, nil
}

func (s *TermsOfUseStorage) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM terms_of_use_acceptances a
			INNER JOIN terms_of_use_versions v ON v.id = a.termsofuseversionid
			WHERE a.tenantid = $1 AND a.userid = $2 AND v.tenantid = $1 AND v.status = $3
		)
	`

	var accepted bool
	if err := s.db.GetContext(ctx, &accepted, query, tenantID, userID, model.TermsOfUseVersionStatusPublished); err != nil {
		return false, fmt.Errorf("unable to get terms of use status for user %d: %w", userID, err)
	}

	return accepted, nil
}

func (s *TermsOfUseStorage) AcceptTermsOfUse(ctx context.Context, tenantID, userID, versionID uint64) (*model.TermsOfUseAcceptance, error) {
	const query = `
		INSERT INTO terms_of_use_acceptances (tenantid, userid, termsofuseversionid, acceptedat)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, tenantid, userid, termsofuseversionid, acceptedat
	`

	var result model.TermsOfUseAcceptance
	if err := s.db.GetContext(ctx, &result, query, tenantID, userID, versionID); err != nil {
		return nil, fmt.Errorf("unable to accept terms of use for user %d: %w", userID, err)
	}

	return &result, nil
}
