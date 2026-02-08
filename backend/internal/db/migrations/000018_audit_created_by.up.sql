-- Add created_by columns to modules and providers for audit tracking

-- Add created_by to modules
ALTER TABLE modules ADD COLUMN created_by UUID REFERENCES users(id);

-- Add created_by to providers
ALTER TABLE providers ADD COLUMN created_by UUID REFERENCES users(id);

-- Add indexes for the new columns
CREATE INDEX idx_modules_created_by ON modules(created_by);
CREATE INDEX idx_providers_created_by ON providers(created_by);

-- Add index on api_keys user_id if not exists (for audit queries)
-- This index already exists from initial schema but adding for completeness
