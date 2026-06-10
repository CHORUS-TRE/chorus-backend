-- +migrate Up

ALTER TABLE public.apps ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';
