//go:build unit

package errors

import (
	"errors"
	"fmt"
	"testing"

	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestChorusError_Error(t *testing.T) {
	err := ErrNotFound.WithMessage("user 42 not found")
	assert.Equal(t, "user 42 not found", err.Error())
}

func TestChorusError_ErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("sql: no rows")
	err := ErrNotFound.Wrap(cause, "user 42 not found")
	assert.Equal(t, "user 42 not found: sql: no rows", err.Error())
}

func TestChorusError_ErrorEmpty(t *testing.T) {
	assert.Equal(t, "", ErrNotFound.Error())
}

func TestWithMessage_ReturnsNewInstance(t *testing.T) {
	original := ErrNotFound
	derived := original.WithMessage("gone")

	assert.NotSame(t, original, derived)
	assert.Equal(t, "", original.Message)
	assert.Equal(t, "gone", derived.Message)

	assert.Equal(t, original.GRPCCode, derived.GRPCCode)
	assert.Equal(t, original.ChorusCode, derived.ChorusCode)
	assert.Equal(t, original.Title, derived.Title)
}

func TestWithCause_ReturnsNewInstance(t *testing.T) {
	cause := fmt.Errorf("db connection lost")
	derived := ErrInternal.WithCause(cause)

	assert.NotSame(t, ErrInternal, derived)
	assert.Nil(t, ErrInternal.CausedBy)
	assert.Equal(t, cause, derived.CausedBy)
}

func TestWrap_SetsMessageAndCause(t *testing.T) {
	cause := fmt.Errorf("timeout")
	derived := ErrInternal.Wrap(cause, "failed to fetch user")

	assert.Equal(t, "failed to fetch user", derived.Message)
	assert.Equal(t, cause, derived.CausedBy)
	assert.Equal(t, codes.Internal, derived.GRPCCode)
	assert.Equal(t, errorspb.ChorusErrorCode_INTERNAL_ERROR, derived.ChorusCode)
}

func TestUnwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := ErrInternal.WithCause(cause)

	assert.Equal(t, cause, errors.Unwrap(err))
}

func TestUnwrap_NilCause(t *testing.T) {
	assert.Nil(t, errors.Unwrap(ErrNotFound))
}

func TestErrorsAs(t *testing.T) {
	cause := fmt.Errorf("db error")
	wrapped := fmt.Errorf("service layer: %w", ErrInternal.Wrap(cause, "store failed"))

	var cErr *ChorusError
	require.True(t, errors.As(wrapped, &cErr))
	assert.Equal(t, codes.Internal, cErr.GRPCCode)
	assert.Equal(t, errorspb.ChorusErrorCode_INTERNAL_ERROR, cErr.ChorusCode)
	assert.Equal(t, "store failed", cErr.Message)
	assert.Equal(t, cause, cErr.CausedBy)
}

func TestErrorsIs_CauseChain(t *testing.T) {
	sentinel := fmt.Errorf("sentinel")
	err := ErrInternal.WithCause(sentinel)

	assert.True(t, errors.Is(err, sentinel))
}

func TestToGRPCStatus(t *testing.T) {
	err := ErrNotFound.WithMessage("workspace 7 not found")
	st := err.ToGRPCStatus()

	assert.Equal(t, codes.NotFound, st.Code())
	assert.Equal(t, "workspace 7 not found", st.Message())

	require.Len(t, st.Details(), 1)
	detail, ok := st.Details()[0].(*errorspb.ErrorDetail)
	require.True(t, ok)
	assert.Equal(t, errorspb.ChorusErrorCode_NOT_FOUND, detail.ChorusCode)
	assert.Equal(t, "Not Found", detail.Title)
	assert.Equal(t, "workspace 7 not found", detail.Message)
}

func TestToGRPCStatus_CausedByNotExposed(t *testing.T) {
	cause := fmt.Errorf("secret db error")
	err := ErrInternal.Wrap(cause, "something went wrong")
	st := err.ToGRPCStatus()

	detail, ok := st.Details()[0].(*errorspb.ErrorDetail)
	require.True(t, ok)

	assert.Equal(t, "something went wrong", detail.Message)
	assert.NotContains(t, detail.Message, "secret db error")
	assert.NotContains(t, st.Message(), "secret db error")
}

func TestToGRPCStatus_RoundTrip(t *testing.T) {
	err := ErrPermissionDenied.WithMessage("not your resource")
	st := err.ToGRPCStatus()

	// Simulate what the interceptor does: convert to gRPC error and back
	grpcErr := st.Err()
	recovered, ok := status.FromError(grpcErr)
	require.True(t, ok)

	assert.Equal(t, codes.PermissionDenied, recovered.Code())
	require.Len(t, recovered.Details(), 1)

	detail, ok := recovered.Details()[0].(*errorspb.ErrorDetail)
	require.True(t, ok)
	assert.Equal(t, errorspb.ChorusErrorCode_PERMISSION_DENIED, detail.ChorusCode)
}

