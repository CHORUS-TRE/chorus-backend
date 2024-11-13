-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE apps ADD COLUMN dockerimageregistry TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
