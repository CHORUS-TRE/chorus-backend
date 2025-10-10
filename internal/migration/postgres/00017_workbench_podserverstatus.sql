-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workbenchs ADD COLUMN serverpodstatus TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
