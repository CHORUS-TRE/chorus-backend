-- +migrate Up

CREATE SEQUENCE public.approval_requests_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.approval_requests (
    id BIGINT NOT NULL DEFAULT nextval('public.approval_requests_seq'::REGCLASS),
    tenantid BIGINT NOT NULL,
    requesterid BIGINT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    details JSONB,
    approverids BIGINT[] DEFAULT '{}',
    approvedbyid BIGINT,
    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    approvedat TIMESTAMP,
    deletedat TIMESTAMP,
    CONSTRAINT approval_requests_pkey PRIMARY KEY (id),
    CONSTRAINT approval_requests_tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT approval_requests_requestercon FOREIGN KEY (requesterid) REFERENCES users(id),
    CONSTRAINT approval_requests_approvedbycon FOREIGN KEY (approvedbyid) REFERENCES users(id)
);
-- +migrate StatementEnd

CREATE INDEX approval_requests_tenantid_idx ON public.approval_requests (tenantid);
CREATE INDEX approval_requests_requesterid_idx ON public.approval_requests (requesterid);
CREATE INDEX approval_requests_status_idx ON public.approval_requests (status);
CREATE INDEX approval_requests_type_idx ON public.approval_requests (type);
