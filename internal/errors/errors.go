package errors

import (
	"fmt"
	"runtime"
	"strings"

	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const modulePrefix = "github.com/CHORUS-TRE/chorus-backend/"

// ChorusError is a custom error type that includes additional
// context and can be converted to a gRPC status error.
type ChorusError struct {
	GRPCCode         codes.Code
	ChorusCode       errorspb.ChorusErrorCode
	Title            string
	Message          string
	CausedBy         error
	ValidationErrors []*errorspb.ValidationError
	Stack            []uintptr
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

	statusWithDetails, err := st.WithDetails(&errorspb.ErrorDetail{
		ChorusCode:       e.ChorusCode,
		Title:            e.Title,
		Message:          e.Message,
		ValidationErrors: e.ValidationErrors,
	})
	if err != nil {
		return st
	}

	return statusWithDetails
}

func (e *ChorusError) clone() *ChorusError {
	return &ChorusError{GRPCCode: e.GRPCCode, ChorusCode: e.ChorusCode, Title: e.Title, Message: e.Message, CausedBy: e.CausedBy, ValidationErrors: e.ValidationErrors, Stack: e.Stack}
}

func (e *ChorusError) Wrap(err error, message string) *ChorusError {
	c := e.clone()
	c.Message = message
	c.CausedBy = err
	c.Stack = e.stack()
	if wrappedErr, ok := err.(*ChorusError); ok {
		// TODO as the stack share common frames, keep only the differing branches
		c.Stack = append(c.Stack, wrappedErr.Stack...)
	}
	return c
}

func (e *ChorusError) Unwrap() error {
	return e.CausedBy
}

// StackTrace formats the captured stack as a human-readable string,
// showing only frames from this module with shortened paths.
func (e *ChorusError) StackTrace() string {
	if len(e.Stack) == 0 {
		return ""
	}
	var b strings.Builder
	frames := runtime.CallersFrames(e.Stack)
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.Function, modulePrefix) {
			fmt.Fprintf(&b, "%s\n\t%s:%d\n",
				strings.TrimPrefix(frame.Function, modulePrefix),
				strings.TrimPrefix(frame.File, modulePrefix),
				frame.Line,
			)
		}
		if !more {
			break
		}
	}
	return b.String()
}

// stack returns the existing stack if present, or captures a new one.
func (e *ChorusError) stack() []uintptr {
	if e.Stack != nil {
		return e.Stack
	}
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs) // skip: runtime.Callers, callerStack, calling method
	return pcs[:n]
}

func (e *ChorusError) WithMessage(message string) *ChorusError {
	c := e.clone()
	c.Message = message
	c.Stack = e.stack()
	return c
}

func (e *ChorusError) WithCause(causedBy error) *ChorusError {
	c := e.clone()
	c.CausedBy = causedBy
	c.Stack = e.stack()
	return c
}

func (e *ChorusError) WithValidationErrors(fields []ValidationField) *ChorusError {
	ve := make([]*errorspb.ValidationError, len(fields))
	for i, f := range fields {
		ve[i] = &errorspb.ValidationError{
			Field:  f.Field,
			Reason: f.Reason,
		}
	}
	c := e.clone()
	c.ValidationErrors = ve
	c.Stack = e.stack()
	return c
}
