package middleware

import (
	"context"
	"strings"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

type validation struct {
	next service.TermsOfUseer
}

func Validation() func(service.TermsOfUseer) service.TermsOfUseer {
	return func(next service.TermsOfUseer) service.TermsOfUseer {
		return &validation{next: next}
	}
}

func (v *validation) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	if strings.TrimSpace(content) == "" {
		return nil, cerr.ErrValidation.WithMessage("Content is required")
	}
	return v.next.CreateTermsOfUseVersion(ctx, tenantID, content)
}

func (v *validation) UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error) {
	if versionID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Version ID is required")
	}
	if strings.TrimSpace(content) == "" {
		return nil, cerr.ErrValidation.WithMessage("Content is required")
	}
	return v.next.UpdateTermsOfUseVersion(ctx, tenantID, versionID, content)
}

func (v *validation) PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	if versionID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Version ID is required")
	}
	return v.next.PublishTermsOfUseVersion(ctx, tenantID, versionID)
}

func (v *validation) GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	if versionID == 0 {
		return nil, cerr.ErrValidation.WithMessage("Version ID is required")
	}
	return v.next.GetTermsOfUseVersion(ctx, tenantID, versionID)
}

func (v *validation) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	return v.next.ListTermsOfUseVersions(ctx, tenantID, pagination)
}

func (v *validation) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error) {
	return v.next.GetCurrentTermsOfUseVersion(ctx, tenantID)
}

func (v *validation) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	return v.next.ListTermsOfUseAcceptances(ctx, tenantID, userID)
}

func (v *validation) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error) {
	return v.next.GetMyTermsOfUseStatus(ctx, tenantID, userID)
}

func (v *validation) AcceptTermsOfUse(ctx context.Context, tenantID, userID uint64) (*model.TermsOfUseAcceptance, error) {
	return v.next.AcceptTermsOfUse(ctx, tenantID, userID)
}
