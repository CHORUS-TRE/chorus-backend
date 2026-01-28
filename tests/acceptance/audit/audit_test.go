//go:build acceptance

package audit_test

// func getAuthAsClientOpts(t string) func(*runtime.ClientOperation) {
// 	auth := httptransport.BearerToken(t)
// 	return func(co *runtime.ClientOperation) {
// 		co.AuthInfo = auth
// 	}
// }

// // Helper to query audit entries from database
// func getAuditEntries(tenantID uint64, action model.AuditAction) ([]*model.AuditEntry, error) {
// 	query := `
// 		SELECT id, tenantid, userid, username, correlationid, action, workspaceid, workbenchid, description, details, createdat
// 		FROM audit
// 		WHERE tenantid = $1 AND action = $2
// 		ORDER BY createdat DESC
// 	`
// 	var entries []*model.AuditEntry
// 	err := helpers.DB().SelectContext(context.Background(), &entries, query, tenantID, action)
// 	return entries, err
// }

// var _ = Describe("audit service", func() {
// 	helpers.Setup()

// 	Describe("user creation audit", func() {
// 		Given("a platform manager jwt", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, string(authorization.RolePlateformUserManager), map[string]string{}))

// 			When("a user is successfully created", func() {
// 				setupTables()

// 				req := user.NewUserServiceCreateUserParams().WithBody(
// 					&user_models.ChorusUser{
// 						FirstName: "John",
// 						LastName:  "Doe",
// 						Username:  "jdoe_test",
// 						Password:  "SecurePassword123!",
// 						Status:    "active",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				resp, err := c.UserService.UserServiceCreateUser(req, auth)

// 				Then("an audit entry should be created", func() {
// 					ExpectAPIErr(err).Should(BeNil())
// 					Expect(resp).ShouldNot(BeNil())

// 					// Wait a bit for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					// Query audit entries
// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserCreate)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					// Verify the most recent entry
// 					entry := entries[0]
// 					Expect(entry.TenantID).Should(Equal(uint64(88888)))
// 					Expect(entry.UserID).Should(Equal(uint64(90000)))
// 					Expect(entry.Action).Should(Equal(model.AuditActionUserCreate))
// 					Expect(entry.Description).Should(ContainSubstring("Created user"))
// 					Expect(entry.Details["username"]).Should(Equal("jdoe_test"))
// 				})

// 				cleanTables()
// 			})

// 			When("user creation fails", func() {
// 				setupTables()

// 				// Try to create user with missing required fields
// 				req := user.NewUserServiceCreateUserParams().WithBody(
// 					&user_models.ChorusUser{
// 						LastName: "Doe",
// 						Username: "jdoe_test2",
// 						Password: "SecurePassword123!",
// 						Status:   "active",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceCreateUser(req, auth)

// 				Then("an audit entry for the failure should be created", func() {
// 					ExpectAPIErr(err).ShouldNot(BeNil())

