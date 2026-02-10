-- Add description column to api_keys table
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS description TEXT;