func TestWithValidationErrors(t *testing.T) {
	fields := []ValidationField{
		{Field: "FirstName", Reason: "required"},
		{Field: "Email", Reason: "email"},
	}
	err := ErrValidation.WithMessage("Validation error").WithValidationErrors(fields)

	assert.Equal(t, "Validation error", err.Message)
	assert.Equal(t, errorspb.ChorusErrorCode_VALIDATION_ERROR, err.ChorusCode)
	require.Len(t, err.ValidationErrors, 2)
	assert.Equal(t, "FirstName", err.ValidationErrors[0].Field)
	assert.Equal(t, "required", err.ValidationErrors[0].Reason)
	assert.Equal(t, "Email", err.ValidationErrors[1].Field)
	assert.Equal(t, "email", err.ValidationErrors[1].Reason)
}

func TestWithValidationErrors_PreservesCause(t *testing.T) {
	cause := fmt.Errorf("validator error")
	err := ErrValidation.Wrap(cause, "Validation error").WithValidationErrors([]ValidationField{
		{Field: "Name", Reason: "required"},
	})

	assert.Equal(t, cause, err.CausedBy)
	require.Len(t, err.ValidationErrors, 1)
}

func TestToGRPCStatus_WithValidationErrors(t *testing.T) {
	fields := []ValidationField{
		{Field: "FirstName", Reason: "required"},
		{Field: "LastName", Reason: "required"},
	}
	err := ErrValidation.WithMessage("Validation error").WithValidationErrors(fields)
	st := err.ToGRPCStatus()

	assert.Equal(t, codes.InvalidArgument, st.Code())

	require.Len(t, st.Details(), 1)
	detail, ok := st.Details()[0].(*errorspb.ErrorDetail)
	require.True(t, ok)

	assert.Equal(t, errorspb.ChorusErrorCode_VALIDATION_ERROR, detail.ChorusCode)
	assert.Equal(t, "Validation error", detail.Message)
	require.Len(t, detail.ValidationErrors, 2)
	assert.Equal(t, "FirstName", detail.ValidationErrors[0].Field)
	assert.Equal(t, "required", detail.ValidationErrors[0].Reason)
	assert.Equal(t, "LastName", detail.ValidationErrors[1].Field)
	assert.Equal(t, "required", detail.ValidationErrors[1].Reason)
}

func TestToGRPCStatus_WithoutValidationErrors(t *testing.T) {
	err := ErrNotFound.WithMessage("not found")
	st := err.ToGRPCStatus()

	detail, ok := st.Details()[0].(*errorspb.ErrorDetail)
	require.True(t, ok)
	assert.Empty(t, detail.ValidationErrors)
}

func TestNestedErrorChain(t *testing.T) {
	// Simulate: store returns sql error → user service wraps it → workspace service wraps that
	sqlErr := fmt.Errorf("sql: no rows in result set")
	userServiceErr := ErrNotFound.Wrap(sqlErr, "Unable to get user 42")
	workspaceServiceErr := ErrInternal.Wrap(userServiceErr, "Unable to create workspace")

	// The full chain is visible via Error()
	t.Logf("Error():    %s", workspaceServiceErr.Error())
	t.Logf("CausedBy:   %s", workspaceServiceErr.CausedBy)
	assert.Equal(t, "Unable to create workspace: Unable to get user 42: sql: no rows in result set", workspaceServiceErr.Error())

	// errors.As finds the outermost ChorusError
	var cErr *ChorusError
	require.True(t, errors.As(workspaceServiceErr, &cErr))
	assert.Equal(t, "Unable to create workspace", cErr.Message)

	// errors.Is can still find the root cause
	assert.True(t, errors.Is(workspaceServiceErr, sqlErr))

	// Unwrap gives the next error in the chain
	inner := errors.Unwrap(workspaceServiceErr)
	innerChorus, ok := inner.(*ChorusError)
	require.True(t, ok)
	assert.Equal(t, "Unable to get user 42", innerChorus.Message)
	assert.Equal(t, sqlErr, innerChorus.CausedBy)
}

func TestCatalogCoversAllCodes(t *testing.T) {
	catalog := map[errorspb.ChorusErrorCode]*ChorusError{
		errorspb.ChorusErrorCode_INVALID_REQUEST:      ErrInvalidRequest,
		errorspb.ChorusErrorCode_VALIDATION_ERROR:     ErrValidation,
		errorspb.ChorusErrorCode_CONVERSION_ERROR:     ErrConversion,
		errorspb.ChorusErrorCode_NOT_FOUND:            ErrNotFound,
		errorspb.ChorusErrorCode_ALREADY_EXISTS:       ErrAlreadyExists,
		errorspb.ChorusErrorCode_UNAUTHENTICATED:      ErrUnauthenticated,
		errorspb.ChorusErrorCode_INVALID_CREDENTIALS:  ErrInvalidCredentials,
		errorspb.ChorusErrorCode_TWO_FACTOR_REQUIRED:  ErrTwoFactorRequired,
		errorspb.ChorusErrorCode_PERMISSION_DENIED:    ErrPermissionDenied,
		errorspb.ChorusErrorCode_INTERNAL_ERROR:       ErrInternal,
	}

	for code, err := range catalog {
		assert.Equal(t, code, err.ChorusCode, "catalog entry %s has wrong ChorusCode", code)
		assert.NotEmpty(t, err.Title, "catalog entry %s has empty Title", code)
		assert.NotEqual(t, codes.OK, err.GRPCCode, "catalog entry %s has OK gRPC code", code)
	}
}
