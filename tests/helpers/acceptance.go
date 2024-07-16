package helpers

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/openapi"

	"github.com/go-openapi/runtime"
	. "github.com/onsi/gomega"
)

func ExpectAPIError(expectedErr interface{}) Assertion {
	if apiErr, ok := expectedErr.(*runtime.APIError); ok {
		serviceError := openapi.ExtractServiceError(apiErr)
		return Expect(serviceError.Error())
	}

	if err, ok := expectedErr.(error); ok {
		return Expect(err.Error())
	}

	return Expect(expectedErr)
}
