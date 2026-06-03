-- +migrate Up

CREATE SEQUENCE public.terms_of_use_versions_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
CREATE TABLE IF NOT EXISTS public.terms_of_use_versions (
    id BIGINT NOT NULL DEFAULT nextval('public.terms_of_use_versions_seq'::REGCLASS),

    tenantid BIGINT NOT NULL,

    content TEXT NOT NULL DEFAULT '',
    status  TEXT NOT NULL DEFAULT '',

    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT terms_of_use_versions_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id)
);

-- Create a unique index to ensure only one published terms of use version per tenant
CREATE UNIQUE INDEX terms_of_use_versions_one_published_per_tenant
    ON public.terms_of_use_versions (tenantid) WHERE status = 'Published';

CREATE SEQUENCE public.terms_of_use_acceptances_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
CREATE TABLE IF NOT EXISTS public.terms_of_use_acceptances (
    id BIGINT NOT NULL DEFAULT nextval('public.terms_of_use_acceptances_seq'::REGCLASS),

    tenantid     BIGINT NOT NULL,

    userid       BIGINT NOT NULL,
    termsofuseversionid BIGINT NOT NULL,
    acceptedat TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT terms_of_use_acceptances_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT touversioncon FOREIGN KEY (termsofuseversionid) REFERENCES terms_of_use_versions(id),
    UNIQUE (tenantid, userid, termsofuseversionid)
);

-- +migrate Down
DROP TABLE IF EXISTS public.terms_of_use_acceptances;
DROP SEQUENCE IF EXISTS public.terms_of_use_acceptances_seq;
DROP TABLE IF EXISTS public.terms_of_use_versions;
DROP SEQUENCE IF EXISTS public.terms_of_use_versions_seq;
