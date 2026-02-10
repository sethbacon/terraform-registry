-- Remove created_by columns from modules and providers

DROP INDEX IF EXISTS idx_modules_created_by;
DROP INDEX IF EXISTS idx_providers_created_by;

ALTER TABLE modules DROP COLUMN IF EXISTS created_by;
ALTER TABLE providers DROP COLUMN IF EXISTS created_by;
