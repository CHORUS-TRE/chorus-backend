-- +migrate Up

-- +migrate StatementBegin
INSERT INTO role_definitions (name) VALUES
('Authenticated'), ('SuperAdmin') 
ON CONFLICT (name) DO NOTHING;
-- +migrate StatementEnd
