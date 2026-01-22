//go:build acceptance

package approval_request_test

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/openapi"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	approval_request_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/approval-request/client"
	approval_request "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/approval-request/client/approval_request_service"
	approval_request_models "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/approval-request/models"
	workspace_file_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace-file/client"
	workspace_file "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace-file/client/workspace_file_service"
	workspace_file_models "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace-file/models"
	workspace_client "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/client"
	workspace "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/client/workspace_service"
	workspace_models "github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/models"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testTenantID    = 88888
	approverUserID  = 97881
	requesterUserID = 97882
)

var schemes = []string{"http"}

func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
	auth := httptransport.BearerToken(t)
	return func(co *runtime.ClientOperation) {
		co.AuthInfo = auth
	}
}

func WorkspaceServiceHTTPClient() *workspace_client.ChorusWorkspaceService {
	return workspace_client.New(openapi.NewNopCloserClientTransport(helpers.ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func WorkspaceFileServiceHTTPClient() *workspace_file_client.ChorusWorkspaceFileService {
	return workspace_file_client.New(openapi.NewNopCloserClientTransport(helpers.ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

func ApprovalRequestServiceHTTPClient() *approval_request_client.ChorusApprovalRequestService {
	return approval_request_client.New(openapi.NewNopCloserClientTransport(helpers.ComponentURL(), "", schemes, logger.TechLog), strfmt.Default)
}

var _ = Describe("approval request service", func() {
	helpers.Setup()

	Describe("data extraction approval workflow", func() {

		Given("an approver user and a requester user", func() {

			When("requester creates a workspace, uploads a file, and creates a data extraction request", func() {

				var workspaceID string
				var approvalRequestID string

				BeforeEach(func() {
					setupTables()
				})

				AfterEach(func() {
					cleanTables()
				})

				It("should allow the approver to approve the request", func() {
					approverAuth := getAuthAsClientOpts(helpers.CreateJWTToken(
						approverUserID, testTenantID,
						authorization.RoleSuperAdmin.String(),
						map[string]string{"user": fmt.Sprintf("%d", approverUserID)},
					))

					requesterAuth := getAuthAsClientOpts(helpers.CreateJWTToken(
						requesterUserID, testTenantID,
						authorization.RoleAuthenticated.String(),
						map[string]string{"user": fmt.Sprintf("%d", requesterUserID)},
					))

					By("Step 1: Creating a workspace as the requester")
					wsClient := WorkspaceServiceHTTPClient()
					createWsReq := workspace.NewWorkspaceServiceCreateWorkspaceParams().WithBody(
						&workspace_models.ChorusWorkspace{
							Name:        "Test Workspace for Approval",
							ShortName:   "test-ws",
							Description: "A workspace for testing approval requests",
						},
					)

					createWsResp, err := wsClient.WorkspaceService.WorkspaceServiceCreateWorkspace(createWsReq, requesterAuth)
					ExpectAPIErr(err).Should(BeNil())
					Expect(createWsResp.Payload.Result.Workspace).ShouldNot(BeNil())
					workspaceID = createWsResp.Payload.Result.Workspace.ID
					Expect(workspaceID).ShouldNot(BeEmpty())

					By("Step 2: Uploading a file to the workspace")
					fileClient := WorkspaceFileServiceHTTPClient()
					testFileContent := "This is test file content for data extraction"
					testFilePath := "test-data/sample.txt"
					encodedContent := base64.StdEncoding.EncodeToString([]byte(testFileContent))

					createFileReq := workspace_file.NewWorkspaceFileServiceCreateWorkspaceFileParams().
						WithWorkspaceID(workspaceID).
						WithFile(&workspace_file_models.ChorusWorkspaceFile{
							Path:    testFilePath,
							Name:    "sample.txt",
							Content: strfmt.Base64(encodedContent),
						})

					createFileResp, err := fileClient.WorkspaceFileService.WorkspaceFileServiceCreateWorkspaceFile(createFileReq, requesterAuth)
					ExpectAPIErr(err).Should(BeNil())
					Expect(createFileResp.Payload.Result.File).ShouldNot(BeNil())

					By("Step 3: Creating a data extraction request as the requester")
					approvalClient := ApprovalRequestServiceHTTPClient()
					createApprovalReq := approval_request.NewApprovalRequestServiceCreateDataExtractionRequestParams().
						WithBody(&approval_request_models.ChorusCreateDataExtractionRequestRequest{
							SourceWorkspaceID: workspaceID,
							Title:             "Test Data Extraction Request",
							Description:       "Request to extract sample.txt for analysis",
							FilePaths:         []string{testFilePath},
							ApproverIds:       []string{fmt.Sprintf("%d", approverUserID)},
						})

					createApprovalResp, err := approvalClient.ApprovalRequestService.ApprovalRequestServiceCreateDataExtractionRequest(createApprovalReq, requesterAuth)
					ExpectAPIErr(err).Should(BeNil())
					Expect(createApprovalResp.Payload.Result.ApprovalRequest).ShouldNot(BeNil())
					approvalRequestID = createApprovalResp.Payload.Result.ApprovalRequest.ID
					Expect(approvalRequestID).ShouldNot(BeEmpty())
					Expect(*createApprovalResp.Payload.Result.ApprovalRequest.Status).Should(Equal(approval_request_models.ChorusApprovalRequestStatusAPPROVALREQUESTSTATUSPENDING))

					By("Step 4: Approving the request as the approver")
					approveReq := approval_request.NewApprovalRequestServiceApproveApprovalRequestParams().
						WithID(approvalRequestID).
						WithBody(&approval_request_models.ApprovalRequestServiceApproveApprovalRequestBody{
							Approve: true,
							Comment: "Approved for data extraction",
						})

					approveResp, err := approvalClient.ApprovalRequestService.ApprovalRequestServiceApproveApprovalRequest(approveReq, approverAuth)
					ExpectAPIErr(err).Should(BeNil())
					Expect(approveResp.Payload.Result.ApprovalRequest).ShouldNot(BeNil())
					Expect(*approveResp.Payload.Result.ApprovalRequest.Status).Should(Equal(approval_request_models.ChorusApprovalRequestStatusAPPROVALREQUESTSTATUSAPPROVED))

					By("Step 5: Verifying the approval request status")
					getReq := approval_request.NewApprovalRequestServiceGetApprovalRequestParams().WithID(approvalRequestID)
					getResp, err := approvalClient.ApprovalRequestService.ApprovalRequestServiceGetApprovalRequest(getReq, approverAuth)
					ExpectAPIErr(err).Should(BeNil())
					Expect(getResp.Payload.Result.ApprovalRequest).ShouldNot(BeNil())
					Expect(*getResp.Payload.Result.ApprovalRequest.Status).Should(Equal(approval_request_models.ChorusApprovalRequestStatusAPPROVALREQUESTSTATUSAPPROVED))
					Expect(getResp.Payload.Result.ApprovalRequest.ApprovedByID).Should(Equal(fmt.Sprintf("%d", approverUserID)))
				})
			})
		})
	})
})

func setupTables() {
	cleanTables()

	q := `
	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, totpsecret)
	VALUES (97881, 88888, 'approver', 'user', 'approver', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active',
			'EsO1rvIdhjNqAO5lLWreh/XBxvTfM7/1itvYdHwIw0V7HWuH77asgxEZJwdEBhaAVu5rSwbTDZZGLolC');
	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, totpsecret)
	VALUES (97882, 88888, 'requester', 'user', 'requester', '$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42', 'active',
			'EsO1rvIdhjNqAO5lLWreh/XBxvTfM7/1itvYdHwIw0V7HWuH77asgxEZJwdEBhaAVu5rSwbTDZZGLolC');

	INSERT INTO role_definitions (id, name) VALUES (98881, 'SuperAdmin');
	INSERT INTO role_definitions (id, name) VALUES (98882, 'Public');
	INSERT INTO role_definitions (id, name) VALUES (98883, 'Authenticated');

	INSERT INTO user_role (id, userid, roleid) VALUES(99881, 97881, 98881);
	INSERT INTO user_role (id, userid, roleid) VALUES(99882, 97881, 98882);
	INSERT INTO user_role (id, userid, roleid) VALUES(99883, 97882, 98883);
	`
	helpers.Populate(q)

	helpers.Dump("SELECT * FROM users WHERE tenantid = 88888")
}

func cleanTables() {
	q := `
	DELETE FROM approval_request_files WHERE approval_request_id IN (SELECT id FROM approval_requests WHERE tenant_id = 88888);
	DELETE FROM approval_request_approvers WHERE approval_request_id IN (SELECT id FROM approval_requests WHERE tenant_id = 88888);
	DELETE FROM approval_requests WHERE tenant_id = 88888;
	DELETE FROM workspace_files WHERE workspace_id IN (SELECT id FROM workspaces WHERE tenantid = 88888);
	DELETE FROM workspaces WHERE tenantid = 88888;
	DELETE FROM user_role WHERE id IN (99881, 99882, 99883);
	DELETE FROM role_definitions WHERE id IN (98881, 98882, 98883);
	DELETE FROM users WHERE tenantid = 88888;
	DELETE FROM tenants WHERE id = 88888;
	`
	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
		panic(err.Error())
	}
}
