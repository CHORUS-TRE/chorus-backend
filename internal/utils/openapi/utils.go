package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/pkg/errors"
	"io"
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
			return errors.Wrap(originalError, err.Error())
		}
		var realError openAPIError
		err = json.Unmarshal(bodyBytes, &realError)
		if err != nil {
			var realError openAPINumericError
			err = json.Unmarshal(bodyBytes, &realError)
			if err != nil {
				return errors.Wrap(originalError, string(bodyBytes))
			}
			return errors.Wrap(originalError, fmt.Sprintf("%d: %s, %s", realError.Code, realError.Error, realError.Message))
		}
		return errors.Wrap(originalError, fmt.Sprintf("%s: %s, %s", realError.Code, realError.Error, realError.Message))
	}
	return originalError
}
