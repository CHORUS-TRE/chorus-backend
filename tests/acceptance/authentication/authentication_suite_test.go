//go:build acceptance

package authentication_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo/v2"
)

func TestAuthenticationService(t *testing.T) {
	helpers.RunSuite(t, "Authenticate Service Suite")
}

var _ = AfterSuite(func() {
	cleanTables()
})

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
