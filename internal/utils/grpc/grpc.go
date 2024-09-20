package grpc

import (
	"database/sql"
	"errors"

	val "github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	auth_service "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/service"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

func ErrorCode(err error) codes.Code {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, database.ErrNoRowsUpdated) || errors.Is(err, database.ErrNoRowsDeleted) {
		return codes.NotFound
	}

	// Find the root cause.
	var cause error
	for {
		if c := errors.Unwrap(err); c != nil {
			cause = c
		} else {
			break
		}
	}

	switch cause.(type) {
	case *val.InvalidValidationError, val.ValidationErrors, *service.InvalidParametersErr, *user_service.ErrWeakPassword:
		return codes.InvalidArgument
	case *service.ResourceAlreadyExistsErr:
		return codes.AlreadyExists
	case *auth_service.ErrUnauthorized, *user_service.ErrUnauthorized:
		return codes.Unauthenticated
	default:
		return codes.Internal
	}
}
