-- SDA Framework — PostgreSQL initialization for development
-- Creates two databases: platform (global) and tenant_dev (first dev tenant).
-- In production, each tenant gets its own PostgreSQL INSTANCE (not just a database).

CREATE DATABASE sda_platform;
CREATE DATABASE sda_tenant_dev;

-- Grant access
GRANT ALL PRIVILEGES ON DATABASE sda_platform TO sda;
GRANT ALL PRIVILEGES ON DATABASE sda_tenant_dev TO sda;
