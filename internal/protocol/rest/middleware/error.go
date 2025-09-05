package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	errorspb "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

type CustomErrorResponse struct {
	Code    int           `json:"code"`    // gRPC code
	Message string        `json:"message"` // error message
	Details []interface{} `json:"details"` // details as array
}

// CustomHTTPError handles gRPC errors and returns custom HTTP error responses
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"code": 13, "message": "An unexpected error occurred", "error": "An unexpected error occurred", "details": []}`

	w.Header().Set("Content-Type", "application/json")

	// Convert gRPC error to HTTP status
	st, _ := status.FromError(err)

	// Get http status
	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	// Create custom error response
	var customErr CustomErrorResponse
	customErr.Code = int(st.Code())
	customErr.Message = st.Message()
	customErr.Details = []interface{}{}

	// Check if we have custom error detail in status details
	for _, detail := range st.Details() {
		if errorDetail, ok := detail.(*errorspb.ErrorDetail); ok {
			// Add current request context
			errorDetail.Instance = r.URL.Path

			// Create structured error detail as an object in the details array
			detailObj := map[string]interface{}{
				"chorusCode": errorDetail.ChorusCode,
				"instance":   errorDetail.Instance,
				"title":      errorDetail.Title,
				"message":    errorDetail.Message,
				"timestamp":  errorDetail.Timestamp.AsTime().String(),
			}
			customErr.Details = append(customErr.Details, detailObj)
		}
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
