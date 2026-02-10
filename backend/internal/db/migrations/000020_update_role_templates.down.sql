-- Migration 020 Down: Revert Role Template Updates

-- Revert viewer role
UPDATE role_templates
SET
    description = 'Read-only access to modules, providers, and mirror status',
    scopes = '["modules:read", "providers:read", "mirrors:read"]'::jsonb,
    updated_at = NOW()
WHERE name = 'viewer';

-- Revert publisher role
UPDATE role_templates
SET
    scopes = '["modules:read", "modules:write", "providers:read", "providers:write"]'::jsonb,
    updated_at = NOW()
WHERE name = 'publisher';

-- Revert devops back to mirror_manager
UPDATE role_templates
SET
    name = 'mirror_manager',
    display_name = 'Mirror Manager',
    description = 'Can configure and manage provider mirroring',
    scopes = '["modules:read", "providers:read", "providers:write", "mirrors:read", "mirrors:manage"]'::jsonb,
    updated_at = NOW()
WHERE name = 'devops';
