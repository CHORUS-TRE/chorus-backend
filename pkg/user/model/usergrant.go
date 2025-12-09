package model

import (
	"time"
)

// UserGrant maps an entry in the 'user_grants' database table.
// Nullable fields have pointer types.
type UserGrant struct {
	ID uint64

	TenantID uint64
	UserID   uint64

	ClientID     string
	Scope        string
	GrantedUntil *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
