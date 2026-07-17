//go:build acceptance

package app_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
)

func TestAppService(t *testing.T) {
	helpers.RunSuite(t, "App Service Suite")
}
