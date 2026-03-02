-- +migrate Up

CREATE SEQUENCE public.job_locks_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.job_locks (
    id BIGINT NOT NULL DEFAULT nextval('public.job_locks_seq'::REGCLASS),
    name TEXT NOT NULL,
    owner TEXT NOT NULL,
    lockedat TIMESTAMP NOT NULL DEFAULT NOW(),
    expiresat TIMESTAMP NOT NULL,
    completedat TIMESTAMP,
    status TEXT,
    message TEXT,
    CONSTRAINT job_locks_pkey PRIMARY KEY (id)
);
-- +migrate StatementEnd

CREATE UNIQUE INDEX job_locks_name_active_idx ON public.job_locks (name) WHERE completedat IS NULL;
CREATE INDEX job_locks_expiresat_idx ON public.job_locks (expiresat) WHERE completedat IS NULL;
CREATE INDEX job_locks_completedat_idx ON public.job_locks (completedat);
