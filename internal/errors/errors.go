package errors

import (
	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChorusError struct {
	GRPCCode   codes.Code
	ChorusCode string
	Instance   string
	Title      string
	Message    string
	CausedBy   error
}

func (e *ChorusError) Error() string {
	return e.Message
}

func (e *ChorusError) ToGRPCStatus() *status.Status {
	st := status.New(e.GRPCCode, e.Message)

	// Create error details with Chorus-specific information
	errorDetail := &errorspb.ErrorDetail{
		ChorusCode: e.ChorusCode,
		Instance:   e.Instance,
		Title:      e.Title,
		Message:    e.Message,
		Timestamp:  timestamppb.Now(),
	}

	// Add details to the status
	statusWithDetails, err := st.WithDetails(errorDetail)
	if err != nil {
		// If adding details fails, return original status
		return st
	}

	return statusWithDetails
}

func NewInvalidCredentialsError() *ChorusError {
	return &ChorusError{
		GRPCCode:   codes.Unauthenticated,
		ChorusCode: "INVALID_CREDENTIALS",
		Title:      "Invalid Credentials",
		Message:    "The provided credentials are invalid.",
	}
}

func NewPermissionDeniedError(message string) *ChorusError {
	return &ChorusError{
		GRPCCode:   codes.PermissionDenied,
		ChorusCode: "PERMISSION_DENIED",
		Title:      "Permission Denied",
		Message:    message,
	}
}

func NewInternalError(message string, causedBy error) *ChorusError {
	return &ChorusError{
		GRPCCode:   codes.Internal,
		ChorusCode: "INTERNAL_ERROR",
		Title:      "Internal Server Error",
		Message:    message,
		CausedBy:   causedBy,
	}
}
