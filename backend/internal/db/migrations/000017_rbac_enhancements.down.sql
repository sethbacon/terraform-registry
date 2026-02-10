-- Migration 017: RBAC Enhancements (Rollback)

-- Remove columns from mirror_configurations
ALTER TABLE mirror_configurations DROP COLUMN IF EXISTS requires_approval;
ALTER TABLE mirror_configurations DROP COLUMN IF EXISTS approval_status;

-- Remove role_template_id from api_keys
DROP INDEX IF EXISTS idx_api_keys_role_template;
ALTER TABLE api_keys DROP COLUMN IF EXISTS role_template_id;

-- Drop mirror policies
DROP INDEX IF EXISTS idx_mirror_policies_priority;
DROP INDEX IF EXISTS idx_mirror_policies_active;
DROP INDEX IF EXISTS idx_mirror_policies_type;
DROP INDEX IF EXISTS idx_mirror_policies_org;
DROP TABLE IF EXISTS mirror_policies;

-- Drop mirror approval requests
DROP INDEX IF EXISTS idx_mirror_approval_requests_provider;
DROP INDEX IF EXISTS idx_mirror_approval_requests_mirror;
DROP INDEX IF EXISTS idx_mirror_approval_requests_org;
DROP INDEX IF EXISTS idx_mirror_approval_requests_status;
DROP TABLE IF EXISTS mirror_approval_requests;

-- Drop role templates
DROP INDEX IF EXISTS idx_role_templates_name;
DROP TABLE IF EXISTS role_templates;
