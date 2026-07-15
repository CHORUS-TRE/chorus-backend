package model

import "time"

// Organization maps an entry in the 'organizations' database table.
type Organization struct {
	ID uint64

	TenantID uint64

	Name        string  `validate:"required,generalstring,max=255"`
	Description *string `validate:"omitempty,max=250,generalstring"`

	// Logo is nil when no logo is being written - there is no way to clear an
	// existing logo via update, only add one or replace it with another.
	Logo *OrganizationLogo

	Country *string `validate:"omitempty,iso3166_1_alpha2"`
	City    *string `validate:"omitempty,max=100,generalstring"`

	ContactUserID *uint64
	WebsiteURL    *string `validate:"omitempty,max=2048,url"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// OrganizationLogo bundles the logo bytes with their content type so the two can
// never be provided independently - the pair is validated together whenever this
// struct is present (go-playground/validator dives into non-nil struct pointers
// automatically), and skipped entirely when nil. max=524288 is 512 KB.
type OrganizationLogo struct {
	Logo            []byte `validate:"required,max=524288"`
	LogoContentType string `validate:"required,oneof=image/png image/jpeg image/webp"`
}

// Unwrap returns the logo bytes and content type, or the zero values if l is nil.
func (l *OrganizationLogo) Unwrap() ([]byte, string) {
	if l == nil {
		return nil, ""
	}
	return l.Logo, l.LogoContentType
}

func (Organization) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":        true,
		"name":      true,
		"country":   true,
		"city":      true,
		"createdat": true,
	}

	return validSortTypes[sortType]
}
