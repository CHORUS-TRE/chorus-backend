package model

import "time"

type TermsOfUseVersionStatus string

const (
	TermsOfUseVersionStatusDraft     TermsOfUseVersionStatus = "Draft"
	TermsOfUseVersionStatusPublished TermsOfUseVersionStatus = "Published"
	TermsOfUseVersionStatusArchived  TermsOfUseVersionStatus = "Archived"
)

type TermsOfUseVersion struct {
	ID       uint64
	TenantID uint64

	Content string
	Status  TermsOfUseVersionStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TermsOfUseVersion) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":        true,
		"createdat": true,
		"status":    true,
		"updatedat": true,
	}

	return validSortTypes[sortType]
}

type TermsOfUseAcceptance struct {
	ID       uint64
	TenantID uint64

	UserID              uint64
	TermsOfUseVersionID uint64

	AcceptedAt time.Time
}
