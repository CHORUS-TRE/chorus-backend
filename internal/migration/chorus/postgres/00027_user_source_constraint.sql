-- +migrate Up

-- +migrate StatementBegin
UPDATE users
    SET (source, status, username, updatedat) = ('internal', 'deleted', username || '-' || gen_random_uuid()::text, NOW())
    WHERE source = '';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE users ADD CONSTRAINT users_source_not_empty CHECK (source <> '');
-- +migrate StatementEnd
