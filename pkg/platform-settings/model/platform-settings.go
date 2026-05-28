package model

import "time"

type PlatformSettings struct {
	ID uint64

	TenantID uint64

	Title      string
	Headline   string
	Tagline    string
	WebsiteURL string

	TouVersionID uint64

	MaxWorkspacesPerUser   uint32
	MaxSessionsPerUser     uint32
	MaxAppInstancesPerUser uint32

	CreatedAt time.Time
	UpdatedAt time.Time
}
