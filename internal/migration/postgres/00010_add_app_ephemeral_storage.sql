-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE apps ADD COLUMN maxephemeralstorage TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN minephemeralstorage TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd