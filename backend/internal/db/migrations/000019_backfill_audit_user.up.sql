-- Backfill created_by and published_by fields for existing data
-- This sets all NULL values to the admin@dev.local user

-- First, get the user ID for admin@dev.local
-- Using: d3d54cbf-071b-4835-9563-529681a60a99 (from LoginPage.tsx mock user)

-- Update modules created_by
UPDATE modules
SET created_by = 'd3d54cbf-071b-4835-9563-529681a60a99'
WHERE created_by IS NULL;

-- Update providers created_by
UPDATE providers
SET created_by = 'd3d54cbf-071b-4835-9563-529681a60a99'
WHERE created_by IS NULL;

-- Update module_versions published_by
UPDATE module_versions
SET published_by = 'd3d54cbf-071b-4835-9563-529681a60a99'
WHERE published_by IS NULL;

-- Update provider_versions published_by
UPDATE provider_versions
SET published_by = 'd3d54cbf-071b-4835-9563-529681a60a99'
WHERE published_by IS NULL;

-- Update api_keys user_id (if the column exists and has NULL values)
UPDATE api_keys
SET user_id = 'd3d54cbf-071b-4835-9563-529681a60a99'
WHERE user_id IS NULL;
