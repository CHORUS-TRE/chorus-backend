package model

import "time"

const TechnicalTenantID = uint64(9999999)

type Tenant struct {
	ID           uint64
	Name         string
	CreationDate time.Time `db:"createdat"`
	UpdateDate   time.Time `db:"updatedat"`
}
