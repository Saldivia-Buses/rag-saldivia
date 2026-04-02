-- Platform DB — initial schema
-- This database is the control plane. It stores tenant registry,
-- subscription plans, feature flags, and global configuration.
-- Each tenant has its OWN PostgreSQL instance — this DB only tracks metadata.

-- Subscription plans
CREATE TABLE plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,                          -- 'Starter', 'Business', 'Professional', 'Enterprise'
    max_users   INTEGER NOT NULL DEFAULT 10,
    max_storage_mb INTEGER NOT NULL DEFAULT 5120,       -- 5GB default
    ai_credits_monthly INTEGER NOT NULL DEFAULT 1000,
    price_usd   NUMERIC(10,2) NOT NULL DEFAULT 0,
    features    JSONB NOT NULL DEFAULT '{}',             -- feature flags included in this plan
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tenant registry
CREATE TABLE tenants (
    id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    slug         TEXT NOT NULL UNIQUE,                   -- subdomain: 'saldivia', 'empresa2'
    name         TEXT NOT NULL,                          -- display name: 'Saldivia Buses'
    plan_id      TEXT NOT NULL REFERENCES plans(id),
    postgres_url TEXT NOT NULL,                          -- connection string for tenant's PG instance
    redis_url    TEXT NOT NULL,                          -- connection string for tenant's Redis instance
    enabled      BOOLEAN NOT NULL DEFAULT true,
    logo_url     TEXT,
    domain       TEXT,                                   -- custom domain (enterprise)
    settings     JSONB NOT NULL DEFAULT '{}',            -- tenant-level settings
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);

-- Module registry — all available modules in the system
CREATE TABLE modules (
    id          TEXT PRIMARY KEY,                        -- 'fleet', 'construction', 'crm', 'docai'
    name        TEXT NOT NULL,                           -- 'Gestion de Flota'
    category    TEXT NOT NULL,                           -- 'core', 'platform', 'vertical', 'ai_service'
    description TEXT,
    icon        TEXT,                                    -- lucide icon name
    version     TEXT NOT NULL DEFAULT '0.1.0',
    requires    TEXT[] DEFAULT '{}',                     -- dependencies: ['docs']
    tier_min    TEXT NOT NULL DEFAULT 'starter',         -- minimum plan tier
    enabled     BOOLEAN NOT NULL DEFAULT true            -- globally available
);

-- Which modules each tenant has
CREATE TABLE tenant_modules (
    tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id   TEXT NOT NULL REFERENCES modules(id),
    enabled     BOOLEAN NOT NULL DEFAULT true,
    config      JSONB NOT NULL DEFAULT '{}',             -- per-tenant module config
    enabled_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    enabled_by  TEXT NOT NULL,                           -- who activated it
    PRIMARY KEY (tenant_id, module_id)
);

-- Global configuration
CREATE TABLE global_config (
    key         TEXT PRIMARY KEY,
    value       JSONB NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  TEXT NOT NULL
);

-- RAG models registry
CREATE TABLE rag_models (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    provider    TEXT NOT NULL,                           -- 'local', 'nvidia-api', 'openrouter'
    endpoint    TEXT NOT NULL,
    model_name  TEXT NOT NULL,
    category    TEXT NOT NULL,                           -- 'llm', 'embedding', 'reranker', 'vlm'
    vram_mb     INTEGER DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    config      JSONB NOT NULL DEFAULT '{}'
);

-- Feature flags (global or per-tenant)
CREATE TABLE feature_flags (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    tenant_id   TEXT REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL = global
    enabled     BOOLEAN NOT NULL DEFAULT false,
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feature_flags_tenant ON feature_flags(tenant_id);

-- Deploy log
CREATE TABLE deploy_log (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    service     TEXT NOT NULL,
    version_from TEXT NOT NULL,
    version_to  TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',         -- 'pending', 'success', 'failed', 'rollback'
    deployed_by TEXT NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    notes       TEXT
);

-- Seed default plans
INSERT INTO plans (id, name, max_users, max_storage_mb, ai_credits_monthly, price_usd) VALUES
    ('starter',       'Starter',       10,   5120,    1000,   49),
    ('business',      'Business',      50,   51200,   5000,   299),
    ('professional',  'Professional',  200,  102400,  20000,  999),
    ('enterprise',    'Enterprise',    -1,   -1,      -1,     0);  -- -1 = unlimited, custom pricing

-- Seed core modules (always active, can't be disabled)
INSERT INTO modules (id, name, category, tier_min) VALUES
    ('chat',          'Chat + RAG',           'core',        'starter'),
    ('auth',          'Auth + RBAC',          'core',        'starter'),
    ('notifications', 'Notificaciones',       'core',        'starter');

-- Seed platform modules
INSERT INTO modules (id, name, category, tier_min) VALUES
    ('docs',          'Gestion Documental',   'platform',    'starter'),
    ('kb',            'Knowledge Base',        'platform',    'starter'),
    ('tasks',         'Tareas/Tickets',        'platform',    'starter'),
    ('forms',         'Formularios',           'platform',    'starter'),
    ('announcements', 'Anuncios',              'platform',    'business'),
    ('crm',           'CRM',                   'platform',    'business'),
    ('workflows',     'Workflows',             'platform',    'business'),
    ('calendar',      'Calendario',            'platform',    'business'),
    ('reports',       'Reportes',              'platform',    'professional'),
    ('whatsapp',      'WhatsApp Business',     'platform',    'business'),
    ('email',         'Email Integration',     'platform',    'business');

-- Seed vertical modules
INSERT INTO modules (id, name, category, tier_min) VALUES
    ('fleet',         'Transporte/Logistica',  'vertical',    'starter'),
    ('construction',  'Construccion',           'vertical',    'starter'),
    ('professional',  'Servicios Profesionales','vertical',   'starter');

-- Seed AI service modules
INSERT INTO modules (id, name, category, tier_min) VALUES
    ('agents',        'Agentes Empresariales', 'ai_service',  'business'),
    ('docai',         'Document AI',           'ai_service',  'starter'),
    ('speech',        'Speech (ASR/TTS)',      'ai_service',  'business'),
    ('vision',        'Vision (Deteccion)',    'ai_service',  'business'),
    ('optimization',  'Optimization (Rutas)',  'ai_service',  'business'),
    ('feedback',      'Calidad IA',            'ai_service',  'starter'),
    ('forecast',      'Forecasting',           'ai_service',  'professional'),
    ('translation',   'Traduccion',            'ai_service',  'business');
