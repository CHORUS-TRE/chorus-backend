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
	Instance string `json:"instance"`
	Code     string `json:"code"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

// CustomHTTPError handles gRPC errors and returns custom HTTP error responses
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"instance": "/error", "error": "Internal Server Error", "status": 500, "message": "An unexpected error occurred"}`

	w.Header().Set("Content-Type", "application/json")

	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	httpStatus := runtime.HTTPStatusFromCode(s.Code())

	customErr := CustomErrorResponse{
		Instance: r.URL.Path,
		Code:     "CHORUS_ERROR_CODE", // To be defined in GRPC error types
		Error:    http.StatusText(httpStatus),
		Message:  s.Message(),
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
