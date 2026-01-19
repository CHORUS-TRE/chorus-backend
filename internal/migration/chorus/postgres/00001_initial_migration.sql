-- +migrate Up

CREATE SEQUENCE public.tenants_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.tenants (
    id BIGINT NOT NULL DEFAULT nextval('public.tenants_seq'::REGCLASS),
    name TEXT NULL,
    createdat TIMESTAMP NULL,
    updatedat TIMESTAMP NULL,
    CONSTRAINT tenants_pkey PRIMARY KEY (id),
    UNIQUE (name)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.users_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.users (
    id BIGINT NOT NULL DEFAULT nextval('public.users_seq'::REGCLASS),
    tenantid BIGINT NULL,
    firstname TEXT NULL,
    lastname TEXT NULL,
    username TEXT NULL,
    password TEXT NULL,
    status TEXT NULL,
    totpsecret TEXT NULL,
    createdat TIMESTAMP NULL,
    updatedat TIMESTAMP NULL,
    totpenabled BOOLEAN NULL DEFAULT false,
    passwordchanged BOOLEAN NULL DEFAULT false,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    UNIQUE (username)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.totp_recovery_codes_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.totp_recovery_codes (
    id BIGINT NOT NULL DEFAULT nextval('public.totp_recovery_codes_seq'::REGCLASS),
    userid BIGINT NULL,
    code TEXT NULL,
    tenantid BIGINT NULL,
    CONSTRAINT totp_recovery_codes_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.roles_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.roles (
    id BIGINT NOT NULL DEFAULT nextval('public.roles_seq'::REGCLASS),
    name TEXT NOT NULL,
    CONSTRAINT roles_pkey PRIMARY KEY (id),
    UNIQUE (name)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.user_role_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.user_role (
    id BIGINT NOT NULL DEFAULT nextval('public.user_role_seq'::REGCLASS),
    userid BIGINT NULL,
    roleid BIGINT NULL,
    CONSTRAINT usercon FOREIGN KEY (userid) REFERENCES users(id),
    CONSTRAINT rolecon FOREIGN KEY (roleid) REFERENCES roles(id),
    CONSTRAINT user_role_pkey PRIMARY KEY (id)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.attachments_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.attachments (
    id BIGINT NOT NULL DEFAULT nextval('public.attachments_seq'::REGCLASS),
    tenantid BIGINT NULL,
    resourceid TEXT NULL,
    resourcetype TEXT NULL,
    filename TEXT NULL,
    key TEXT NULL,
    value TEXT NULL,
    contenttype TEXT NULL,
    location TEXT NULL DEFAULT 'local',
    documentcategory TEXT NULL,
    createdat TIMESTAMP NULL,
    updatedat TIMESTAMP NULL,
    deleted BOOLEAN NULL DEFAULT false,
    CONSTRAINT attachments_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id)
);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE public.notifications (
    id TEXT NOT NULL,
    tenantid BIGINT NULL,
    message TEXT NOT NULL,
    createdat TIMESTAMP NULL DEFAULT now(),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT notifications_pkey PRIMARY KEY (id)
);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE public.notifications_read_by (
    tenantid BIGINT NULL,
    notificationid TEXT NOT NULL,
    userid BIGINT NOT NULL,
    readat TIMESTAMP NULL,
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT notifications_read_by_notificationid_fkey FOREIGN KEY (notificationid) REFERENCES public.notifications(id),
    CONSTRAINT notifications_read_by_userid_fkey FOREIGN KEY (userid) REFERENCES public.users(id),
    CONSTRAINT notifications_read_by_pkey PRIMARY KEY (notificationid, userid)
);
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TABLE public.notifications_read_by_archive (
    tenantid BIGINT NULL,
    notificationid TEXT NOT NULL,
    userid BIGINT NOT NULL,
    username TEXT NULL,
    readat TIMESTAMP NULL,
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT notifications_read_by_archive_notificationid_fkey FOREIGN KEY (notificationid) REFERENCES public.notifications(id),
    CONSTRAINT notifications_read_by_archive_pkey PRIMARY KEY (notificationid, userid)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.files_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.files (
    id BIGINT NOT NULL DEFAULT nextval('public.files_seq'::REGCLASS),
    tenantid BIGINT NOT NULL,
    key TEXT NOT NULL,
    hash TEXT NOT NULL,
    status TEXT NOT NULL,
    createdat TIMESTAMPTZ NOT NULL DEFAULT now(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT files_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    UNIQUE (tenantid, key)
);
-- +migrate StatementEnd

CREATE SEQUENCE public.file_attributes_seq MINVALUE 1 MAXVALUE 9007199254740991 INCREMENT 1 START 1;
-- +migrate StatementBegin
CREATE TABLE public.file_attributes (
    id BIGINT NOT NULL DEFAULT nextval('public.file_attributes_seq'::REGCLASS),
    tenantid BIGINT NOT NULL,
    fileid BIGINT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    contenttype TEXT NOT NULL DEFAULT '',
    createdat TIMESTAMPTZ NOT NULL DEFAULT now(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT file_attributes_pkey PRIMARY KEY (id),
    CONSTRAINT tenantcon FOREIGN KEY (tenantid) REFERENCES tenants(id),
    CONSTRAINT file_attributes_fileid_fkey FOREIGN KEY (fileid) REFERENCES public.files(id),
    UNIQUE (tenantid, fileid, key)
);
-- +migrate StatementEnd
