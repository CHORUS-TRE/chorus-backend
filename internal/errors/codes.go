package errors

import (
	"errors"

	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc/codes"
)

// Error catalog — each entry maps to a ChorusErrorCode enum value defined in api/proto/v1/errors.proto.
// Use these as templates via .WithMessage(), .Wrap(), or .WithCause() to create specific error instances.
var (
	ErrNoRowsUpdated = errors.New("database: no rows updated")
	ErrNoRowsDeleted = errors.New("database: no rows deleted")

	// Client errors
	ErrInvalidRequest = &ChorusError{GRPCCode: codes.InvalidArgument, ChorusCode: errorspb.ChorusErrorCode_INVALID_REQUEST, Title: "Invalid Request"}
	ErrValidation     = &ChorusError{GRPCCode: codes.InvalidArgument, ChorusCode: errorspb.ChorusErrorCode_VALIDATION_ERROR, Title: "Validation Error"}
	ErrConversion     = &ChorusError{GRPCCode: codes.Internal, ChorusCode: errorspb.ChorusErrorCode_CONVERSION_ERROR, Title: "Conversion Error"}

	// Resource errors
	ErrNotFound      = &ChorusError{GRPCCode: codes.NotFound, ChorusCode: errorspb.ChorusErrorCode_NOT_FOUND, Title: "Not Found"}
	ErrAlreadyExists = &ChorusError{GRPCCode: codes.AlreadyExists, ChorusCode: errorspb.ChorusErrorCode_ALREADY_EXISTS, Title: "Already Exists"}

	// Authentication & authorization errors
	ErrUnauthenticated    = &ChorusError{GRPCCode: codes.Unauthenticated, ChorusCode: errorspb.ChorusErrorCode_UNAUTHENTICATED, Title: "Unauthenticated"}
	ErrInvalidCredentials = &ChorusError{GRPCCode: codes.Unauthenticated, ChorusCode: errorspb.ChorusErrorCode_INVALID_CREDENTIALS, Title: "Invalid Credentials", Message: "The provided credentials are invalid."}
	ErrTwoFactorRequired  = &ChorusError{GRPCCode: codes.FailedPrecondition, ChorusCode: errorspb.ChorusErrorCode_TWO_FACTOR_REQUIRED, Title: "Two-Factor Authentication Required"}
	ErrPermissionDenied   = &ChorusError{GRPCCode: codes.PermissionDenied, ChorusCode: errorspb.ChorusErrorCode_PERMISSION_DENIED, Title: "Permission Denied"}

	// Server errors
	ErrInternal = &ChorusError{GRPCCode: codes.Internal, ChorusCode: errorspb.ChorusErrorCode_INTERNAL_ERROR, Title: "Internal Server Error"}
)
