-- Reverse the backfill (set back to NULL)
-- Note: This is destructive and will remove audit information

-- We can't reliably reverse this migration since we don't know which records
-- were originally NULL vs which ones had the admin user set intentionally.
-- This down migration is a no-op to be safe.

-- If you really need to clear the data:
-- UPDATE modules SET created_by = NULL WHERE created_by = 'd3d54cbf-071b-4835-9563-529681a60a99';
-- UPDATE providers SET created_by = NULL WHERE created_by = 'd3d54cbf-071b-4835-9563-529681a60a99';
-- UPDATE module_versions SET published_by = NULL WHERE published_by = 'd3d54cbf-071b-4835-9563-529681a60a99';
-- UPDATE provider_versions SET published_by = NULL WHERE published_by = 'd3d54cbf-071b-4835-9563-529681a60a99';
-- UPDATE api_keys SET user_id = NULL WHERE user_id = 'd3d54cbf-071b-4835-9563-529681a60a99';
