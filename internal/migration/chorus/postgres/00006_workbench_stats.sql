-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workbenchs ADD COLUMN accessedat TIMESTAMP NULL;
ALTER TABLE workbenchs ADD COLUMN accessedcount BIGINT NOT NULL DEFAULT 0;
-- +migrate StatementEnd
