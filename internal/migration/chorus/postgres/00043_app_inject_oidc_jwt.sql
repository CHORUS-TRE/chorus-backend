-- +migrate Up

ALTER TABLE public.apps ADD COLUMN injectoidcjwtclientid TEXT NOT NULL DEFAULT '';
ALTER TABLE public.app_instances ADD COLUMN injectoidcjwttoken TEXT NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE public.apps DROP COLUMN injectoidcjwtclientid;
ALTER TABLE public.app_instances DROP COLUMN injectoidcjwttoken;
