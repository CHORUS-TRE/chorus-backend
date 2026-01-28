//go:build acceptance

package audit_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
)

func TestAuditService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Audit Service Suite", []Reporter{junitReporter})
}

var _ = AfterSuite(func() {
	// cleanTables() - defined in audit_test.go
})

func Then(text string, body func()) bool {
	return It("then "+text, body)
}

func ExpectAPIErr(err interface{}) Assertion {
	return helpers.ExpectAPIError(err)
}

func Given(text string, body func()) bool {
	return Context("given "+text, body)
}
