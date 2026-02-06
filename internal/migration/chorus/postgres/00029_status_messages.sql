-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE public.workbenches ADD COLUMN serverpodmessage TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE public.app_instances ADD COLUMN k8smessage TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
