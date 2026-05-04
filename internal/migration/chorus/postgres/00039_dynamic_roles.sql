-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE role_definitions
ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'system',
ADD COLUMN IF NOT EXISTS dynamic BOOLEAN NOT NULL DEFAULT false;
-- +migrate StatementEnd

-- +migrate StatementBegin
UPDATE role_definitions SET dynamic = false WHERE dynamic IS NULL;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS dynamic_role_permissions (
    roledefinitionid BIGINT NOT NULL,
    permissionname TEXT NOT NULL,
    PRIMARY KEY (roledefinitionid, permissionname),
    FOREIGN KEY (roledefinitionid) REFERENCES role_definitions(id) ON DELETE CASCADE
);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE INDEX IF NOT EXISTS dynamic_role_permissions_permissionname_idx
ON dynamic_role_permissions(permissionname);
-- +migrate StatementEnd