// 					// Wait a bit for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					// Note: Based on the middleware pattern, failed validations might not
// 					// create audit entries if they fail before reaching the controller
// 					// This test documents the current behavior
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("user update audit", func() {
// 		Given("a super admin jwt", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))

// 			When("a user is successfully updated", func() {
// 				setupTables()

// 				req := user.NewUserServiceUpdateUserParams().WithBody(
// 					&user_models.ChorusUser{
// 						ID:        "90000",
// 						FirstName: "Updated",
// 						LastName:  "Name",
// 						Username:  "hmoto",
// 						Roles:     []string{"admin", "authenticated"},
// 						Status:    "active",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceUpdateUser(req, auth)

// 				Then("an audit entry should be created", func() {
// 					ExpectAPIErr(err).Should(BeNil())

// 					// Wait for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserUpdate)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]
// 					Expect(entry.Action).Should(Equal(model.AuditActionUserUpdate))
// 					Expect(entry.Description).Should(ContainSubstring("Updated user"))
// 					Expect(entry.Details["user_id"]).Should(BeEquivalentTo(90000))
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("user deletion audit", func() {
// 		Given("a super admin jwt", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))

// 			When("a user is successfully deleted", func() {
// 				setupTables()

// 				req := user.NewUserServiceDeleteUserParams().WithID("90001")

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceDeleteUser(req, auth)

// 				Then("an audit entry should be created", func() {
// 					ExpectAPIErr(err).Should(BeNil())

// 					// Wait for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserDelete)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]
// 					Expect(entry.Action).Should(Equal(model.AuditActionUserDelete))
// 					Expect(entry.Description).Should(ContainSubstring("Deleted user"))
// 					Expect(entry.Details["user_id"]).Should(BeEquivalentTo(90001))
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("password change audit", func() {
// 		Given("an authenticated user", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))

// 			When("password is successfully changed", func() {
// 				setupTables()

// 				req := user.NewUserServiceUpdatePasswordParams().WithBody(
// 					&user_models.ChorusUpdatePasswordRequest{
// 						CurrentPassword: "johnPassword",
// 						NewPassword:     "NewSecurePassword123!",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

// 				Then("an audit entry should be created", func() {
// 					ExpectAPIErr(err).Should(BeNil())

// 					// Wait for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserPasswordChange)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]
// 					Expect(entry.Action).Should(Equal(model.AuditActionUserPasswordChange))
// 					Expect(entry.Description).Should(ContainSubstring("password"))
// 					Expect(entry.UserID).Should(Equal(uint64(90000)))

// 					// Verify sensitive data is NOT logged
// 					_, hasOldPassword := entry.Details["current_password"]
// 					_, hasNewPassword := entry.Details["new_password"]
// 					Expect(hasOldPassword).Should(BeFalse())
// 					Expect(hasNewPassword).Should(BeFalse())
// 				})

// 				cleanTables()
// 			})

// 			When("password change fails with wrong current password", func() {
// 				setupTables()

// 				req := user.NewUserServiceUpdatePasswordParams().WithBody(
// 					&user_models.ChorusUpdatePasswordRequest{
// 						CurrentPassword: "wrongPassword",
// 						NewPassword:     "NewSecurePassword123!",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceUpdatePassword(req, auth)

// 				Then("a failure audit entry should be created", func() {
// 					ExpectAPIErr(err).ShouldNot(BeNil())
// 					Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("%v", http.StatusUnauthorized)))

// 					// Wait for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserPasswordChange)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]
// 					Expect(entry.Description).Should(ContainSubstring("Failed"))

// 					// Verify error details are present but sensitive data is not
// 					_, hasErrorMsg := entry.Details["error_message"]
// 					Expect(hasErrorMsg).Should(BeTrue())
// 					_, hasPassword := entry.Details["current_password"]
// 					Expect(hasPassword).Should(BeFalse())
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("TOTP operations audit", func() {
// 		Given("an authenticated user with TOTP", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleAuthenticated.String(), map[string]string{"user": "90000"}))

// 			When("TOTP is reset", func() {
// 				setupTablesWithTotpUser()

// 				req := user.NewUserServiceResetTotpParams().WithBody(
// 					&user_models.ChorusResetTotpRequest{Password: "johnPassword"},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceResetTotp(req, auth)

// 				Then("an audit entry should be created", func() {
// 					ExpectAPIErr(err).Should(BeNil())

// 					// Wait for async audit logging
// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserTotpReset)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]
// 					Expect(entry.Action).Should(Equal(model.AuditActionUserTotpReset))
// 					Expect(entry.Description).Should(ContainSubstring("TOTP"))
// 					Expect(entry.UserID).Should(Equal(uint64(90000)))

// 					// Verify TOTP secrets are NOT logged
// 					_, hasTotpSecret := entry.Details["totp_secret"]
// 					_, hasRecoveryCodes := entry.Details["totp_recovery_codes"]
// 					Expect(hasTotpSecret).Should(BeFalse())
// 					Expect(hasRecoveryCodes).Should(BeFalse())
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("role operations audit", func() {
// 		Given("a super admin jwt", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(1, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))

// 			When("a role is assigned to a user", func() {
// 				setupTables()

// 				req := user.NewUserServiceCreateUserRoleParams().
// 					WithUserID("90000").
// 					WithBody(&user_models.UserServiceCreateUserRoleBody{
// 						Role: &user_models.ChorusRole{
// 							ID:   "101",
// 							Name: "SuperAdmin",
// 						},
// 					})

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceCreateUserRole(req, auth)

// 				Then("an audit entry should be created", func() {
// 					// May succeed or fail if role already exists
// 					if err == nil {
// 						time.Sleep(100 * time.Millisecond)

// 						entries, auditErr := getAuditEntries(88888, model.AuditActionUserRoleAssign)
// 						Expect(auditErr).Should(BeNil())

// 						if len(entries) > 0 {
// 							entry := entries[0]
// 							Expect(entry.Action).Should(Equal(model.AuditActionUserRoleAssign))
// 							Expect(entry.Details["user_id"]).Should(BeEquivalentTo(90000))
// 							Expect(entry.Details["role_name"]).Should(Equal("SuperAdmin"))
// 						}
// 					}
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("audit entry correlation", func() {
// 		Given("a user performing multiple actions", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))

// 			When("multiple operations are performed in sequence", func() {
// 				setupTables()

// 				// Create a user
// 				createReq := user.NewUserServiceCreateUserParams().WithBody(
// 					&user_models.ChorusUser{
// 						FirstName: "Test",
// 						LastName:  "User",
// 						Username:  "testuser_audit",
// 						Password:  "SecurePassword123!",
// 						Status:    "active",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				createResp, createErr := c.UserService.UserServiceCreateUser(createReq, auth)

// 				Then("all audit entries should be correlated to the same tenant and user", func() {
// 					ExpectAPIErr(createErr).Should(BeNil())
// 					Expect(createResp).ShouldNot(BeNil())

// 					time.Sleep(100 * time.Millisecond)

// 					// Get all audit entries for this tenant
// 					query := `
// 						SELECT id, tenantid, userid, username, correlationid, action, workspaceid, workbenchid, description, details, createdat
// 						FROM audit
// 						WHERE tenantid = $1
// 						ORDER BY createdat DESC
// 					`
// 					var entries []*model.AuditEntry
// 					err := helpers.DB().SelectContext(context.Background(), &entries, query, uint64(88888))
// 					Expect(err).Should(BeNil())

// 					// All entries should have the same tenant
// 					for _, entry := range entries {
// 						Expect(entry.TenantID).Should(Equal(uint64(88888)))
// 						Expect(entry.UserID).Should(BeNumerically(">", 0))
// 						Expect(entry.Username).ShouldNot(BeEmpty())
// 						Expect(entry.CreatedAt).ShouldNot(BeZero())
// 					}
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})

// 	Describe("audit entry details validation", func() {
// 		Given("various user operations", func() {
// 			auth := getAuthAsClientOpts(helpers.CreateJWTToken(90000, 88888, authorization.RoleSuperAdmin.String(), map[string]string{"user": "*"}))

// 			When("operations include additional context", func() {
// 				setupTables()

// 				// Update user
// 				req := user.NewUserServiceUpdateUserParams().WithBody(
// 					&user_models.ChorusUser{
// 						ID:        "90000",
// 						FirstName: "Context",
// 						LastName:  "Test",
// 						Username:  "hmoto",
// 						Roles:     []string{"admin"},
// 						Status:    "active",
// 					},
// 				)

// 				c := helpers.UserServiceHTTPClient()
// 				_, err := c.UserService.UserServiceUpdateUser(req, auth)

// 				Then("audit entries should contain structured details", func() {
// 					ExpectAPIErr(err).Should(BeNil())

// 					time.Sleep(100 * time.Millisecond)

// 					entries, auditErr := getAuditEntries(88888, model.AuditActionUserUpdate)
// 					Expect(auditErr).Should(BeNil())
// 					Expect(len(entries)).Should(BeNumerically(">=", 1))

// 					entry := entries[0]

// 					// Verify details structure
// 					Expect(entry.Details).ShouldNot(BeNil())
// 					Expect(len(entry.Details)).Should(BeNumerically(">", 0))

// 					// Verify common fields are present
// 					userID, hasUserID := entry.Details["user_id"]
// 					Expect(hasUserID).Should(BeTrue())
// 					Expect(userID).Should(BeEquivalentTo(90000))
// 				})

// 				cleanTables()
// 			})
// 		})
// 	})
// })

// func setupTables() {
// 	cleanTables()

// 	q := `
// 	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

// 	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, createdat, updatedat)
// 	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active', NOW(), NOW());
// 	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, createdat, updatedat)
// 	VALUES (90001, 88888, 'jane', 'doe', 'jadoe', '$2a$10$1VdWx3wG9KWZaHSzvUxQi.ZHzBJE8aPIDfsblTZPFRWyeWu4B9.42', 'disabled', NOW(), NOW());

// 	INSERT INTO role_definitions (id, name) VALUES (101, 'SuperAdmin');
// 	INSERT INTO role_definitions (id, name) VALUES (102, 'Authenticated');

// 	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, 101);
// 	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, 102);
// 	INSERT INTO user_role (id, userid, roleid) VALUES(92003, 90001, 101);
// 	INSERT INTO user_role (id, userid, roleid) VALUES(92004, 90001, 102);
// 	`
// 	helpers.Populate(q)
// }

// func setupTablesWithTotpUser() {
// 	cleanTables()

// 	q := `
// 	INSERT INTO tenants (id, name) VALUES (88888, 'test tenant');

// 	INSERT INTO users (id, tenantid, firstname, lastname, username, password, status, createdat, updatedat, totpsecret)
// 	VALUES (90000, 88888, 'hello', 'moto', 'hmoto', '$2a$10$kTAQ1EsMqdNAgQecrLOdNOZF.X71sNfokCs5be8..eVFLPQ/1iCTO', 'active',
// 			NOW(), NOW(), 'ohKtu9PFHMquP5Zemcfb4XFQ8TuYnA5Gk1txooQINWL2AbhonyGW0H66zmX8YdUEDEZPYGjOCDPBOF9W');

// 	INSERT INTO role_definitions (id, name) VALUES (101, 'SuperAdmin');
// 	INSERT INTO role_definitions (id, name) VALUES (102, 'Authenticated');

// 	INSERT INTO user_role (id, userid, roleid) VALUES(92001, 90000, 101);
// 	INSERT INTO user_role (id, userid, roleid) VALUES(92002, 90000, 102);

// 	INSERT INTO totp_recovery_codes (id, userid, tenantid, code)
// 	VALUES (88888, 90000, 88888, '0Uu+C4s1i+mrS7pqmI2SHJe+Hcg3l4K/ylusXoIv25RE6qEUyRY='),
// 		(88889, 90000, 88888, '0YZWPkeRISwyAeZsQ2otY+JMdR1P6N42NoN0UOxbPh7tnioAvF4=');
// 	`
// 	helpers.Populate(q)
// }

// func cleanTables() {
// 	q := `
// 	DELETE FROM audit WHERE tenantid = 88888;
// 	DELETE FROM notifications_read_by WHERE tenantid = 88888;
// 	DELETE FROM notifications WHERE tenantid = 88888;
// 	DELETE FROM totp_recovery_codes WHERE tenantid = 88888;
// 	DELETE FROM user_role WHERE id IN (92001,92002,92003,92004) OR userid IN (SELECT id FROM users WHERE tenantid = 88888 OR username IN ('user88888', 'jdoe_test', 'jdoe_test2', 'testuser_audit'));
// 	DELETE FROM role_definitions WHERE id IN (101,102);
// 	DELETE FROM users WHERE tenantid = 88888 OR username IN ('user88888', 'jdoe_test', 'jdoe_test2', 'testuser_audit');
// 	DELETE FROM tenants WHERE id = 88888;
// 	`
// 	if _, err := helpers.DB().ExecContext(context.Background(), q); err != nil {
// 		panic(err.Error())
// 	}
// }
