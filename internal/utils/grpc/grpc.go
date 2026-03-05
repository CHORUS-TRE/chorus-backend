package grpc

import (
	"database/sql"
	"errors"

	val "github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/service"
)

func ErrorCode(err error) codes.Code {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, cerr.ErrNoRowsUpdated) || errors.Is(err, cerr.ErrNoRowsDeleted) {
		return codes.NotFound
	}

	// Find the root cause.
	cause := err
	for {
		next := errors.Unwrap(cause)
		if next == nil {
			break
		}
		cause = next
	}

	switch cause.(type) {
	case *val.InvalidValidationError, val.ValidationErrors, *service.InvalidParametersErr:
		return codes.InvalidArgument
	case *service.ResourceAlreadyExistsErr:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
