-- +migrate Up

CREATE SEQUENCE public.user_grants_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.user_grants (
    id BIGINT NOT NULL DEFAULT nextval('public.user_grants_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,
    userid BIGINT NOT NULL,
    
    clientid TEXT NOT NULL,
    scope TEXT NOT NULL,
    granteduntil TIMESTAMP NULL,
    
    createdat TIMESTAMP NOT NULL DEFAULT now(),
    updatedat TIMESTAMP NOT NULL DEFAULT now(),
    deletedat TIMESTAMP NULL,
    
    CONSTRAINT user_grants_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id),
    CONSTRAINT user_grants_unique UNIQUE (tenantid, userid, clientid, scope)
);
-- +migrate StatementEnd

CREATE INDEX idx_user_grants_tenant_user_client ON public.user_grants (tenantid, userid, clientid) WHERE deletedat IS NULL;