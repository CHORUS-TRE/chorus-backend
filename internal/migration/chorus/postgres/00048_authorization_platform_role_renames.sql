-- +migrate Up

-- +migrate StatementBegin
UPDATE role_definitions SET name = 'PlatformUserManager' WHERE name = 'PlateformUserManager';
-- +migrate StatementEnd

-- +migrate StatementBegin
UPDATE role_definitions SET name = 'PlatformDataManager' WHERE name = 'DataManager';
-- +migrate StatementEnd
