//go:build acceptance

package authentication_test

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestAuthenticationService(t *testing.T) {
	RegisterFailHandler(Fail)
	config.DefaultReporterConfig.NoColor = true // With colors, the tests reports don't print nicely in Jenkins logs.
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Authenticate Service Suite", []Reporter{junitReporter})
}

var _ = AfterSuite(func() {
	cleanTables()
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