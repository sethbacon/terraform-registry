-- Reverse migration for mirror configurations
DROP TABLE IF EXISTS mirror_sync_history CASCADE;
DROP TABLE IF EXISTS mirror_configurations CASCADE;
