package errors

import (
	"fmt"

	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChorusError is a custom error type that includes additional
// context and can be converted to a gRPC status error.
type ChorusError struct {
	GRPCCode         codes.Code
	ChorusCode       errorspb.ChorusErrorCode
	Title            string
	Message          string
	CausedBy         error
	ValidationErrors []*errorspb.ValidationError
}

// ValidationField represents a single field validation failure.
type ValidationField struct {
	Field  string
	Reason string
}

func (e *ChorusError) Error() string {
	if e.CausedBy != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.CausedBy)
	}
	return e.Message
}

func (e *ChorusError) ToGRPCStatus() *status.Status {
	st := status.New(e.GRPCCode, e.Message)

	// Create error details with Chorus-specific information
	errorDetail := &errorspb.ErrorDetail{
		ChorusCode:       e.ChorusCode,
		Title:            e.Title,
		Message:          e.Message,
		ValidationErrors: e.ValidationErrors,
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

func (e *ChorusError) WithValidationErrors(fields []ValidationField) *ChorusError {
	ve := make([]*errorspb.ValidationError, len(fields))
	for i, f := range fields {
		ve[i] = &errorspb.ValidationError{
			Field:  f.Field,
			Reason: f.Reason,
		}
	}
	return &ChorusError{
		GRPCCode:         e.GRPCCode,
		ChorusCode:       e.ChorusCode,
		Title:            e.Title,
		Message:          e.Message,
		CausedBy:         e.CausedBy,
		ValidationErrors: ve,
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
