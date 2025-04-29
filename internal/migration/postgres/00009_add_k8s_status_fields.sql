-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workbenchs ADD COLUMN k8sstatus TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE app_instances ADD COLUMN k8sstate TEXT NOT NULL DEFAULT '';
ALTER TABLE app_instances ADD COLUMN k8sstatus TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
