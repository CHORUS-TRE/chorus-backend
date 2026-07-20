//go:build unit || integration || acceptance

package helpers

import (
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/openapi"
)

// RunSuite bootstraps the test configuration and runs the Ginkgo specs of the
// calling package.
func RunSuite(t *testing.T, description string) {
	Setup()
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, description)
}

// Given and Then complement the ginkgo When/Describe containers with BDD
// phrasing. Offset(1) makes failure locations point at the caller.
func Given(text string, body func()) bool {
	return ginkgo.Context("given "+text, ginkgo.Offset(1), body)
}

func Then(text string, body func()) bool {
	return ginkgo.It("then "+text, ginkgo.Offset(1), body)
}

// ExpectAPIError is a helper function to assert API errors in tests.
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
