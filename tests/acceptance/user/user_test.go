//go:build acceptance

package user_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	user "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/client/user_service"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pquerna/otp/totp"
)

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	auth := httptransport.BearerToken(t)
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = auth
	}
}

var _ = Describe("user service", func() {
	helpers.Setup()

	Describe("list users", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route GET '/api/rest/v1/users' is called", func() {
				req := user.NewUserServiceListUsersParams()

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceListUsers(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

				When("the route GET '/api/rest/v1/users' is called", func() {
					req := user.NewUserServiceListUsersParams()

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceListUsers(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

			When("the route GET '/api/rest/v1/users' is called", func() {
				setupTables()
				req := user.NewUserServiceListUsersParams()

				c := helpers.UserServiceHTTPClient()
				resp, err := c.UserService.UserServiceListUsers(req, auth)

				Then("users should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(len(resp.Payload.Result.Users)).Should(Equal(2))
				})
				cleanTables()
			})
		})
	})

	Describe("get user", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("then route GET 'api/rest/v1/users/{id} is called", func() {
				req := user.NewUserServiceGetUserParams().WithID("90000")

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceGetUser(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAuthenticated.String()))

				When("the route GET '/api/rest/v1/users/{id}' is called", func() {

					req := user.NewUserServiceGetUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceGetUser(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAdmin.String()))

			When("the route GET '/api/rest/v1/users/{id}' is called", func() {
				setupTables()

				req := user.NewUserServiceGetUserParams().WithID("90000")

				c := helpers.UserServiceHTTPClient()
				resp, err := c.UserService.UserServiceGetUser(req, auth)

				Then("a user should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					me := resp.Payload.Result.User
					Expect(me.Username).Should(Equal("hmoto"))
				})
				cleanTables()
			})
		})
	})

	Describe("get me", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("then route GET 'api/rest/v1/users/me is called", func() {

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceGetUserMe(nil, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

				When("the route GET '/api/rest/v1/users/me' is called", func() {

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceGetUserMe(nil, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route GET '/api/rest/v1/users/me' is called", func() {
				setupTables()

				c := helpers.UserServiceHTTPClient()
				resp, err := c.UserService.UserServiceGetUserMe(nil, auth)

				Then("a user should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					me := resp.Payload.Result.Me
					Expect(me.Username).Should(Equal("hmoto"))
					Expect(me.PasswordChanged).Should(BeFalse())
					Expect(me.TotpEnabled).Should(BeFalse())
				})
				cleanTables()
			})
		})
	})

	Describe("delete user", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {
				req := user.NewUserServiceDeleteUserParams().WithID("90000")

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceDeleteUser(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAuthenticated.String()))

				When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {
					req := user.NewUserServiceDeleteUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceDeleteUser(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})

			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

			When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {
				setupTables()

				req := user.NewUserServiceDeleteUserParams().WithID("90000")

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceDeleteUser(req, auth)

				Then("a user should be deleted", func() {
					ExpectAPIErr(err).Should(BeNil())
				})
				cleanTables()
			})
		})
	})

	Describe("update user", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route PUT '/api/rest/v1/users' is called", func() {
				req := user.NewUserServiceUpdateUserParams().WithBody(
					&models.ChorusUpdateUserRequest{
						User: &models.ChorusUser{
							FirstName: "Bob",
							ID:        "90000",
							LastName:  "Smith",
							Roles:     []string{"admin", "authenticated"},
							Status:    "disabled",
							Username:  "Bobby",
						},
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdateUser(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAuthenticated.String()))

				When("the route PUT '/api/rest/v1/users' is called", func() {
					req := user.NewUserServiceUpdateUserParams().WithBody(
						&models.ChorusUpdateUserRequest{
							User: &models.ChorusUser{
								FirstName: "Bob",
								ID:        "90000",
								LastName:  "Smith",
								Roles:     []string{"admin", "authenticated"},
								Status:    "disabled",
								Username:  "Bobby",
							},
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdateUser(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

			When("the route PUT '/api/rest/v1/users' is called", func() {
				setupTables()

				req := user.NewUserServiceUpdateUserParams().WithBody(
					&models.ChorusUpdateUserRequest{
						User: &models.ChorusUser{
							FirstName: "Bob",
							ID:        "90000",
							LastName:  "Smith",
							Roles:     []string{"admin", "authenticated"},
							Status:    "disabled",
							Username:  "Bobby",
						},
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdateUser(req, auth)

				Then("a user should be updated", func() {
					ExpectAPIErr(err).Should(BeNil())
				})
				cleanTables()
			})
		})

		Given("a valid jwt-token", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

			When("the route PUT '/api/rest/v1/users' is called with an invalid role", func() {
				setupTables()

				req := user.NewUserServiceUpdateUserParams().WithBody(
					&models.ChorusUpdateUserRequest{
						User: &models.ChorusUser{
							FirstName: "Bob",
							ID:        "90000",
							LastName:  "Smith",
							Roles:     []string{"admin", "chorus"},
							Status:    "disabled",
							Username:  "Bobby",
						},
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdateUser(req, auth)

				Then("a bad request error should be returned", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
				})
				cleanTables()
			})
		})
	})

	Describe("update password", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {
				req := user.NewUserServiceUpdatePasswordParams().WithBody(
					&models.ChorusUpdatePasswordRequest{
						CurrentPassword: "toto",
						NewPassword:     "titi",
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

				When("the route PUT 'api/rest/v1/users/me/password' is called", func() {
					req := user.NewUserServiceUpdatePasswordParams().WithBody(
						&models.ChorusUpdatePasswordRequest{
							CurrentPassword: "toto",
							NewPassword:     "titi",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdatePassword(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("an identified user and a weak password", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {
				setupTables()
				req := user.NewUserServiceUpdatePasswordParams().WithBody(
					&models.ChorusUpdatePasswordRequest{
						CurrentPassword: "johnPassword",
						NewPassword:     "titi",
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

				Then("an error should be returned", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
				})
				cleanTables()
			})
		})

		Given("an identified user and a strong password without TOTP", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {
				setupTables()
				req := user.NewUserServiceUpdatePasswordParams().WithBody(
					&models.ChorusUpdatePasswordRequest{
						CurrentPassword: "johnPassword",
						NewPassword:     "titiTOTO12345??",
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

				Then("a user's password should be updated", func() {
					ExpectAPIErr(err).Should(BeNil())
				})
				cleanTables()
			})
		})

		Given("an identified user and a strong password with TOTP", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {
				setupTablesWithTotpUser()
				req := user.NewUserServiceUpdatePasswordParams().WithBody(
					&models.ChorusUpdatePasswordRequest{
						CurrentPassword: "johnPassword",
						NewPassword:     "titiTOTO12345??",
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

				Then("a user's password should be updated", func() {
					ExpectAPIErr(err).Should(BeNil())
				})
				cleanTables()
			})
		})
	})

	Describe("reset totp", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {
				req := user.NewUserServiceResetTotpParams().WithBody(
					&models.ChorusResetTotpRequest{Password: "johnPassword"},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceResetTotp(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

				When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "johnPassword"},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceResetTotp(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("an identified user but a wrong password", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {
				setupTablesWithTotpUser()
				req := user.NewUserServiceResetTotpParams().WithBody(
					&models.ChorusResetTotpRequest{Password: "wrong password"},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceResetTotp(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
				cleanTables()
			})
		})

		Given("an identified user and a correct password", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {
				setupTablesWithTotpUser()
				req := user.NewUserServiceResetTotpParams().WithBody(
					&models.ChorusResetTotpRequest{Password: "johnPassword"},
				)

				c := helpers.UserServiceHTTPClient()
				res, err := c.UserService.UserServiceResetTotp(req, auth)

				Then("a totpSecret and recovery codes should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(res).ShouldNot(BeNil())
					Expect(res.Payload.Result.TotpSecret).ShouldNot((Equal("")))
					Expect(len(res.Payload.Result.TotpRecoveryCodes)).Should(BeNumerically(">=", 10))
				})
				cleanTables()
			})
		})
	})

	Describe("enable totp", func() {

		Given("an invalid jwt-token", func() {

			auth := getAuthAsClientOpts("invalid")

			When("the route POST '/api/rest/v1/users/me/totp/enable' is called", func() {
				req := user.NewUserServiceEnableTotpParams().WithBody(
					&models.ChorusEnableTotpRequest{
						Totp: "totp",
					},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceEnableTotp(req, auth)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

				When("the route POST '/api/rest/v1/users/me/totp/enable' is called", func() {
					req := user.NewUserServiceEnableTotpParams().WithBody(
						&models.ChorusEnableTotpRequest{
							Totp: "totp",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceEnableTotp(req, auth)

					Then("a permission error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})
	})

	Describe("reset and enable totp", func() {

		Given("an identified user, a correct password and a correct totp", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the routes POST '/api/rest/v1/users/me/totp/reset' and POST '/api/rest/v1/users/me/totp/enable' are called", func() {
				setupTablesWithTotpUser()
				req := user.NewUserServiceResetTotpParams().WithBody(
					&models.ChorusResetTotpRequest{Password: "johnPassword"},
				)

				c := helpers.UserServiceHTTPClient()
				res, err := c.UserService.UserServiceResetTotp(req, auth)

				if err != nil {
					Then("Totp is now enabled for the user and no error should be returned", func() {
						Fail(fmt.Sprintf("Error resetting TOTP: %v", err))
					})
					cleanTables()
					return
				}

				totpSecret := res.Payload.Result.TotpSecret
				totp, _ := totp.GenerateCode(totpSecret, time.Now().UTC())

				reqEnable := user.NewUserServiceEnableTotpParams().WithBody(
					&models.ChorusEnableTotpRequest{
						Totp: totp,
					},
				)

				_, errEnable := c.UserService.UserServiceEnableTotp(reqEnable, auth)

				Then("Totp is now enabled for the user and no error should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					ExpectAPIErr(errEnable).Should(BeNil())
				})
				cleanTables()
			})
		})

		Given("an identified user, a correct password but an incorrect totp", func() {

			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, model.RoleAuthenticated.String()))

			When("the routes POST '/api/rest/v1/users/me/totp/reset' and POST '/api/rest/v1/users/me/totp/enable' are called", func() {
				setupTablesWithTotpUser()
				req := user.NewUserServiceResetTotpParams().WithBody(
					&models.ChorusResetTotpRequest{Password: "johnPassword"},
				)

				c := helpers.UserServiceHTTPClient()
				_, err := c.UserService.UserServiceResetTotp(req, auth)

				reqEnable := user.NewUserServiceEnableTotpParams().WithBody(
					&models.ChorusEnableTotpRequest{
						Totp: "1234567",
					},
				)

				_, err = c.UserService.UserServiceEnableTotp(reqEnable, auth)

				Then("Totp is not enabled for the user and an error should be returned", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
				})
				cleanTables()
			})
		})
	})

	Describe("create user", func() {

		// toggle if switching to admin only registration
		// Given("an invalid jwt-token", func() {
		// 	setupBaseTables()
		// 	auth := getAuthAsClientOpts("invalid")

		// 	When("the route POST '/api/rest/v1/users' is called", func() {
		// 		req := user.NewUserServiceCreateUserParams().WithBody(
		// 			&models.ChorusCreateUserRequest{User: &models.ChorusUser{
		// 				FirstName: "first", LastName: "last", Username: "user",
		// 				Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
		// 				TotpEnabled: true,
		// 			}},
		// 		)

		// 		c := helpers.UserServiceHTTPClient()
		// 		_, err := c.UserService.UserServiceCreateUser(req, auth)

		// 		Then("an authentication error should be raised", func() {
		// 			ExpectAPIErr(err).ShouldNot(BeNil())
		// 			Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
		// 		})
		// 		cleanTables()
		// 	})
		// })

		// Given("a valid jwt-token", func() {

		// 	Given("an unauthorized role", func() {
		// 		setupBaseTables()
		// 		auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "client"))

		// 		When("the route POST '/api/rest/v1/users' is called", func() {
		// 			req := user.NewUserServiceCreateUserParams().WithBody(
		// 				&models.ChorusCreateUserRequest{User: &models.ChorusUser{
		// 					FirstName: "first", LastName: "last", Username: "user",
		// 					Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
		// 					TotpEnabled: true,
		// 				}},
		// 			)

		// 			c := helpers.UserServiceHTTPClient()
		// 			_, err := c.UserService.UserServiceCreateUser(req, auth)

		// 			Then("a permission error should be raised", func() {
		// 				ExpectAPIErr(err).ShouldNot(BeNil())
		// 				Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
		// 			})
		// 			cleanTables()
		// 		})
		// 	})
		// })

		// Given("a valid jwt-token", func() {

		// 	auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

		// 	When("the route POST '/api/rest/v1/users' is called with an invalid role", func() {
		// 		setupTables()
		// 		req := user.NewUserServiceCreateUserParams().WithBody(
		// 			&models.ChorusCreateUserRequest{User: &models.ChorusUser{
		// 				FirstName: "first", LastName: "last", Username: "user",
		// 				Password: "pass", Status: "active", Roles: []string{"admin", "chorus"},
		// 			}},
		// 		)

		// 		c := helpers.UserServiceHTTPClient()
		// 		_, err := c.UserService.UserServiceCreateUser(req, auth)

		// 		Then("a bad request error should be returned", func() {
		// 			ExpectAPIErr(err).ShouldNot(BeNil())
		// 			Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
		// 		})
		// 		cleanTables()
		// 	})
		// })

		// Given("a valid jwt-token", func() {

		// 	Given("an empty field in request", func() {

		// 		auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

		// 		When("the route POST '/api/rest/v1/users' is called", func() {
		// 			req := user.NewUserServiceCreateUserParams().WithBody(
		// 				&models.ChorusCreateUserRequest{User: &models.ChorusUser{
		// 					FirstName: "first", LastName: "", Username: "user",
		// 					Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
		// 					TotpEnabled: true,
		// 				}},
		// 			)

		// 			c := helpers.UserServiceHTTPClient()
		// 			_, err := c.UserService.UserServiceCreateUser(req, auth)

		// 			Then("a validation error should be raised", func() {
		// 				ExpectAPIErr(err).ShouldNot(BeNil())
		// 				Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
		// 			})
		// 		})
		// 	})
		// })
		Given("no auth", func() {

			Given("an empty field in request", func() {

				When("the route POST '/api/rest/v1/users' is called", func() {
					req := user.NewUserServiceCreateUserParams().WithBody(
						&models.ChorusCreateUserRequest{User: &models.ChorusUser{
							FirstName: "first", LastName: "last", Username: "user",
							Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
							TotpEnabled: true,
						}},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceCreateUser(req)

					Then("a validation error should be raised", func() {
						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
					})
				})
			})
		})

		Given("no auth", func() {

			Given("an empty field in request", func() {

				When("the route POST '/api/rest/v1/users' is called", func() {
					setupBaseTables()
					req := user.NewUserServiceCreateUserParams().WithBody(
						&models.ChorusCreateUserRequest{User: &models.ChorusUser{
							FirstName: "firstfirst", LastName: "last", Username: "user88888",
							Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
							TotpEnabled: true,
						}},
					)

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceCreateUser(req)

					Then("an user id should be returned", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.User.ID).ShouldNot(Equal(0))
					})
					cleanTables()
				})
			})
		})

		/*
			Given("a valid jwt-token", func() {

				auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, model.RoleAdmin.String()))

				When("the route POST '/api/rest/v1/users' is called", func() {
					setupTables()
					req := user.NewUserServiceCreateUserParams().WithBody(
						&models.ChorusUser{
							FirstName: "first", LastName: "last", Username: "user",
							Password: "pass", Status: "status", Roles: []string{"client"},
							TotpEnabled: true,
						},
					)

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceCreateUser(req, auth)

					Then("an user id should be returned", func() {
						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.ID).ShouldNot(Equal(0))
					})
					cleanTables()
				})
			})
		*/
	})
})

func setupBaseTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');


	INSERT INTO roles (id, name) VALUES (1, 'admin');
	INSERT INTO roles (id, name) VALUES (2, 'operator');
	INSERT INTO roles (id, name) VALUES (3, 'chorus');
	INSERT INTO roles (id, name) VALUES (4, 'authenticated');
	`
	helpers.Populate(q)
}

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id,tenantid, firstname, lastname, username, password, status, createdat, updatedat)
	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active', NOW(), NOW());
	INSERT INTO users (id,tenantid, firstname, lastname, username, password, status, createdat, updatedat)
	VALUES (90001,88888, 'jane', 'doe', 'jadoe', '$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42', 'disabled', NOW(), NOW());

	INSERT INTO roles (id, name) VALUES (1, 'admin');
	INSERT INTO roles (id, name) VALUES (2, 'operator');
	INSERT INTO roles (id, name) VALUES (3, 'chorus');
	INSERT INTO roles (id, name) VALUES (4, 'authenticated');

	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, 1);
	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, 2);
	INSERT INTO user_role (id, userid, roleid) VALUES(92003, 90001, 3);
	INSERT INTO user_role (id, userid, roleid) VALUES(92004, 90001, 4);
	`
	helpers.Populate(q)
}

func setupTablesWithTotpUser() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id,tenantid, firstname, lastname, username, password, status, createdat, updatedat, totpsecret)
	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active',
			NOW(), NOW(), 'ohKtu9PFHMquP5Zemcfb4XFQ8TuYnA5Gk1txooQINWL2AbhonyGW0H66zmX8YdUEDEZPYGjOCDPBOF9W');

	INSERT INTO roles (id, name) VALUES (1, 'admin');
	INSERT INTO roles (id, name) VALUES (2, 'operator');
	INSERT INTO roles (id, name) VALUES (3, 'chorus');
	INSERT INTO roles (id, name) VALUES (4, 'authenticated');

	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, 1);
	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, 2);
	INSERT INTO user_role (id, userid, roleid) VALUES(92003, 90000, 4);

	INSERT INTO totp_recovery_codes (id, userid, tenantid, code)
	VALUES (88888, 90000, 88888, '0Uu+C4s1i+mrS7pqmI2SHJe+Hcg3l4K/ylusXoIv25RE6qEUyRY='),
		(88889, 90000, 88888, '0YZWPkeRISwyAeZsQ2otY+JMdR1P6N42NoN0UOxbPh7tnioAvF4=');
	`
	helpers.Populate(q)
}

func cleanTables() {
	q := `
	DELETE FROM notifications_read_by where tenantid = 88888;
	DELETE FROM notifications where tenantid = 88888;
	DELETE FROM totp_recovery_codes where tenantid = 88888;
	DELETE FROM user_role where id in (92001,92002,92003) OR userid=90000 OR roleid in (1,2,3, 4) OR userid in (SELECT id FROM users WHERE tenantid = 88888 or username='user88888');
	
	DELETE FROM users where tenantid = 88888 or username='user88888';
	DELETE FROM roles where id in (1,2,3, 4);
	DELETE FROM tenants where id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
