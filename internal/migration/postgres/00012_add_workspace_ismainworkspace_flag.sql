-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workspaces ADD COLUMN ismain BOOLEAN NOT NULL DEFAULT FALSE;
-- +migrate StatementEnd