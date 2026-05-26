-- +migrate Up

ALTER TABLE public.apps RENAME COLUMN kioskconfigurl TO browserconfigurl;
ALTER TABLE public.apps RENAME COLUMN kioskconfigjwturl TO browserconfigjwturl;
ALTER TABLE public.apps RENAME COLUMN kioskconfigjwtoidcclientid TO browserconfigjwtoidcclientid;

ALTER TABLE public.app_instances RENAME COLUMN kioskconfigjwttoken TO browserconfigjwttoken;