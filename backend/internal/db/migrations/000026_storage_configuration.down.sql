-- Migration 026 rollback: Remove storage configuration tables

DROP TABLE IF EXISTS storage_config;
DROP TABLE IF EXISTS system_settings;
