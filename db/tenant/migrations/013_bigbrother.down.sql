-- 013_bigbrother.down.sql
DROP TABLE IF EXISTS bb_pending_writes;
DROP TABLE IF EXISTS bb_credentials;
DROP TABLE IF EXISTS bb_events;
DROP TABLE IF EXISTS bb_computer_info;
DROP TABLE IF EXISTS bb_plc_registers;
DROP TABLE IF EXISTS bb_capabilities;
DROP TABLE IF EXISTS bb_ports;
DROP TABLE IF EXISTS bb_devices;
ALTER TABLE audit_log DROP COLUMN IF EXISTS tenant_id;
