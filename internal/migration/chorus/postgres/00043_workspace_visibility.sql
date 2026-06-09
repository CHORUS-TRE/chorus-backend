-- +migrate Up

ALTER TABLE public.platform_settings
    ADD COLUMN defaultworkspacevisibility TEXT NOT NULL DEFAULT 'private';

ALTER TABLE public.workspaces
    ADD COLUMN visibility    TEXT   NOT NULL DEFAULT 'private',
    ADD COLUMN contactuserid BIGINT NULL REFERENCES users(id);

-- +migrate Down

ALTER TABLE public.workspaces
    DROP COLUMN IF EXISTS contactuserid,
    DROP COLUMN IF EXISTS visibility;

ALTER TABLE public.platform_settings
    DROP COLUMN IF EXISTS defaultworkspacevisibility;
