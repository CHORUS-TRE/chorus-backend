-- +migrate Up

CREATE SEQUENCE public.workbenchs_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.workbenchs (
    id BIGINT NOT NULL DEFAULT nextval('public.workbenchs_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,
    userid BIGINT NOT NULL,
    workspaceid BIGINT NOT NULL,
    
    name TEXT NOT NULL,
    shortname TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL,
    
    createdat TIMESTAMP NOT NULL,
    updatedat TIMESTAMP NOT NULL,
    deletedat TIMESTAMP NULL,
    
    CONSTRAINT workbenchs_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id),
    CONSTRAINT workspacecon FOREIGN KEY (workspaceid) REFERENCES workspaces(id)
);
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE app_instances ADD COLUMN workbenchid BIGINT NOT NULL DEFAULT 0 REFERENCES workbenchs(id);
-- +migrate StatementEnd