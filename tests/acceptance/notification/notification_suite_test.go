//go:build acceptance

package notification_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo/v2"
)

func TestNotificationService(t *testing.T) {
	helpers.RunSuite(t, "Notification Service Suite")
}

var _ = AfterSuite(func() {
	cleanTables()
})

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
