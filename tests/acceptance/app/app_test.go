//go:build acceptance

package app_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	app "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/app/client/app_service"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/app/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	auth := httptransport.BearerToken(t)
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = auth
	}
}

var _ = Describe("app service", func() {
	helpers.Setup()

	Describe("list apps", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route GET '/api/rest/v1/apps' is called", func() {
				req := app.NewAppServiceListAppsParams()

				c := helpers.AppServiceHTTPClient()
				_, err := c.AppService.AppServiceListApps(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		// Given("a valid jwt-token", func() {

		// 	Given("an unauthorized role", func() {

		// 		auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

		// 		When("the route GET '/api/rest/v1/apps' is called", func() {
		// 			req := app.NewAppServiceListAppsParams()

		// 			c := helpers.AppServiceHTTPClient()
		// 			_, err := c.AppService.AppServiceListApps(req, auth)

		// 			Then("a permission error should be raised", func() {
		// 				ExpectAPIErr(err).ShouldNot(BeNil())
		// 				Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
		// 			})
		// 		})
		// 	})
		// })

		Given("a valid jwt-token", func() {

			Given("an authenticated role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAuthenticated.String()))

				When("the route GET '/api/rest/v1/apps' is called", func() {
					setupTables()
					req := app.NewAppServiceListAppsParams()

					c := helpers.AppServiceHTTPClient()
					resp, err := c.AppService.AppServiceListApps(req, auth)

					Then("apps should be returned", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(len(resp.Payload.Result.Apps)).Should(Equal(2))
					})
					cleanTables()
				})
			})
		})

		Describe("get app", func() {

			Given("an invalid jwt-token", func() {

				auth := getAuthAsClientOpts("invalid")

				When("the route GET '/api/rest/v1/apps/{id}' is called", func() {
					req := app.NewAppServiceGetAppParams().WithID("90000")

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceGetApp(req, auth)

					Then("an authentication error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
					})
				})
			})

			Given("a valid jwt-token", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

				When("a valid app id is provided", func() {
					setupTables()
					req := app.NewAppServiceGetAppParams().WithID("90000")

					c := helpers.AppServiceHTTPClient()
					resp, err := c.AppService.AppServiceGetApp(req, auth)

					Then("the app should be returned", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.App.ID).Should(Equal("90000"))
						Expect(resp.Payload.Result.App.Name).Should(Equal("test-app-1"))
					})
					cleanTables()
				})

				When("an invalid app id is provided", func() {
					setupTables()
					req := app.NewAppServiceGetAppParams().WithID("99999")

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceGetApp(req, auth)

					Then("a not found error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusNotFound)))
					})
					cleanTables()
				})
			})
		})

		Describe("create app", func() {

			Given("an invalid jwt-token", func() {

				auth := getAuthAsClientOpts("invalid")

				When("the route POST '/api/rest/v1/apps' is called", func() {
					newApp := &models.ChorusApp{
						Name:        "new-test-app",
						Description: "A test app for acceptance testing",
					}
					createReq := &models.ChorusCreateAppRequest{
						App: newApp,
					}
					req := app.NewAppServiceCreateAppParams().WithBody(createReq)

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceCreateApp(req, auth)

					Then("an authentication error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
					})
				})
			})

			Given("a valid jwt-token", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

				When("valid app data is provided", func() {
					setupTables()
					newApp := &models.ChorusApp{
						Name:        "new-test-app",
						Description: "A test app for acceptance testing",
					}
					createReq := &models.ChorusCreateAppRequest{
						App: newApp,
					}
					req := app.NewAppServiceCreateAppParams().WithBody(createReq)

					c := helpers.AppServiceHTTPClient()
					resp, err := c.AppService.AppServiceCreateApp(req, auth)

					Then("the app should be created", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.App.Name).Should(Equal("new-test-app"))
						Expect(resp.Payload.Result.App.Description).Should(Equal("A test app for acceptance testing"))
						Expect(resp.Payload.Result.App.ID).ShouldNot(BeEmpty())
					})
					cleanTables()
				})

				When("invalid app data is provided", func() {
					setupTables()
					newApp := &models.ChorusApp{
						Name: "", // Empty name should cause validation error
					}
					createReq := &models.ChorusCreateAppRequest{
						App: newApp,
					}
					req := app.NewAppServiceCreateAppParams().WithBody(createReq)

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceCreateApp(req, auth)

					Then("a validation error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
					})
					cleanTables()
				})
			})
		})

		Describe("update app", func() {

			Given("a valid jwt-token", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

				When("valid app data is provided", func() {
					setupTables()
					updatedApp := &models.ChorusApp{
						ID:          "90000",
						Name:        "updated-test-app",
						Description: "Updated description",
					}
					updateReq := &models.ChorusUpdateAppRequest{
						App: updatedApp,
					}
					req := app.NewAppServiceUpdateAppParams().WithBody(updateReq)

					c := helpers.AppServiceHTTPClient()
					resp, err := c.AppService.AppServiceUpdateApp(req, auth)

					Then("the app should be updated", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.App.Name).Should(Equal("updated-test-app"))
						Expect(resp.Payload.Result.App.Description).Should(Equal("Updated description"))
					})
					cleanTables()
				})
			})
		})

		Describe("delete app", func() {

			Given("a valid jwt-token", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

				When("a valid app id is provided", func() {
					setupTables()
					req := app.NewAppServiceDeleteAppParams().WithID("90000")

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceDeleteApp(req, auth)

					Then("the app should be deleted", func() {
						ExpectAPIErr(err).Should(BeNil())
					})
					cleanTables()
				})

				When("an invalid app id is provided", func() {
					setupTables()
					req := app.NewAppServiceDeleteAppParams().WithID("99999")

					c := helpers.AppServiceHTTPClient()
					_, err := c.AppService.AppServiceDeleteApp(req, auth)

					Then("a not found error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusNotFound)))
					})
					cleanTables()
				})
			})
		})
	})
})

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, createdat, updatedat)
	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active', NOW(), NOW());

	INSERT INTO roles (id, name) VALUES (1, 'admin');
	INSERT INTO roles (id, name) VALUES (2, 'operator');

	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, 1);

	INSERT INTO apps (id, tenantid, userid, name, description, createdat, updatedat)
	VALUES (90000, 88888, 90000, 'test-app-1', 'First test app', NOW(), NOW());
	INSERT INTO apps (id, tenantid, userid, name, description, createdat, updatedat)
	VALUES (90001, 88888, 90000, 'test-app-2', 'Second test app', NOW(), NOW());
	`
	helpers.Populate(q)
}

func cleanTables() {
	q := `
	DELETE FROM apps WHERE tenantid = 88888;
	DELETE FROM user_role WHERE id = 92001;
	DELETE FROM users WHERE tenantid = 88888;
	DELETE FROM roles WHERE id IN (1,2);
	DELETE FROM tenants WHERE id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
