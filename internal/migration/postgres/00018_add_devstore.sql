-- +migrate Up

CREATE SEQUENCE public.devstore_id_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE devstore (
    id BIGINT NOT NULL DEFAULT nextval('public.devstore_id_seq'::REGCLASS),

    tenantid BIGINT NOT NULL,

    scope TEXT NOT NULL,
    scopeid BIGINT NOT NULL,

    key TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '',

    createdat TIMESTAMP NULL,
    updatedat TIMESTAMP NULL,
    
    CONSTRAINT devstore_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    UNIQUE (tenantid, scope, scopeid, key)
);
-- +migrate StatementEnd