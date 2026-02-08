-- Migration 020: Update Role Templates
-- - Add new scopes: organizations:read, organizations:write, scm:read, scm:manage
-- - Rename mirror_manager to devops with updated scopes
-- - Update viewer and publisher roles with new scopes

-- Update viewer role to include new read scopes
UPDATE role_templates
SET
    description = 'Read-only access to modules, providers, mirrors, organizations, and SCM configurations',
    scopes = '["modules:read", "providers:read", "mirrors:read", "organizations:read", "scm:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'viewer';

-- Update publisher role to include scm:read
UPDATE role_templates
SET
    scopes = '["modules:read", "modules:write", "providers:read", "providers:write", "scm:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'publisher';

-- Rename mirror_manager to devops and update scopes
UPDATE role_templates
SET
    name = 'devops',
    display_name = 'DevOps',
    description = 'Can manage SCM integrations and provider mirroring for CI/CD pipelines',
    scopes = '["modules:read", "providers:read", "providers:write", "mirrors:read", "mirrors:manage", "scm:read", "scm:manage"]'::jsonb,
    updated_at = NOW()
WHERE name = 'mirror_manager';
