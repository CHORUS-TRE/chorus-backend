-- +migrate Up

-- Add new role definitions
-- +migrate StatementBegin
INSERT INTO role_definitions (name) VALUES
('DataManager')
ON CONFLICT (name) DO NOTHING;
-- +migrate StatementEnd

-- Rename WorkspacePI to WorkspaceDataManager in existing user roles
-- +migrate StatementBegin
UPDATE role_definitions
SET name = 'WorkspaceDataManager'
WHERE name = 'WorkspacePI';
-- +migrate StatementEnd

-- Add WorkspaceDataManager role to all users who have WorkspaceAdmin role
-- This ensures existing workspace admins can still download files
-- +migrate StatementBegin
WITH workspace_admin_roles AS (
    SELECT ur.userid, urc.value AS workspace_id
    FROM user_role ur
    JOIN role_definitions rd ON ur.roleid = rd.id
    JOIN user_role_context urc ON ur.id = urc.userroleid AND urc.contextdimension = 'workspace'
    WHERE rd.name = 'WorkspaceAdmin'
),
data_manager_role AS (
    SELECT id AS roleid FROM role_definitions WHERE name = 'WorkspaceDataManager'
),
existing_data_manager_roles AS (
    SELECT ur.userid, urc.value AS workspace_id
    FROM user_role ur
    JOIN role_definitions rd ON ur.roleid = rd.id
    JOIN user_role_context urc ON ur.id = urc.userroleid AND urc.contextdimension = 'workspace'
    WHERE rd.name = 'WorkspaceDataManager'
),
roles_to_add AS (
    SELECT war.userid, dmr.roleid, war.workspace_id,
           ROW_NUMBER() OVER (PARTITION BY war.userid ORDER BY war.workspace_id) AS rn
    FROM workspace_admin_roles war
    CROSS JOIN data_manager_role dmr
    LEFT JOIN existing_data_manager_roles edm 
        ON war.userid = edm.userid AND war.workspace_id = edm.workspace_id
    WHERE edm.userid IS NULL
),
ins AS (
    INSERT INTO user_role (userid, roleid)
    SELECT userid, roleid
    FROM roles_to_add
    RETURNING id, userid
),
num AS (
    SELECT i.id AS userroleid, i.userid,
           ROW_NUMBER() OVER (PARTITION BY i.userid ORDER BY i.id) AS rn
    FROM ins i
)
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT n.userroleid, 'workspace', rta.workspace_id::TEXT
FROM num n
JOIN roles_to_add rta
    ON rta.userid = n.userid AND rta.rn = n.rn;
-- +migrate StatementEnd
