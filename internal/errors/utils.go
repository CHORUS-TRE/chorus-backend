package errors

import (
	"database/sql"
	"errors"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	val "github.com/go-playground/validator/v10"
)

// WrapStoreError translates store errors to the appropriate ChorusError.
// sql.ErrNoRows, database.ErrNoRowsUpdated and database.ErrNoRowsDeleted → ErrNotFound, everything else → ErrInternal.
func WrapStoreError(err error, message string) *ChorusError {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, database.ErrNoRowsUpdated) || errors.Is(err, database.ErrNoRowsDeleted) {
		return ErrNotFound.Wrap(err, message)
	}
	return ErrInternal.Wrap(err, message)
}

// WrapValidationError formats a validator error into a clean ChorusError
// with structured field-level validation details.
func WrapValidationError(err error) *ChorusError {
	var ve val.ValidationErrors
	if errors.As(err, &ve) {
		fields := make([]ValidationField, len(ve))
		for i, fe := range ve {
			fields[i] = ValidationField{
				Field: fe.Field(),
				Rule:  fe.Tag(),
			}
		}
		return ErrValidation.Wrap(err, "Validation error").WithValidationErrors(fields)
	}
	return ErrValidation.Wrap(err, err.Error())
}
