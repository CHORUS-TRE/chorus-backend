-- +migrate Up

ALTER TABLE public.notifications ADD COLUMN type TEXT NOT NULL;
ALTER TABLE public.notifications ADD COLUMN refreshjwtrequired BOOLEAN NOT NULL DEFAULT false;

