//go:build acceptance

package user_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo/v2"
)

func TestUserService(t *testing.T) {
	helpers.RunSuite(t, "User Service Suite")
}

var _ = AfterSuite(func() {
	cleanTables()
})

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
