package model

// TotpRecoveryCode maps an entry in the `totp_recovery_codes` table.
type TotpRecoveryCode struct {
	ID       uint64
	TenantID uint64
	UserID   uint64
	Code     string
}
