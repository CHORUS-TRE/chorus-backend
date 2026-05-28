package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
)

func PlatformSettingsFromBusiness(s *model.PlatformSettings) (*chorus.PlatformSettings, error) {
	ca, err := ToProtoTimestamp(s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	return &chorus.PlatformSettings{
		Id:                     s.ID,
		TenantId:               s.TenantID,
		Title:                  s.Title,
		Headline:               s.Headline,
		Tagline:                s.Tagline,
		WebsiteURL:             s.WebsiteURL,
		TouVersionId:           s.TouVersionID,
		MaxWorkspacesPerUser:   s.MaxWorkspacesPerUser,
		MaxSessionsPerUser:     s.MaxSessionsPerUser,
		MaxAppInstancesPerUser: s.MaxAppInstancesPerUser,
		CreatedAt:              ca,
		UpdatedAt:              ua,
	}, nil
}

func PlatformSettingsToBusiness(p *chorus.PlatformSettings) *model.PlatformSettings {
	return &model.PlatformSettings{
		Title:                  p.Title,
		Headline:               p.Headline,
		Tagline:                p.Tagline,
		WebsiteURL:             p.WebsiteURL,
		TouVersionID:           p.TouVersionId,
		MaxWorkspacesPerUser:   p.MaxWorkspacesPerUser,
		MaxSessionsPerUser:     p.MaxSessionsPerUser,
		MaxAppInstancesPerUser: p.MaxAppInstancesPerUser,
	}
}
