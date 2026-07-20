//go:build acceptance

package user_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	user "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/client/user_service"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo/v2"
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

	AfterEach(func() {
		cleanTables()
	})

	Describe("list users", func() {

		Given("an invalid jwt-token", func() {

			When("the route GET '/api/rest/v1/users' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceListUsersParams()

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceListUsers(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route GET '/api/rest/v1/users' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "Public", map[string]string{"user": "1"}))
						req := user.NewUserServiceListUsersParams()

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceListUsers(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route GET '/api/rest/v1/users' is called", func() {

				Then("users should be returned with email", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))
					req := user.NewUserServiceListUsersParams()

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceListUsers(req, auth)

					ExpectAPIErr(err).Should(BeNil())
					Expect(len(resp.Payload.Result.Users)).Should(Equal(2))
					for _, u := range resp.Payload.Result.Users {
						Expect(u.Email).ShouldNot(Equal(""))
					}
				})
			})
		})
	})

	Describe("get user", func() {

		Given("an invalid jwt-token", func() {

			When("then route GET 'api/rest/v1/users/{id} is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceGetUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceGetUser(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route GET '/api/rest/v1/users/{id}' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "1"}))
						req := user.NewUserServiceGetUserParams().WithID("90000")

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceGetUser(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route GET '/api/rest/v1/users/{id}' is called", func() {

				Then("a user should be returned", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceGetUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceGetUser(req, auth)

					ExpectAPIErr(err).Should(BeNil())
					me := resp.Payload.Result.User
					Expect(me.Username).Should(Equal("hmoto"))
					Expect(me.Email).Should(Equal("hmoto@example.com"))
				})
			})
		})
	})

	Describe("get me", func() {

		Given("an invalid jwt-token", func() {

			When("then route GET 'api/rest/v1/users/me is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceGetUserMe(nil, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route GET '/api/rest/v1/users/me' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "Public", map[string]string{"user": "1"}))

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceGetUserMe(nil, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route GET '/api/rest/v1/users/me' is called", func() {

				Then("a user should be returned", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceGetUserMe(nil, auth)

					ExpectAPIErr(err).Should(BeNil())
					me := resp.Payload.Result.Me
					Expect(me.Username).Should(Equal("hmoto"))
					Expect(me.Email).Should(Equal("hmoto@example.com"))
					Expect(me.PasswordChanged).Should(BeFalse())
					Expect(me.TotpEnabled).Should(BeFalse())
				})
			})
		})
	})

	Describe("delete user", func() {

		Given("an invalid jwt-token", func() {

			When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceDeleteUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceDeleteUser(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "1"}))
						req := user.NewUserServiceDeleteUserParams().WithID("90000")

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceDeleteUser(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route DELETE '/api/rest/v1/users/{id}' is called", func() {

				Then("a user should be deleted", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))
					req := user.NewUserServiceDeleteUserParams().WithID("90000")

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceDeleteUser(req, auth)

					ExpectAPIErr(err).Should(BeNil())
				})
			})
		})
	})

	Describe("update user", func() {

		Given("an invalid jwt-token", func() {

			When("the route PUT '/api/rest/v1/users' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceUpdateUserParams().WithBody(
						&models.ChorusUser{
							FirstName: "Bob",
							ID:        "90000",
							LastName:  "Smith",
							Roles:     []string{"admin", "authenticated"},
							Status:    "disabled",
							Username:  "Bobby",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdateUser(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route PUT '/api/rest/v1/users' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "1"}))
						req := user.NewUserServiceUpdateUserParams().WithBody(
							&models.ChorusUser{
								FirstName: "Bob",
								ID:        "90000",
								LastName:  "Smith",
								Roles:     []string{"admin", "authenticated"},
								Status:    "disabled",
								Username:  "Bobby",
							},
						)

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceUpdateUser(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("a valid jwt-token", func() {

			When("the route PUT '/api/rest/v1/users' is called", func() {

				Then("a user should be updated", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))
					req := user.NewUserServiceUpdateUserParams().WithBody(
						&models.ChorusUser{
							FirstName: "Bob",
							ID:        "90000",
							LastName:  "Smith",
							Email:     "bob@example.com",
							Roles:     []string{"admin", "authenticated"},
							Status:    "disabled",
							Username:  "Bobby",
							Source:    "keycloak",
						},
					)

					c := helpers.UserServiceHTTPClient()
					resp, err := c.UserService.UserServiceUpdateUser(req, auth)

					ExpectAPIErr(err).Should(BeNil())
					Expect(resp.Payload.Result.User.Email).Should(Equal("bob@example.com"))
				})
			})
		})
	})

	Describe("update password", func() {

		Given("an invalid jwt-token", func() {

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceUpdatePasswordParams().WithBody(
						&models.ChorusUpdatePasswordRequest{
							CurrentPassword: "toto",
							NewPassword:     "titi",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdatePassword(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route PUT 'api/rest/v1/users/me/password' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "Public", map[string]string{"user": "1"}))
						req := user.NewUserServiceUpdatePasswordParams().WithBody(
							&models.ChorusUpdatePasswordRequest{
								CurrentPassword: "toto",
								NewPassword:     "titi",
							},
						)

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceUpdatePassword(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("an identified user and a weak password", func() {

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {

				Then("an error should be returned", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceUpdatePasswordParams().WithBody(
						&models.ChorusUpdatePasswordRequest{
							CurrentPassword: "johnPassword",
							NewPassword:     "titi",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdatePassword(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
				})
			})
		})

		Given("an identified user and a strong password without TOTP", func() {

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {

				Then("a user's password should be updated", func() {
					setupTables()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceUpdatePasswordParams().WithBody(
						&models.ChorusUpdatePasswordRequest{
							CurrentPassword: "johnPassword",
							NewPassword:     "titiTOTO12345??",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdatePassword(req, auth)

					ExpectAPIErr(err).Should(BeNil())
				})
			})
		})

		Given("an identified user and a strong password with TOTP", func() {

			When("the route PUT 'api/rest/v1/users/me/password' is called", func() {

				Then("a user's password should be updated", func() {
					setupTablesWithTotpUser()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceUpdatePasswordParams().WithBody(
						&models.ChorusUpdatePasswordRequest{
							CurrentPassword: "johnPassword",
							NewPassword:     "titiTOTO12345??",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceUpdatePassword(req, auth)

					ExpectAPIErr(err).Should(BeNil())
				})
			})
		})
	})

	Describe("reset totp", func() {

		Given("an invalid jwt-token", func() {

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "johnPassword"},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceResetTotp(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "Public", map[string]string{"user": "1"}))
						req := user.NewUserServiceResetTotpParams().WithBody(
							&models.ChorusResetTotpRequest{Password: "johnPassword"},
						)

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceResetTotp(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})

		Given("an identified user but a wrong password", func() {

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {

				Then("an authentication error should be raised", func() {
					setupTablesWithTotpUser()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "wrong password"},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceResetTotp(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("an identified user and a correct password", func() {

			When("the route POST '/api/rest/v1/users/me/totp/reset' is called", func() {

				Then("a totpSecret and recovery codes should be returned", func() {
					setupTablesWithTotpUser()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "johnPassword"},
					)

					c := helpers.UserServiceHTTPClient()
					res, err := c.UserService.UserServiceResetTotp(req, auth)

					ExpectAPIErr(err).Should(BeNil())
					Expect(res).ShouldNot(BeNil())
					Expect(res.Payload.Result.TotpSecret).ShouldNot((Equal("")))
					Expect(len(res.Payload.Result.TotpRecoveryCodes)).Should(BeNumerically(">=", 10))
				})
			})
		})
	})

	Describe("enable totp", func() {

		Given("an invalid jwt-token", func() {

			When("the route POST '/api/rest/v1/users/me/totp/enable' is called", func() {

				Then("an authentication error should be raised", func() {
					auth := getAuthAsClientOpts("invalid")
					req := user.NewUserServiceEnableTotpParams().WithBody(
						&models.ChorusEnableTotpRequest{
							Totp: "totp",
						},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceEnableTotp(req, auth)

					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
			})
		})

		Given("a valid jwt-token", func() {

			Given("an unauthorized role", func() {

				When("the route POST '/api/rest/v1/users/me/totp/enable' is called", func() {

					Then("a permission error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, "Public", map[string]string{"user": "1"}))
						req := user.NewUserServiceEnableTotpParams().WithBody(
							&models.ChorusEnableTotpRequest{
								Totp: "totp",
							},
						)

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceEnableTotp(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
					})
				})
			})
		})
	})

	Describe("reset and enable totp", func() {

		Given("an identified user, a correct password and a correct totp", func() {

			When("the routes POST '/api/rest/v1/users/me/totp/reset' and POST '/api/rest/v1/users/me/totp/enable' are called", func() {

				Then("Totp is now enabled for the user and no error should be returned", func() {
					setupTablesWithTotpUser()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "johnPassword"},
					)

					c := helpers.UserServiceHTTPClient()
					res, err := c.UserService.UserServiceResetTotp(req, auth)
					ExpectAPIErr(err).Should(BeNil())

					totpSecret := res.Payload.Result.TotpSecret
					code, _ := totp.GenerateCode(totpSecret, time.Now().UTC())

					reqEnable := user.NewUserServiceEnableTotpParams().WithBody(
						&models.ChorusEnableTotpRequest{
							Totp: code,
						},
					)

					_, errEnable := c.UserService.UserServiceEnableTotp(reqEnable, auth)

					ExpectAPIErr(errEnable).Should(BeNil())
				})
			})
		})

		Given("an identified user, a correct password but an incorrect totp", func() {

			When("the routes POST '/api/rest/v1/users/me/totp/reset' and POST '/api/rest/v1/users/me/totp/enable' are called", func() {

				Then("Totp is not enabled for the user and an error should be returned", func() {
					setupTablesWithTotpUser()
					auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))
					req := user.NewUserServiceResetTotpParams().WithBody(
						&models.ChorusResetTotpRequest{Password: "johnPassword"},
					)

					c := helpers.UserServiceHTTPClient()
					_, err := c.UserService.UserServiceResetTotp(req, auth)
					ExpectAPIErr(err).Should(BeNil())

					reqEnable := user.NewUserServiceEnableTotpParams().WithBody(
						&models.ChorusEnableTotpRequest{
							Totp: "1234567",
						},
					)

					_, errEnable := c.UserService.UserServiceEnableTotp(reqEnable, auth)

					ExpectAPIErr(errEnable).ShouldNot(BeNil())
				})
			})
		})
	})

	Describe("create user", func() {

		Given("a platform manager jwt", func() {

			Given("an empty field in request", func() {

				When("the route POST '/api/rest/v1/users' is called", func() {

					Then("a validation error should be raised", func() {
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, string(authorization.RolePlateformUserManager), map[string]string{}))

						req := user.NewUserServiceCreateUserParams().WithBody(
							&models.ChorusUser{
								LastName: "last", Username: "user",
								Password: "pass", Status: "active", Roles: []string{"admin", "authenticated"},
								TotpEnabled: true,
							},
						)

						c := helpers.UserServiceHTTPClient()
						_, err := c.UserService.UserServiceCreateUser(req, auth)

						ExpectAPIErr(err).ShouldNot(BeNil())
						Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusBadRequest)))
					})
				})
			})
		})

		Given("no auth", func() {

			Given("a complete request", func() {

				When("the route POST '/api/rest/v1/users' is called", func() {

					Then("a user should be returned", func() {
						setupBaseTables()
						auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, string(authorization.RolePlateformUserManager), map[string]string{}))

						req := user.NewUserServiceCreateUserParams().WithBody(
							&models.ChorusUser{
								FirstName: "first", LastName: "last", Username: "user88888",
								Email:    "user88888@example.com",
								Password: "pass", Status: "active",
								TotpEnabled: true,
							},
						)

						c := helpers.UserServiceHTTPClient()
						resp, err := c.UserService.UserServiceCreateUser(req, auth)

						ExpectAPIErr(err).Should(BeNil())
						Expect(resp.Payload.Result.User).ShouldNot(Equal(nil))
						Expect(resp.Payload.Result.User.Email).Should(Equal("user88888@example.com"))
					})
				})
			})
		})
	})
})

func setupBaseTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	`
	helpers.Populate(q)
}

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id,tenantid, firstname, lastname, username, email, password, status, createdat, updatedat)
	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', 'hmoto@example.com', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active', NOW(), NOW());
	INSERT INTO users (id,tenantid, firstname, lastname, username, email, password, status, createdat, updatedat)
	VALUES (90001,88888, 'jane', 'doe', 'jadoe', 'jadoe@example.com', '$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42', 'disabled', NOW(), NOW());

	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, (SELECT id FROM role_definitions WHERE name='Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, (SELECT id FROM role_definitions WHERE name='SuperAdmin'));
	INSERT INTO user_role (id, userid, roleid) VALUES(92003, 90001, (SELECT id FROM role_definitions WHERE name='Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES(92004, 90001, (SELECT id FROM role_definitions WHERE name='SuperAdmin'));
	`
	helpers.Populate(q)
}

func setupTablesWithTotpUser() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id,tenantid, firstname, lastname, username, email, password, status, createdat, updatedat, totpsecret)
	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', 'hmoto@example.com', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active',
			NOW(), NOW(), 'ohKtu9PFHMquP5Zemcfb4XFQ8TuYnA5Gk1txooQINWL2AbhonyGW0H66zmX8YdUEDEZPYGjOCDPBOF9W');

	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, (SELECT id FROM role_definitions WHERE name='Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, (SELECT id FROM role_definitions WHERE name='Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES(92003, 90000, (SELECT id FROM role_definitions WHERE name='SuperAdmin'));

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
	DELETE FROM user_role where id in (92001,92002,92003) OR userid=90000 OR roleid in (101,102) OR userid in (SELECT id FROM users WHERE tenantid = 88888 or username='user88888');
	DELETE FROM role_definitions where id in (101,102);
	DELETE FROM users where tenantid = 88888 or username='user88888';
	DELETE FROM tenants where id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
