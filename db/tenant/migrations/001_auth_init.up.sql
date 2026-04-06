-- Tenant DB — auth tables (applied to every tenant's PostgreSQL instance)
-- This is the base schema that every tenant starts with.

-- Users
CREATE TABLE IF NOT EXISTS users (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email           TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    avatar_url      TEXT,
    mfa_secret      TEXT,                                    -- TOTP secret (encrypted)
    mfa_enabled     BOOLEAN NOT NULL DEFAULT false,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    failed_logins   INTEGER NOT NULL DEFAULT 0,
    locked_until    TIMESTAMPTZ,                             -- brute force lockout
    last_login_at   TIMESTAMPTZ,
    last_login_ip   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Note: email already has a UNIQUE constraint which creates an implicit index.

-- Roles
CREATE TABLE IF NOT EXISTS roles (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL UNIQUE,                         -- 'admin', 'manager', 'user', 'viewer'
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT false,               -- system roles can't be deleted
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Permissions
CREATE TABLE IF NOT EXISTS permissions (
    id          TEXT PRIMARY KEY,                              -- 'users.read', 'collections.write'
    name        TEXT NOT NULL,
    description TEXT,
    category    TEXT NOT NULL                                  -- 'users', 'collections', 'config', etc.
);

-- Role-Permission mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User-Role mapping
CREATE TABLE IF NOT EXISTS user_roles (
    user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id  TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,                          -- bcrypt hash of the refresh token
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at  TIMESTAMPTZ                                    -- NULL = active
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- Audit log (immutable)
CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id     TEXT REFERENCES users(id),                     -- NULL for system events
    action      TEXT NOT NULL,                                 -- 'user.login', 'collection.create', etc.
    resource    TEXT,                                          -- what was affected
    details     JSONB NOT NULL DEFAULT '{}',                   -- action-specific data
    ip_address  TEXT,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at);

-- Seed system roles
INSERT INTO roles (id, name, description, is_system) VALUES
    ('role-admin',   'admin',   'Full access to all tenant features',  true),
    ('role-manager', 'manager', 'Manage users and content',            true),
    ('role-user',    'user',    'Standard user access',                true),
    ('role-viewer',  'viewer',  'Read-only access',                    true)
ON CONFLICT (id) DO NOTHING;

-- Seed base permissions
INSERT INTO permissions (id, name, description, category) VALUES
    ('users.read',        'Ver usuarios',           'Listar y ver perfiles de usuarios',     'users'),
    ('users.write',       'Gestionar usuarios',     'Crear, editar, eliminar usuarios',      'users'),
    ('roles.read',        'Ver roles',              'Listar roles y permisos',               'roles'),
    ('roles.write',       'Gestionar roles',        'Crear, editar, eliminar roles',         'roles'),
    ('collections.read',  'Ver colecciones',        'Listar y consultar colecciones RAG',    'collections'),
    ('collections.write', 'Gestionar colecciones',  'Crear, eliminar colecciones',           'collections'),
    ('chat.read',         'Usar chat',              'Enviar mensajes y ver historial',       'chat'),
    ('chat.write',        'Gestionar chat',         'Eliminar sesiones, moderar',            'chat'),
    ('config.read',       'Ver configuracion',      'Ver config del tenant',                 'config'),
    ('config.write',      'Gestionar configuracion','Cambiar config RAG, modulos',           'config'),
    ('docs.read',         'Ver documentos',         'Acceder a gestion documental',          'docs'),
    ('docs.write',        'Gestionar documentos',   'Subir, editar, eliminar documentos',    'docs'),
    ('ingest.write',      'Ingestar documentos',    'Subir documentos al pipeline RAG',      'ingest'),
    ('audit.read',        'Ver audit log',          'Consultar registro de auditoría',       'audit')
ON CONFLICT (id) DO NOTHING;

-- Assign all permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-admin', id FROM permissions
ON CONFLICT DO NOTHING;

-- Manager: everything except config.write, roles.write, audit.read
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'role-manager', id FROM permissions
WHERE id NOT IN ('config.write', 'roles.write', 'audit.read')
ON CONFLICT DO NOTHING;

-- User: read + chat + docs + ingest
INSERT INTO role_permissions (role_id, permission_id)
VALUES
    ('role-user', 'collections.read'),
    ('role-user', 'chat.read'),
    ('role-user', 'chat.write'),
    ('role-user', 'docs.read'),
    ('role-user', 'docs.write'),
    ('role-user', 'ingest.write')
ON CONFLICT DO NOTHING;

-- Viewer: read-only
INSERT INTO role_permissions (role_id, permission_id)
VALUES
    ('role-viewer', 'collections.read'),
    ('role-viewer', 'chat.read'),
    ('role-viewer', 'docs.read')
ON CONFLICT DO NOTHING;
