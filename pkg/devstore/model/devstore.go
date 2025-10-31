package model

import "time"

type DevStoreScope string

const (
	DevStoreScopeGlobal    DevStoreScope = "global"
	DevStoreScopeUser      DevStoreScope = "user"
	DevStoreScopeWorkspace DevStoreScope = "workspace"
)

type DevstoreEntry struct {
	ID uint64

	TenantID uint64

	Scope   DevStoreScope
	ScopeID uint64

	Key   string
	Value string

	CreatedAt time.Time
	UpdatedAt time.Time
}
