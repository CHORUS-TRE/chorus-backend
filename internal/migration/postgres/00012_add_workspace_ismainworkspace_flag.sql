-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workspaces ADD COLUMN ismain BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX unique_main_workspace_per_user
ON workspaces(userid)
WHERE deletedat IS NULL AND ismain = true; 
-- +migrate StatementEnd