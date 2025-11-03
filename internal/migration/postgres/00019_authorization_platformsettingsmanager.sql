-- +migrate Up

-- +migrate StatementBegin
INSERT INTO role_definitions (name) VALUES
('PlatformSettingsManager');
-- +migrate StatementEnd
