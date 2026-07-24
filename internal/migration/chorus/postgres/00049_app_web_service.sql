-- +migrate Up

ALTER TABLE public.apps ADD COLUMN webserviceport TEXT NOT NULL DEFAULT '';
ALTER TABLE public.apps ADD COLUMN webservicepath TEXT NOT NULL DEFAULT '';
ALTER TABLE public.apps ADD COLUMN webservicetokenparam TEXT NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE public.apps DROP COLUMN IF EXISTS webserviceport;
ALTER TABLE public.apps DROP COLUMN IF EXISTS webservicepath;
ALTER TABLE public.apps DROP COLUMN IF EXISTS webservicetokenparam;
