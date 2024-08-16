package openapi

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
)

type openAPIError struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}
type openAPINumericError struct {
	Code    uint64 `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Extract the 'real' error from the service called using swagger openapi client
func ExtractServiceError(originalError error) error {
	switch originalError.(type) {
	case *runtime.APIError:
		apiErr := originalError.(*runtime.APIError)
		resp := apiErr.Response.(runtime.ClientResponse)
		bodyBytes, err := io.ReadAll(resp.Body())
		if err != nil {
			return fmt.Errorf("%s: %w", err.Error(), originalError)
		}
		var realError openAPIError
		err = json.Unmarshal(bodyBytes, &realError)
		if err != nil {
			var realError openAPINumericError
			err = json.Unmarshal(bodyBytes, &realError)
			if err != nil {
				return fmt.Errorf("%s: %w", string(bodyBytes), originalError)
			}
			return fmt.Errorf("%d: %s, %s: %w", realError.Code, realError.Error, realError.Message, originalError)
		}
		return fmt.Errorf("%s: %s, %s: %w", realError.Code, realError.Error, realError.Message, originalError)
	}
	return originalError
}
