-- +migrate Up

CREATE SEQUENCE public.platform_settings_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
CREATE TABLE IF NOT EXISTS public.platform_settings (
    id BIGINT NOT NULL DEFAULT nextval('public.platform_settings_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,

    title       TEXT NOT NULL DEFAULT '',
    headline    TEXT NOT NULL DEFAULT '',
    tagline     TEXT NOT NULL DEFAULT '',
    websiteurl TEXT NOT NULL DEFAULT '',

    touversionid BIGINT NOT NULL DEFAULT 0,

    maxworkspacesperuser    BIGINT NOT NULL DEFAULT 0,
    maxsessionsperuser      BIGINT NOT NULL DEFAULT 0,
    maxappinstancesperuser BIGINT NOT NULL DEFAULT 0,
    
    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    UNIQUE (tenantid)
);
