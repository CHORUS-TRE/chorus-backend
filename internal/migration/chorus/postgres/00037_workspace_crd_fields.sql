-- +migrate Up

ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicy TEXT NOT NULL DEFAULT 'Airgapped';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS allowedfqdns TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicystatus TEXT NOT NULL DEFAULT '';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS networkpolicymessage TEXT NOT NULL DEFAULT '';
ALTER TABLE public.workspaces ADD COLUMN IF NOT EXISTS clipboard TEXT NOT NULL DEFAULT 'disabled';

CREATE SEQUENCE public.workspace_services_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.workspace_services (
    id BIGINT NOT NULL DEFAULT nextval('public.workspace_services_seq'::REGCLASS),

    tenantid BIGINT NOT NULL,
    workspaceid BIGINT NOT NULL,
    name TEXT NOT NULL,

    state TEXT NOT NULL DEFAULT 'Running',

    chartregistry TEXT NOT NULL DEFAULT '',
    chartrepository TEXT NOT NULL DEFAULT '',
    charttag TEXT NOT NULL DEFAULT '',

    valuesoverride JSONB NOT NULL DEFAULT '{}',

    credentialssecretname TEXT NOT NULL DEFAULT '',
    credentialspaths TEXT[] NOT NULL DEFAULT '{}',

    connectioninfotemplate TEXT NOT NULL DEFAULT '',
    computedvalues JSONB NOT NULL DEFAULT '{}',

    status TEXT NOT NULL DEFAULT '',
    statusmessage TEXT NOT NULL DEFAULT '',
    connectioninfo TEXT NOT NULL DEFAULT '',
    secretname TEXT NOT NULL DEFAULT '',

    createdat TIMESTAMP NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMP NOT NULL DEFAULT NOW(),
    deletedat TIMESTAMP NULL,

    CONSTRAINT workspace_services_pkey PRIMARY KEY (id),
    CONSTRAINT workspace_services_tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT workspace_services_workspacecon FOREIGN KEY (workspaceid) REFERENCES workspaces(id)
);
-- +migrate StatementEnd

CREATE UNIQUE INDEX workspace_services_workspace_name_idx ON public.workspace_services (workspaceid, name) WHERE deletedat IS NULL;

-- +migrate Down

DROP INDEX IF EXISTS public.workspace_services_workspace_name_idx;
DROP TABLE IF EXISTS public.workspace_services;
DROP SEQUENCE IF EXISTS public.workspace_services_seq;

ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicy;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS allowedfqdns;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicystatus;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS networkpolicymessage;
ALTER TABLE public.workspaces DROP COLUMN IF EXISTS clipboard;
