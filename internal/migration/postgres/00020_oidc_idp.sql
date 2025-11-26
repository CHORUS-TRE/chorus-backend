-- +migrate Up

-- Authentication Sessions Table
CREATE SEQUENCE public.authn_sessions_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.authn_sessions (
    id TEXT NOT NULL,
    tenantid BIGINT NULL,
    session_data JSONB NOT NULL,
    callbackid TEXT NULL,
    authcode TEXT NULL,
    pushedauthreqid TEXT NULL,
    cibaauthid TEXT NULL,
    createdattimestamp INTEGER NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT now(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT authn_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id)
);
-- +migrate StatementEnd

-- Indexes for authentication session lookups
CREATE INDEX idx_authn_sessions_callbackid ON public.authn_sessions(callbackid) WHERE callbackid IS NOT NULL;
CREATE INDEX idx_authn_sessions_authcode ON public.authn_sessions(authcode) WHERE authcode IS NOT NULL;
CREATE INDEX idx_authn_sessions_pushedauthreqid ON public.authn_sessions(pushedauthreqid) WHERE pushedauthreqid IS NOT NULL;
CREATE INDEX idx_authn_sessions_cibaauthid ON public.authn_sessions(cibaauthid) WHERE cibaauthid IS NOT NULL;
CREATE INDEX idx_authn_sessions_createdattimestamp ON public.authn_sessions(createdattimestamp);
CREATE INDEX idx_authn_sessions_tenantid ON public.authn_sessions(tenantid);

-- Grant Sessions Table
CREATE SEQUENCE public.grant_sessions_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.grant_sessions (
    id TEXT NOT NULL,
    tenantid BIGINT NULL,
    session_data JSONB NOT NULL,
    tokenid TEXT NULL,
    refreshtoken TEXT NULL,
    authcode TEXT NULL,
    createdattimestamp INTEGER NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT now(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT grant_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id)
);
-- +migrate StatementEnd

-- Indexes for grant session lookups
CREATE INDEX idx_grant_sessions_tokenid ON public.grant_sessions(tokenid) WHERE tokenid IS NOT NULL;
CREATE INDEX idx_grant_sessions_refreshtoken ON public.grant_sessions(refreshtoken) WHERE refreshtoken IS NOT NULL;
CREATE INDEX idx_grant_sessions_authcode ON public.grant_sessions(authcode) WHERE authcode IS NOT NULL;
CREATE INDEX idx_grant_sessions_createdattimestamp ON public.grant_sessions(createdattimestamp);
CREATE INDEX idx_grant_sessions_tenantid ON public.grant_sessions(tenantid);

-- Logout Sessions Table
CREATE SEQUENCE public.logout_sessions_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.logout_sessions (
    id TEXT NOT NULL,
    tenantid BIGINT NULL,
    session_data JSONB NOT NULL,
    callbackid TEXT NULL,
    createdattimestamp INTEGER NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT now(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT logout_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id)
);
-- +migrate StatementEnd

-- Indexes for logout session lookups
CREATE INDEX idx_logout_sessions_callbackid ON public.logout_sessions(callbackid) WHERE callbackid IS NOT NULL;
CREATE INDEX idx_logout_sessions_createdattimestamp ON public.logout_sessions(createdattimestamp);
CREATE INDEX idx_logout_sessions_tenantid ON public.logout_sessions(tenantid);
