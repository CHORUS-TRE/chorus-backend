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
-- +migrate StatementBegin
WITH rd AS (
  SELECT id AS roleid FROM role_definitions WHERE name = 'WorkspaceAdmin'
),
src AS (
  SELECT w.userid, rd.roleid, w.id AS workspace_id,
         ROW_NUMBER() OVER (PARTITION BY w.userid ORDER BY w.id) AS rn
  FROM workspaces w
  CROSS JOIN rd
),
ins AS (
  INSERT INTO user_role (userid, roleid)
  SELECT userid, roleid
  FROM src
  RETURNING id, userid
),
num AS (
  SELECT i.id AS userroleid, i.userid,
         ROW_NUMBER() OVER (PARTITION BY i.userid ORDER BY i.id) AS rn
  FROM ins i
)
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT n.userroleid, 'workspace', s.workspace_id::text
FROM num n
JOIN src s
  ON s.userid = n.userid AND s.rn = n.rn
LEFT JOIN user_role_context urc
  ON urc.userroleid = n.userroleid AND urc.contextdimension = 'workspace'
WHERE urc.userroleid IS NULL;
-- +migrate StatementEnd

-- +migrate StatementEnd

-- +migrate StatementBegin
WITH rd AS (
  SELECT id AS roleid FROM role_definitions WHERE name = 'WorkbenchAdmin'
),
src AS (
  SELECT w.userid, rd.roleid, w.id AS workbench_id, w.workspaceid::text AS workspace_val
  FROM workbenchs w
  CROSS JOIN rd
),
need AS (
  SELECT s.*,
         ROW_NUMBER() OVER (PARTITION BY s.userid ORDER BY s.workbench_id) AS rn
  FROM src s
  WHERE NOT EXISTS (
    SELECT 1
    FROM user_role ur
    JOIN user_role_context c
      ON c.userroleid = ur.id
     AND c.contextdimension = 'workbench'
     AND c.value = s.workbench_id::text
    WHERE ur.userid = s.userid
      AND ur.roleid = s.roleid
  )
),
ins AS (
  INSERT INTO user_role (userid, roleid)
  SELECT userid, roleid
  FROM need
  RETURNING id, userid
),
ins_num AS (
  SELECT i.id AS userroleid, i.userid,
         ROW_NUMBER() OVER (PARTITION BY i.userid ORDER BY i.id) AS rn
  FROM ins i
),
pairs AS (
  SELECT n.userroleid, d.workbench_id::text AS workbench_val, d.workspace_val
  FROM ins_num n
  JOIN need d
    ON d.userid = n.userid AND d.rn = n.rn
)
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT p.userroleid, v.dim, v.val
FROM pairs p
CROSS JOIN LATERAL (VALUES
  ('workbench', p.workbench_val),
  ('workspace', p.workspace_val)
) AS v(dim, val)
ON CONFLICT (userroleid, contextdimension) DO NOTHING;
-- +migrate StatementEnd