//go:build acceptance

package info_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
)

func TestInfo(t *testing.T) {
	helpers.RunSuite(t, "Info Suite")
}

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
