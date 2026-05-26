-- +migrate Up

ALTER TABLE public.apps RENAME COLUMN kioskconfigurl TO browserconfigurl;
ALTER TABLE public.apps RENAME COLUMN kioskconfigjwturl TO browserconfigjwturl;
ALTER TABLE public.apps RENAME COLUMN kioskconfigjwtoidcclientid TO browserconfigjwtoidcclientid;

ALTER TABLE public.app_instances RENAME COLUMN kioskconfigjwttoken TO browserconfigjwttoken;

-- +migrate Down

ALTER TABLE public.apps RENAME COLUMN browserconfigurl TO kioskconfigurl;
ALTER TABLE public.apps RENAME COLUMN browserconfigjwturl TO kioskconfigjwturl;
ALTER TABLE public.apps RENAME COLUMN browserconfigjwtoidcclientid TO kioskconfigjwtoidcclientid;

ALTER TABLE public.app_instances RENAME COLUMN browserconfigjwttoken TO kioskconfigjwttoken;