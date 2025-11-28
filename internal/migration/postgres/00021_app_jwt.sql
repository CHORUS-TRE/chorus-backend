-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE public.apps
ADD COLUMN IF NOT EXISTS kioskconfigjwturl TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS kioskconfigjwtoidcclientid TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE public.app_instances
ADD COLUMN IF NOT EXISTS kioskconfigjwttoken TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd
