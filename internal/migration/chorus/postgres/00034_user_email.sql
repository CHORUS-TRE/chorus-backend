-- +migrate Up
ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT '';
UPDATE users SET email = username WHERE username ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$';

-- +migrate Down
ALTER TABLE users DROP COLUMN email;
