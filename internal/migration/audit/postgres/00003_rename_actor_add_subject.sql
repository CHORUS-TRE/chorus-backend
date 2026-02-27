-- +migrate Up

-- Rename existing user fields to actor (these currently store the JWT user)
-- +migrate StatementBegin
ALTER TABLE public.audit RENAME COLUMN userid TO actorid;
ALTER TABLE public.audit RENAME COLUMN username TO actorusername;
-- +migrate StatementEnd

-- Add userid as a context field (the user being acted upon, like workspaceid/workbenchid)
-- +migrate StatementBegin
ALTER TABLE public.audit ADD COLUMN userid BIGINT NULL;
CREATE INDEX idx_audit_userid ON public.audit(userid);
-- +migrate StatementEnd
