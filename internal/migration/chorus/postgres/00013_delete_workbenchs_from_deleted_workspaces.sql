-- +migrate Up

-- +migrate StatementBegin
UPDATE workbenchs
SET status = 'deleted',
    updatedat = NOW(),
    deletedat = NOW()
WHERE workspaceid IN (
    SELECT id
    FROM workspaces
    WHERE deletedat IS NOT NULL
)
AND deletedat IS NULL;
-- +migrate StatementEnd

-- +migrate StatementBegin
UPDATE app_instances
SET status = 'deleted',
    updatedat = NOW(),
    deletedat = NOW()
WHERE workspaceid IN (
    SELECT id
    FROM workspaces
    WHERE deletedat IS NOT NULL
)
AND deletedat IS NULL;
-- +migrate StatementEnd

