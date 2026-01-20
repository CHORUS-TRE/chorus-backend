-- +migrate Up

CREATE SEQUENCE public.audit_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.audit (
    id BIGINT NOT NULL DEFAULT nextval('public.audit_seq'::REGCLASS),
    tenantid BIGINT NULL,
    userid BIGINT NULL,
    username TEXT NULL,
    action TEXT NULL,
    resourcetype TEXT NULL,
    resourceid BIGINT NULL,
    correlationid TEXT NULL,
    method TEXT NULL,
    statuscode INT NULL,
    errormessage TEXT NULL,
    description TEXT NULL,
    details JSONB NULL,
    createdat TIMESTAMP NOT NULL,
    CONSTRAINT audit_pkey PRIMARY KEY (id)
);
-- +migrate StatementEnd
