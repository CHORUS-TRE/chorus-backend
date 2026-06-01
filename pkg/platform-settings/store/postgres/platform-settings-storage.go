package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

var _ service.PlatformSettingsStore = (*PlatformSettingsStorage)(nil)

type PlatformSettingsStorage struct {
	db *sqlx.DB
}

func NewPlatformSettingsStorage(db *sqlx.DB) *PlatformSettingsStorage {
	return &PlatformSettingsStorage{db: db}
}

func (s *PlatformSettingsStorage) GetPlatformSettings(ctx context.Context, tenantID uint64) (*model.PlatformSettings, error) {
	const query = `
		SELECT id, tenantid, title, headline, tagline, websiteurl, maxworkspacesperuser, maxsessionsperuser, maxappinstancesperuser, createdat, updatedat
		FROM platform_settings
		WHERE tenantid = $1
	`

	var settings model.PlatformSettings
	if err := s.db.GetContext(ctx, &settings, query, tenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.PlatformSettings{TenantID: tenantID}, nil
		}
		return nil, fmt.Errorf("unable to get platform settings for tenant %d: %w", tenantID, err)
	}

	return &settings, nil
}

func (s *PlatformSettingsStorage) UpsertPlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	const query = `
		INSERT INTO public.platform_settings (tenantid, title, headline, tagline, websiteurl, maxworkspacesperuser, maxsessionsperuser, maxappinstancesperuser, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (tenantid) DO UPDATE SET
		    title                  = EXCLUDED.title,
		    headline               = EXCLUDED.headline,
		    tagline                = EXCLUDED.tagline,
		    websiteurl             = EXCLUDED.websiteurl,
		    maxworkspacesperuser   = EXCLUDED.maxworkspacesperuser,
		    maxsessionsperuser     = EXCLUDED.maxsessionsperuser,
		    maxappinstancesperuser = EXCLUDED.maxappinstancesperuser,
		    updatedat              = EXCLUDED.updatedat
		RETURNING id, tenantid, title, headline, tagline, websiteurl, maxworkspacesperuser, maxsessionsperuser, maxappinstancesperuser, createdat, updatedat
	`

	var result model.PlatformSettings
	if err := s.db.GetContext(ctx, &result, query,
		settings.TenantID,
		settings.Title,
		settings.Headline,
		settings.Tagline,
		settings.WebsiteURL,
		settings.MaxWorkspacesPerUser,
		settings.MaxSessionsPerUser,
		settings.MaxAppInstancesPerUser,
	); err != nil {
		return nil, fmt.Errorf("unable to upsert platform settings for tenant %d: %w", settings.TenantID, err)
	}

	return &result, nil
}
