-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE apps ADD COLUMN iconbackgroundcolor TEXT NOT NULL DEFAULT 'transparent';
-- +migrate StatementEnd
