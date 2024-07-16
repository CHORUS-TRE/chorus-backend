//go:build acceptance

package authentication_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	authentication "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/authentication/client/authentication_service"
	models "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/authentication/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pquerna/otp/totp"
)

const (
	johnPassword     = "johnPassword"
	janePassword     = "janePassword"
	johnPasswordHash = "$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO"
	janePasswordHash = "$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42"

	johnTOTPSecret = "2UAQQOLQSXTAYP3A7A2TQDKMKQB4Y45D"
)

var _ = Describe("authentication service", func() {
	helpers.Setup()

	Describe("authenticate user", func() {

		Given("an inexistent user", func() {

			When("the route POST '/api/rest/v1/authentication/token' is called", func() {

				setupTables()
				req := authentication.NewAuthenticationServiceAuthenticateParams().WithBody(
					&models.ChorusCredentials{
						Username: "incognito",
						Password: "superpassword",
					},
				)

				c := helpers.AuthenticationServiceHTTPClient()
				_, err := c.AuthenticationService.AuthenticationServiceAuthenticate(req)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
				cleanTables()
			})
		})

		Given("an inactive user", func() {

			When("the route POST '/api/rest/v1/authentication/token' is called", func() {

				setupTables()

				req := authentication.NewAuthenticationServiceAuthenticateParams().WithBody(
					&models.ChorusCredentials{
						Username: "jadoe",
						Password: janePasswordHash,
					},
				)

				c := helpers.AuthenticationServiceHTTPClient()
				_, err := c.AuthenticationService.AuthenticationServiceAuthenticate(req)

				Then("an authentication error should be raised", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
				cleanTables()
			})
		})

		Given("a valid user", func() {

			When("the route POST '/api/rest/v1/authentication/token' is called", func() {

				setupTables()

				code, _ := totp.GenerateCode(johnTOTPSecret, time.Now())
				req := authentication.NewAuthenticationServiceAuthenticateParams().WithBody(
					&models.ChorusCredentials{
						Username: "jodoe",
						Password: johnPassword,
						Totp:     code,
					},
				)

				c := helpers.AuthenticationServiceHTTPClient()
				resp, err := c.AuthenticationService.AuthenticationServiceAuthenticate(req)

				Then("a jwt-token should be returned", func() {
					ExpectAPIErr(err).Should(BeNil())
					Expect(resp.Payload.Result.Token).ShouldNot(Equal(""))
				})
				cleanTables()
			})
		})
	})
})

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, totpsecret)
	VALUES (97881, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active',
			'EsO1rvIdhjNqAO5lLWreh/XBxvTfM7/1itvYdHwIw0V7HWuH77asgxEZJwdEBhaAVu5rSwbTDZZGLolC');
	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status)
	VALUES (97882, 88888, 'jane', 'doe', 'jadoe', '$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42', 'inactive');

	INSERT INTO roles (id, name) VALUES (98881, 'admin_acc_tests');
	INSERT INTO roles (id, name) VALUES (98882, 'client_acc_tests');
	INSERT INTO roles (id, name) VALUES (98883, 'tester_acc_tests');

	INSERT INTO user_role (id, userid, roleid) VALUES(99881, 97881, 98881);
	INSERT INTO user_role (id, userid, roleid) VALUES(99882, 97881, 98882);
	INSERT INTO user_role (id, userid, roleid) VALUES(99883, 97882, 98883);
	`
	helpers.Populate(q)
}

func cleanTables() {
	q := `
	DELETE FROM user_role where id in (99881, 99882, 99883);
	DELETE FROM roles where id in (98881, 98882, 98883);
	DELETE FROM users where tenantid = 88888;
	DELETE FROM tenants where id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
