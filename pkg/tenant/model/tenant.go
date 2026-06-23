package model

import "time"

type Tenant struct {
	ID           uint64
	Name         string
	CreationDate time.Time `db:"createdat"`
	UpdateDate   time.Time `db:"updatedat"`
}
