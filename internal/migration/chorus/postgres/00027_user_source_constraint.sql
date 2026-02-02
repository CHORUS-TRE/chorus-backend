-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE users ADD CONSTRAINT users_source_not_empty CHECK (source <> '');
-- +migrate StatementEnd

-- +migrate Down

-- +migrate StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_source_not_empty;
-- +migrate StatementEnd