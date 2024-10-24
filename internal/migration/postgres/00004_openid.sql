-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE users ADD COLUMN source TEXT NOT NULL DEFAULT 'internal';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE users DROP CONSTRAINT users_username_key;
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE users ADD CONSTRAINT users_username_source_unique UNIQUE (username, source);
-- +migrate StatementEnd