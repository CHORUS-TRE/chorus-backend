-- +migrate Up

-- +migrate StatementBegin
INSERT INTO user_role_context (userroleid, contextdimension, value)
SELECT ur.id, 'user', ur.userid::text as val
FROM user_role ur where roleid=(SELECT id FROM role_definitions WHERE name = 'Authenticated')
-- +migrate StatementEnd
