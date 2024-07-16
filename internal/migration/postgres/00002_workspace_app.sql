-- +migrate Up

CREATE SEQUENCE public.workspaces_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.workspaces (
    id BIGINT NOT NULL DEFAULT nextval('public.workspaces_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,
    userid BIGINT NOT NULL,
    
    name TEXT NOT NULL,
    shortname TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL,
    
    createdat TIMESTAMP NOT NULL,
    updatedat TIMESTAMP NOT NULL,
    deletedat TIMESTAMP NULL,
    
    CONSTRAINT workspaces_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id)
);
-- +migrate StatementEnd


CREATE SEQUENCE public.apps_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.apps (
    id BIGINT NOT NULL DEFAULT nextval('public.apps_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,
    userid BIGINT NOT NULL,
    
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL,
    
    dockerimagename TEXT NOT NULL,
    dockerimagetag TEXT NOT NULL,
    
    createdat TIMESTAMP NOT NULL,
    updatedat TIMESTAMP NOT NULL,
    deletedat TIMESTAMP NULL,
    
    CONSTRAINT apps_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.app_instances_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.app_instances (
    id BIGINT NOT NULL DEFAULT nextval('public.app_instances_seq'::REGCLASS),
    
    tenantid BIGINT NOT NULL,
    userid BIGINT NOT NULL,
    appid BIGINT NOT NULL,
    workspaceid BIGINT NOT NULL,
    
    status TEXT NOT NULL,
    
    createdat TIMESTAMP NOT NULL,
    updatedat TIMESTAMP NOT NULL,
    deletedat TIMESTAMP NULL,
    
    CONSTRAINT app_instances_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id),
    CONSTRAINT appcon FOREIGN KEY (appid) REFERENCES apps(id),
    CONSTRAINT workspacecon FOREIGN KEY (workspaceid) REFERENCES workspaces(id)
);
-- +migrate StatementEnd
