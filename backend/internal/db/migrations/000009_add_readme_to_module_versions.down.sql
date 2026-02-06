-- Reverse migration for adding readme column
ALTER TABLE module_versions DROP COLUMN IF EXISTS readme;
