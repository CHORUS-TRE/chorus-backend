-- +migrate Up

CREATE SEQUENCE public.organizations_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
CREATE TABLE public.organizations (
    id BIGINT NOT NULL DEFAULT nextval('public.organizations_seq'::REGCLASS),

    tenantid BIGINT NOT NULL,

    name        TEXT NOT NULL,
    description TEXT,

    logo            BYTEA,
    logocontenttype TEXT,

    country VARCHAR(2),
    city    TEXT,

    contactuserid BIGINT NULL,
    websiteurl    TEXT,

    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    deletedat TIMESTAMP NULL,

    CONSTRAINT organizations_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT contactusercon FOREIGN KEY (contactuserid) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_organizations_tenantid ON organizations(tenantid) WHERE deletedat IS NULL;

-- +migrate Down

DROP TABLE IF EXISTS public.organizations;
DROP SEQUENCE IF EXISTS public.organizations_seq;
