package model

// User maps an entry in the 'users' database table.
// Nullable fields have pointer types.
type User struct {
	ID       uint64
	TenantID uint64

	FirstName   string
	LastName    string
	Username    string
	Source      string
	Password    string
	Status      string
	TotpEnabled bool

	TotpSecret *string

	Roles []string // Roles is not a column of the 'users' table.
}
