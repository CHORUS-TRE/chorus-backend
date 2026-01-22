-- +migrate Up

DELETE FROM public.notifications_read_by_archive;
DELETE FROM public.notifications_read_by;
DELETE FROM public.notifications;

ALTER TABLE public.notifications DROP COLUMN IF EXISTS type;
ALTER TABLE public.notifications DROP COLUMN IF EXISTS refreshjwtrequired;
ALTER TABLE public.notifications ADD COLUMN userid BIGINT NOT NULL;
ALTER TABLE public.notifications ADD CONSTRAINT notifications_userid_fkey FOREIGN KEY (userid) REFERENCES public.users(id);
ALTER TABLE public.notifications ADD COLUMN content JSONB NOT NULL DEFAULT '{}';

-- +migrate Down

ALTER TABLE public.notifications DROP CONSTRAINT IF EXISTS notifications_userid_fkey;
ALTER TABLE public.notifications DROP COLUMN IF EXISTS userid;
ALTER TABLE public.notifications DROP COLUMN IF EXISTS content;
ALTER TABLE public.notifications ADD COLUMN type TEXT NOT NULL DEFAULT '';
ALTER TABLE public.notifications ADD COLUMN refreshjwtrequired BOOLEAN NOT NULL DEFAULT false;
