//go:build acceptance

package organization_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	organization "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/organization/client/organization_service"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/organization/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	orgTestTenantID = uint64(88700)
	orgTestJWTUser  = uint64(95500)
	orgFixtureID    = "991001"
)

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	auth := httptransport.BearerToken(t)
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = auth
	}
}

var _ = Describe("organization service", func() {
	helpers.Setup()

	Describe("list organizations", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route GET '/api/rest/v1/organizations' is called", func() {
				req := organization.NewOrganizationServiceListOrganizationsParams()

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceListOrganizations(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, "Public", map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

				When("the route GET '/api/rest/v1/organizations' is called", func() {
					req := organization.NewOrganizationServiceListOrganizationsParams()

					c := helpers.OrganizationServiceHTTPClient()
					_, err := c.OrganizationService.OrganizationServiceListOrganizations(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

			When("the route GET '/api/rest/v1/organizations' is called", func() {
				setupTables()
				req := organization.NewOrganizationServiceListOrganizationsParams()

				c := helpers.OrganizationServiceHTTPClient()
				resp, err := c.OrganizationService.OrganizationServiceListOrganizations(req, auth)

				Then("the fixture organization should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(len(resp.Payload.Result.Organizations)).Should(Equal(1))
					Expect(resp.Payload.Result.Organizations[0].Name).Should(Equal("CHUV"))
				})
				cleanTables()
			})
		})
	})

	Describe("get organization", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route GET '/api/rest/v1/organizations/{id}' is called", func() {
				req := organization.NewOrganizationServiceGetOrganizationParams().WithID(orgFixtureID)

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceGetOrganization(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, "Public", map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

				When("the route GET '/api/rest/v1/organizations/{id}' is called", func() {
					req := organization.NewOrganizationServiceGetOrganizationParams().WithID(orgFixtureID)

					c := helpers.OrganizationServiceHTTPClient()
					_, err := c.OrganizationService.OrganizationServiceGetOrganization(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

			When("the route GET '/api/rest/v1/organizations/{id}' is called", func() {
				setupTables()
				req := organization.NewOrganizationServiceGetOrganizationParams().WithID(orgFixtureID)

				c := helpers.OrganizationServiceHTTPClient()
				resp, err := c.OrganizationService.OrganizationServiceGetOrganization(req, auth)

				Then("the organization should be returned without logo bytes", func() {
					ExpectAPIErr(err).Should(BeNil())
					org := resp.Payload.Result.Organization
					Expect(org.Name).Should(Equal("CHUV"))
					Expect(org.Country).Should(Equal("CH"))
					Expect(org.City).Should(Equal("Lausanne"))
					Expect(org.WebsiteURL).Should(Equal("https://www.chuv.ch/"))
				})
				cleanTables()
			})
		})
	})

	// GetOrganizationLogo returns a raw google.api.HttpBody response (binary bytes with
	// their own Content-Type), not a JSON-shaped payload, so it can't be exercised through
	// the generated OpenAPI client the way the other endpoints are - that client picks its
	// response consumer based on the Content-Type header, and has no consumer registered
	// for arbitrary image types. A plain HTTP request is used instead.
	Describe("get organization logo", func() {

		Given("a valid jwt-token with the authenticated role", func() {

			token := helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)})

			When("the route GET '/api/rest/v1/organizations/{id}/logo' is called", func() {
				setupTables()

				httpReq, reqErr := http.NewRequest(http.MethodGet, "http://"+helpers.ComponentURL()+"/api/rest/v1/organizations/"+orgFixtureID+"/logo", nil)
				Expect(reqErr).Should(BeNil())
				httpReq.Header.Set("Authorization", "Bearer "+token)

				resp, err := http.DefaultClient.Do(httpReq)

				Then("the raw logo bytes and content type should be returned", func() {
					Expect(err).Should(BeNil())
					defer resp.Body.Close()
					Expect(resp.StatusCode).Should(Equal(http.StatusOK))
					Expect(resp.Header.Get("Content-Type")).Should(Equal("image/png"))
					body, readErr := io.ReadAll(resp.Body)
					Expect(readErr).Should(BeNil())
					Expect(body).Should(Equal([]byte{0x89, 0x50, 0x4E, 0x47}))
				})
				cleanTables()
			})
		})
	})

	Describe("create organization", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route POST '/api/rest/v1/organizations' is called", func() {
				req := organization.NewOrganizationServiceCreateOrganizationParams().WithBody(
					&models.ChorusCreateOrganizationRequest{Name: "New Org"},
				)

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceCreateOrganization(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

				When("the route POST '/api/rest/v1/organizations' is called", func() {
					setupBaseTables()
					req := organization.NewOrganizationServiceCreateOrganizationParams().WithBody(
						&models.ChorusCreateOrganizationRequest{Name: "New Org"},
					)

					c := helpers.OrganizationServiceHTTPClient()
					_, err := c.OrganizationService.OrganizationServiceCreateOrganization(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
					cleanTables()
				})
			})
		})

		Given("a platform organization manager jwt", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RolePlatformOrganizationManager.String(), map[string]string{}))

			When("the route POST '/api/rest/v1/organizations' is called", func() {
				setupBaseTables()
				req := organization.NewOrganizationServiceCreateOrganizationParams().WithBody(
					&models.ChorusCreateOrganizationRequest{
						Name:        "EPFL",
						Description: "A description",
						Country:     "CH",
						City:        "Lausanne",
						WebsiteURL:  "https://www.epfl.ch/",
					},
				)

				c := helpers.OrganizationServiceHTTPClient()
				resp, err := c.OrganizationService.OrganizationServiceCreateOrganization(req, auth)

				Then("an organization should be created", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(resp.Payload.Result.Organization.Name).Should(Equal("EPFL"))
					Expect(resp.Payload.Result.Organization.ID).ShouldNot(Equal(""))
				})
				cleanTables()
			})
		})
	})

	Describe("update organization", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route PUT '/api/rest/v1/organizations/{id}' is called", func() {
				req := organization.NewOrganizationServiceUpdateOrganizationParams().WithID(orgFixtureID).WithBody(
					&models.OrganizationServiceUpdateOrganizationBody{Name: "CHUV Renamed"},
				)

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceUpdateOrganization(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

				When("the route PUT '/api/rest/v1/organizations/{id}' is called", func() {
					setupTables()
					req := organization.NewOrganizationServiceUpdateOrganizationParams().WithID(orgFixtureID).WithBody(
						&models.OrganizationServiceUpdateOrganizationBody{Name: "CHUV Renamed"},
					)

					c := helpers.OrganizationServiceHTTPClient()
					_, err := c.OrganizationService.OrganizationServiceUpdateOrganization(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
					cleanTables()
				})
			})
		})

		Given("a platform organization manager jwt", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RolePlatformOrganizationManager.String(), map[string]string{}))

			When("the route PUT '/api/rest/v1/organizations/{id}' is called", func() {
				setupTables()
				req := organization.NewOrganizationServiceUpdateOrganizationParams().WithID(orgFixtureID).WithBody(
					&models.OrganizationServiceUpdateOrganizationBody{Name: "CHUV Renamed"},
				)

				c := helpers.OrganizationServiceHTTPClient()
				resp, err := c.OrganizationService.OrganizationServiceUpdateOrganization(req, auth)

				Then("the organization should be updated", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(resp.Payload.Result.Organization.Name).Should(Equal("CHUV Renamed"))
				})
				cleanTables()
			})
		})
	})

	Describe("delete organization", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route DELETE '/api/rest/v1/organizations/{id}' is called", func() {
				req := organization.NewOrganizationServiceDeleteOrganizationParams().WithID(orgFixtureID)

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceDeleteOrganization(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RoleAuthenticated.String(), map[string]string{"user": fmt.Sprintf("%d", orgTestJWTUser)}))

				When("the route DELETE '/api/rest/v1/organizations/{id}' is called", func() {
					setupTables()
					req := organization.NewOrganizationServiceDeleteOrganizationParams().WithID(orgFixtureID)

					c := helpers.OrganizationServiceHTTPClient()
					_, err := c.OrganizationService.OrganizationServiceDeleteOrganization(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
					cleanTables()
				})
			})
		})

		Given("a platform organization manager jwt", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(orgTestJWTUser, orgTestTenantID, authorization.RolePlatformOrganizationManager.String(), map[string]string{}))

			When("the route DELETE '/api/rest/v1/organizations/{id}' is called", func() {
				setupTables()
				req := organization.NewOrganizationServiceDeleteOrganizationParams().WithID(orgFixtureID)

				c := helpers.OrganizationServiceHTTPClient()
				_, err := c.OrganizationService.OrganizationServiceDeleteOrganization(req, auth)

				Then("the organization should be deleted", func() {
					ExpectAPIErr(err).Should(BeNil())
				})
				cleanTables()
			})
		})
	})
})

func setupBaseTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88700, 'org test tenant');
	`
	helpers.Populate(q)
}

func setupTables() {
	cleanTables()

	q := fmt.Sprintf(`
	INSERT INTO tenants (id, name) VALUES (88700, 'org test tenant');

	INSERT INTO organizations (id, tenantid, name, description, logo, logocontenttype, country, city, websiteurl, createdat, updatedat)
	VALUES (%s, 88700, 'CHUV', 'A description', decode('89504E47', 'hex'), 'image/png', 'CH', 'Lausanne', 'https://www.chuv.ch/', NOW(), NOW());
	`, orgFixtureID)
	helpers.Populate(q)
}

func cleanTables() {
	q := `
	DELETE FROM organizations where tenantid = 88700;
	DELETE FROM tenants where id = 88700;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
