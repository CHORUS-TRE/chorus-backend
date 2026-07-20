//go:build acceptance

package health_test

import (
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	health "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/health/client/health_service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("health service", func() {

	Describe("get health check", func() {

		Given("any user", func() {

			When("the route GET '/api/rest/v1/health is called", func() {

				Then("a reply with success status should be returned", func() {
					req := health.NewHealthServiceGetHealthCheckParams()

					c := helpers.HealthServiceHTTPClient()
					_, err := c.HealthService.HealthServiceGetHealthCheck(req)

					ExpectAPIErr(err).Should(BeNil())
				})
			})
		})
	})
})
