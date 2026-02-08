-- Rollback: Revert role template scopes to previous state

-- Revert Publisher role
UPDATE role_templates
SET scopes = '["modules:read", "modules:write", "providers:read", "providers:write", "scm:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'publisher';

-- Revert DevOps role
UPDATE role_templates
SET scopes = '["modules:read", "providers:read", "providers:write", "mirrors:read", "mirrors:manage", "scm:read", "scm:manage"]'::jsonb,
    updated_at = NOW()
WHERE name = 'devops';

-- Remove Auditor role
DELETE FROM role_templates WHERE name = 'auditor';
