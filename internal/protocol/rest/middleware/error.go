package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// CustomHTTPError handles gRPC errors and returns custom HTTP error responses
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"error": "Internal Server Error", "code": 500, "message": "An unexpected error occurred", "details": "chorus-backend-error"}`

	w.Header().Set("Content-Type", "application/json")

	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	httpStatus := runtime.HTTPStatusFromCode(s.Code())

	customErr := CustomErrorResponse{
		Error:   http.StatusText(httpStatus),
		Code:    httpStatus,
		Message: s.Message(),
		Details: "chorus-backend-error",
	}

	buf, merr := json.Marshal(customErr)
	if merr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fallback))
		return
	}

	w.WriteHeader(httpStatus)
	_, _ = w.Write(buf)
}
