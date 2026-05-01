-- +migrate Up

ALTER TABLE public.apps ADD COLUMN IF NOT EXISTS stabilitystatus TEXT NOT NULL DEFAULT 'ready';
