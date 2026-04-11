-- +migrate Up

ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicy TEXT NOT NULL DEFAULT 'Airgapped';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS allowedfqdns TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicystatus TEXT NOT NULL DEFAULT '';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicymessage TEXT NOT NULL DEFAULT '';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS clipboard TEXT NOT NULL DEFAULT 'disabled';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS services JSONB NOT NULL DEFAULT '{}';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS servicestatuses JSONB NOT NULL DEFAULT '{}';

-- +migrate Down

ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicy;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS allowedfqdns;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicystatus;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicymessage;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS clipboard;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS services;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS servicestatuses;
