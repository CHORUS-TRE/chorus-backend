-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE public.audit DROP COLUMN resourcetype;
ALTER TABLE public.audit DROP COLUMN resourceid;
ALTER TABLE public.audit DROP COLUMN method;
ALTER TABLE public.audit DROP COLUMN statuscode;
ALTER TABLE public.audit DROP COLUMN errormessage;
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE public.audit ADD COLUMN workspaceid BIGINT NULL;
ALTER TABLE public.audit ADD COLUMN workbenchid BIGINT NULL;

CREATE INDEX idx_audit_workspaceid ON public.audit(workspaceid);
CREATE INDEX idx_audit_workbenchid ON public.audit(workbenchid);
-- +migrate StatementEnd