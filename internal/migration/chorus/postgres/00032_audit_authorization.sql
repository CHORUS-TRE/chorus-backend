-- +migrate Up

-- +migrate StatementBegin
INSERT INTO role_definitions (name) VALUES
('PlatformAuditor')
ON CONFLICT (name) DO NOTHING;
-- +migrate StatementEnd
