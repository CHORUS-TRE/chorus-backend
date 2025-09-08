-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE roles RENAME TO role_definitions;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE role_definitions_required_contexts (
    roledefinitionid BIGINT NOT NULL,
    contextdimension TEXT NOT NULL,
    contextquantifier TEXT NOT NULL CHECK (contextquantifier IN ('x', '*')),
    PRIMARY KEY (roledefinitionid, contextdimension),
    FOREIGN KEY (roledefinitionid) REFERENCES role_definitions(id) ON DELETE CASCADE
);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE user_role_context (
    userroleid BIGINT NOT NULL,
    contextdimension TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (userroleid, contextdimension),
    FOREIGN KEY (userroleid) REFERENCES user_role(id) ON DELETE CASCADE
);
-- +migrate StatementEnd

-- +migrate StatementBegin
UPDATE role_definitions SET name = 'SuperAdmin' WHERE name = 'admin';
-- +migrate StatementEnd

-- +migrate StatementBegin
UPDATE role_definitions SET name = 'Authenticated' WHERE name = 'authenticated';
-- +migrate StatementEnd

-- +migrate StatementBegin
DELETE FROM user_role where roleid = (SELECT id FROM role_definitions WHERE name = 'chorus');
-- +migrate StatementEnd

-- +migrate StatementBegin
DELETE FROM role_definitions where name = 'chorus';
-- +migrate StatementEnd

--  RolePublic               RoleName = "Public"
-- 	RoleAuthenticated        RoleName = "Authenticated"
-- 	RoleWorkspaceGuest       RoleName = "WorkspaceGuest"
-- 	RoleWorkspaceMember      RoleName = "WorkspaceMember"
-- 	RoleWorkspaceMaintainer  RoleName = "WorkspaceMaintainer"
-- 	RoleWorkspaceAdmin       RoleName = "WorkspaceAdmin"
-- 	RoleWorkbenchViewer      RoleName = "WorkbenchViewer"
-- 	RoleWorkbenchMember      RoleName = "WorkbenchMember"
-- 	RoleWorkbenchAdmin       RoleName = "WorkbenchAdmin"
-- 	RoleHealthchecker        RoleName = "Healthchecker"
-- 	RolePlateformUserManager RoleName = "PlateformUserManager"
-- 	RoleAppStoreAdmin        RoleName = "AppStoreAdmin"
-- 	RoleSuperAdmin           RoleName = "SuperAdmin"

-- +migrate StatementBegin
INSERT INTO role_definitions (name) VALUES
('WorkspaceGuest'),
('WorkspaceMember'),
('WorkspaceMaintainer'),
('WorkspaceAdmin'),
('WorkbenchViewer'),
('WorkbenchMember'),
('WorkbenchAdmin'),
('Healthchecker'),
('PlateformUserManager'),
('AppStoreAdmin');
-- +migrate StatementEnd

-- +migrate StatementBegin
INSERT INTO user_role (userid, roleid)
SELECT userid, (SELECT id FROM role_definitions WHERE name = 'WorkspaceAdmin') FROM workspaces;
-- +migrate StatementEnd

-- +migrate StatementBegin
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT ur.id, 'workspace', w.id::text
FROM workspaces w
JOIN user_role ur ON ur.userid = w.userid AND ur.roleid = (SELECT id FROM role_definitions WHERE name = 'WorkspaceAdmin')
LEFT JOIN user_role_context urc
  ON urc.userroleid = ur.id AND urc.contextdimension = 'workspace'
WHERE urc.userroleid IS NULL;
-- +migrate StatementEnd

-- +migrate StatementBegin
INSERT INTO user_role (userid, roleid)
SELECT userid, (SELECT id FROM role_definitions WHERE name = 'WorkbenchAdmin') FROM workbenchs;
-- +migrate StatementEnd

-- +migrate StatementBegin
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT ur.id, 'workspace', w.workspaceid::text
FROM workbenchs w
JOIN user_role ur ON ur.userid = w.userid AND ur.roleid = (SELECT id FROM role_definitions WHERE name = 'WorkbenchAdmin')
LEFT JOIN user_role_context urc
  ON urc.userroleid = ur.id AND urc.contextdimension = 'workspace'
WHERE urc.userroleid IS NULL;
-- +migrate StatementEnd

-- +migrate StatementBegin
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT ur.id, 'workbench', w.id::text
FROM workbenchs w
JOIN user_role ur ON ur.userid = w.userid AND ur.roleid = (SELECT id FROM role_definitions WHERE name = 'WorkbenchAdmin')
LEFT JOIN user_role_context urc
  ON urc.userroleid = ur.id AND urc.contextdimension = 'workbench'
WHERE urc.userroleid IS NULL;
-- +migrate StatementEnd