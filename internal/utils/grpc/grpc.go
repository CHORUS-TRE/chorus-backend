package grpc

import (
	"database/sql"

	"github.com/CHORUS-TRE/chorus-backend/pkg/common/service"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"

	auth_service "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	val "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
)

func ErrorCode(err error) codes.Code {
	switch errors.Cause(err) {
	case sql.ErrNoRows, database.ErrNoRowsUpdated, database.ErrNoRowsDeleted:
		return codes.NotFound
	}

	switch errors.Cause(err).(type) {
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
