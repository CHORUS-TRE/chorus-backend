//go:build acceptance

package workspace_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo/v2"
)

func TestWorkspaceService(t *testing.T) {
	helpers.RunSuite(t, "Workspace Service Suite")
}

var _ = AfterSuite(func() {
	cleanTables()
})

var (
	Given        = helpers.Given
	Then         = helpers.Then
	ExpectAPIErr = helpers.ExpectAPIError
)
