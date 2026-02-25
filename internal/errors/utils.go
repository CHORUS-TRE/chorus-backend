package errors

import (
	"database/sql"
	"errors"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
)

// WrapStoreError translates store errors to the appropriate ChorusError.
// sql.ErrNoRows, database.ErrNoRowsUpdated and database.ErrNoRowsDeleted → ErrNotFound, everything else → ErrInternal.
func WrapStoreError(err error, message string) *ChorusError {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, database.ErrNoRowsUpdated) || errors.Is(err, database.ErrNoRowsDeleted) {
		return ErrNotFound.Wrap(err, message)
	}
	return ErrInternal.Wrap(err, message)
}
