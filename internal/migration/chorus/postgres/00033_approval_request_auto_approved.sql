-- +migrate Up

ALTER TABLE public.approval_requests ADD COLUMN autoapproved BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE public.approval_requests ADD COLUMN approvalmessage TEXT NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE public.approval_requests DROP COLUMN autoapproved;
ALTER TABLE public.approval_requests DROP COLUMN approvalmessage;
