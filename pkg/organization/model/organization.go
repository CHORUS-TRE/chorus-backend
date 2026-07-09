package model

import "time"

// Organization maps an entry in the 'organizations' database table.
type Organization struct {
	ID uint64

	TenantID uint64

	Name        string
	Description *string

	Logo            []byte
	LogoContentType *string

	Country *string
	City    *string

	ContactUserID *uint64
	WebsiteURL    *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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
