package errors

import (
	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChorusError struct {
	GRPCCode   codes.Code
	ChorusCode errorspb.ChorusErrorCode
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
		Title:      e.Title,
		Message:    e.Message,
	}

	// Add details to the status
	statusWithDetails, err := st.WithDetails(errorDetail)
	if err != nil {
		// If adding details fails, return original status
		return st
	}

	return statusWithDetails
}

func (e *ChorusError) WithMessage(message string) *ChorusError {
	return &ChorusError{
		GRPCCode:   e.GRPCCode,
		ChorusCode: e.ChorusCode,
		Title:      e.Title,
		Message:    message,
		CausedBy:   e.CausedBy,
	}
}

func (e *ChorusError) WithCause(causedBy error) *ChorusError {
	return &ChorusError{
		GRPCCode:   e.GRPCCode,
		ChorusCode: e.ChorusCode,
		Title:      e.Title,
		Message:    e.Message,
		CausedBy:   causedBy,
	}
}

func (e *ChorusError) Wrap(err error, message string) *ChorusError {
	return &ChorusError{
		GRPCCode:   e.GRPCCode,
		ChorusCode: e.ChorusCode,
		Title:      e.Title,
		Message:    message,
		CausedBy:   err,
	}
}

func (e *ChorusError) Unwrap() error {
	return e.CausedBy
}
