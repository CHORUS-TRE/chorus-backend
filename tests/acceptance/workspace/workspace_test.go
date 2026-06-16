//go:build acceptance

package workspace_test

import (
	"context"
	"fmt"
	"net/http"

	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	workspace_svc "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/client/workspace_service"
	workspace_models "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = httptransport.BearerToken(t)
	}
}

var _ = Describe("workspace service", func() {
	helpers.Setup()

	Describe("get workspace", func() {

		Given("no jwt-token", func() {
			When("GET /api/rest/v1/workspaces/{id} is called", func() {
				setupTables()
				req := workspace_svc.NewWorkspaceServiceGetWorkspaceParams().WithID("80001")
				c := helpers.WorkspaceServiceHTTPClient()
				_, err := c.WorkspaceService.WorkspaceServiceGetWorkspace(req, getAuthAsClientOpts("invalid"))

				Then("an authentication error is returned", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))
				})
				cleanTables()
			})
		})

		Given("a valid jwt-token with no access to the workspace", func() {
			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90001, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90001"}))

			When("GET /api/rest/v1/workspaces/{id} is called", func() {
				setupTables()
				req := workspace_svc.NewWorkspaceServiceGetWorkspaceParams().WithID("80001")
				c := helpers.WorkspaceServiceHTTPClient()
				_, err := c.WorkspaceService.WorkspaceServiceGetWorkspace(req, auth)

				Then("a permission error is returned", func() {
					ExpectAPIErr(err).ShouldNot(BeNil())
					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusForbidden)))
				})
				cleanTables()
			})
		})

		Given("a valid jwt-token with WorkspaceAdmin role on the workspace", func() {
			// Token fails because role is in "roles" field instead of "r"
			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleWorkspaceAdmin.String(), map[string]string{"workspace": "80001"}))

			When("GET /api/rest/v1/workspaces/{id} is called", func() {
				setupTables()
				req := workspace_svc.NewWorkspaceServiceGetWorkspaceParams().WithID("80001")
				c := helpers.WorkspaceServiceHTTPClient()
				resp, err := c.WorkspaceService.WorkspaceServiceGetWorkspace(req, auth)

				Then("the workspace is returned with correct fields", func() {
					ExpectAPIErr(err).Should(BeNil())
					ws := resp.Payload.Result.Workspace
					Expect(ws.Name).Should(Equal("Private WS"))
					Expect(ws.ShortName).Should(Equal("private-ws"))
					Expect(ws.Description).Should(Equal("A private workspace"))
					Expect(*ws.Visibility).Should(Equal(workspace_models.ChorusWorkspaceVisibilityWORKSPACEVISIBILITYPRIVATE))
				})
				cleanTables()
			})
		})
	})
})

// ---------------------------------------------------------------------------
// DB helpers
// ---------------------------------------------------------------------------

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id, tenantid, firstname, lastname, username, email, password, status, createdat, updatedat)
	VALUES (90000, 88888, 'alice', 'admin', 'aadmin', 'aadmin@example.com', 'x', 'active', NOW(), NOW());

	INSERT INTO users (id, tenantid, firstname, lastname, username, email, password, status, createdat, updatedat)
	VALUES (90001, 88888, 'bob', 'user', 'buser', 'buser@example.com', 'x', 'active', NOW(), NOW());

	INSERT INTO workspaces (id, tenantid, userid, name, shortname, description, status, visibility, createdat, updatedat)
	VALUES (80001, 88888, 90000, 'Private WS', 'private-ws', 'A private workspace', 'active', 'private', NOW(), NOW());

	INSERT INTO workspaces (id, tenantid, userid, name, shortname, description, status, visibility, createdat, updatedat)
	VALUES (80002, 88888, 90000, 'Public WS', 'public-ws', 'A public workspace', 'active', 'public', NOW(), NOW());

	INSERT INTO user_role (id, userid, roleid) VALUES (92001, 90000, (SELECT id FROM role_definitions WHERE name = 'Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES (92002, 90001, (SELECT id FROM role_definitions WHERE name = 'Authenticated'));
	INSERT INTO user_role (id, userid, roleid) VALUES (92003, 90000, (SELECT id FROM role_definitions WHERE name = 'WorkspaceAdmin'));
	INSERT INTO user_role (id, userid, roleid) VALUES (92004, 90000, (SELECT id FROM role_definitions WHERE name = 'WorkspaceAdmin'));

	INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES (92001, 'user', 90000);
	INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES (92002, 'user', 90001);
	INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES (92003, 'workspace', 80001);
	INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES (92004, 'workspace', 80002);
	`
	helpers.Populate(q)
}

func cleanTables() {
	q := `
	DELETE FROM user_role_context WHERE userroleid IN (92001, 92002, 92003, 92004);
	DELETE FROM user_role WHERE id IN (92001, 92002, 92003, 92004);
	DELETE FROM workspaces WHERE tenantid = 88888;
	DELETE FROM users WHERE tenantid = 88888;
	DELETE FROM tenants WHERE id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
