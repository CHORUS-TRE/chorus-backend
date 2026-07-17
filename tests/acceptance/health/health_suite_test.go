//go:build acceptance

package health_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
)

func TestHealthService(t *testing.T) {
	helpers.RunSuite(t, "Health Service Suite")
}

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
