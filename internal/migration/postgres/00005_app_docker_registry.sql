-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE apps ADD COLUMN dockerimagerregistry TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
