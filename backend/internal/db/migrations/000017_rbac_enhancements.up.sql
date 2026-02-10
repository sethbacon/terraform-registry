-- Migration 017: RBAC Enhancements
-- Adds role templates, mirror approval workflows, and mirror policies

-- ============================================================================
-- 1. Role Templates
-- Predefined scope bundles for common use cases
-- ============================================================================

CREATE TABLE role_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    scopes JSONB NOT NULL DEFAULT '[]',
    is_system BOOLEAN DEFAULT false,  -- System templates cannot be deleted
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default system role templates
INSERT INTO role_templates (name, display_name, description, scopes, is_system) VALUES
('viewer', 'Viewer', 'Read-only access to modules, providers, and mirror status',
 '["modules:read", "providers:read", "mirrors:read"]'::jsonb, true),
('publisher', 'Publisher', 'Can upload and manage modules and providers',
 '["modules:read", "modules:write", "providers:read", "providers:write"]'::jsonb, true),
('mirror_manager', 'Mirror Manager', 'Can configure and manage provider mirroring',
 '["modules:read", "providers:read", "providers:write", "mirrors:read", "mirrors:manage"]'::jsonb, true),
('admin', 'Administrator', 'Full access to all registry features',
 '["admin"]'::jsonb, true);

CREATE INDEX idx_role_templates_name ON role_templates(name);

-- ============================================================================
-- 2. Mirror Approval Workflows
-- Track approval requests for mirroring specific providers
-- ============================================================================

CREATE TABLE mirror_approval_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mirror_config_id UUID REFERENCES mirror_configurations(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- What is being requested
    provider_namespace VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255),  -- NULL means entire namespace

    -- Request details
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),

    -- Approval details
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    review_notes TEXT,

    -- Auto-approval settings
    auto_approved BOOLEAN DEFAULT false,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP  -- Optional expiration for time-limited approvals
);

CREATE INDEX idx_mirror_approval_requests_status ON mirror_approval_requests(status);
CREATE INDEX idx_mirror_approval_requests_org ON mirror_approval_requests(organization_id);
CREATE INDEX idx_mirror_approval_requests_mirror ON mirror_approval_requests(mirror_config_id);
CREATE INDEX idx_mirror_approval_requests_provider ON mirror_approval_requests(provider_namespace, provider_name);

-- ============================================================================
-- 3. Mirror Policies
-- Define allowed/denied upstream registries and namespaces
-- ============================================================================

CREATE TABLE mirror_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Policy type: 'allow' or 'deny'
    policy_type VARCHAR(10) NOT NULL CHECK (policy_type IN ('allow', 'deny')),

    -- What this policy applies to
    upstream_registry VARCHAR(512),  -- NULL means all registries
    namespace_pattern VARCHAR(255),  -- Supports wildcards like 'hashicorp/*' or '*'
    provider_pattern VARCHAR(255),   -- Supports wildcards like 'aws' or '*'

    -- Policy settings
    priority INT DEFAULT 0,  -- Higher priority policies are evaluated first
    is_active BOOLEAN DEFAULT true,
    requires_approval BOOLEAN DEFAULT false,  -- If true, matching mirrors need approval

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,

    UNIQUE (organization_id, name)
);

CREATE INDEX idx_mirror_policies_org ON mirror_policies(organization_id);
CREATE INDEX idx_mirror_policies_type ON mirror_policies(policy_type);
CREATE INDEX idx_mirror_policies_active ON mirror_policies(is_active) WHERE is_active = true;
CREATE INDEX idx_mirror_policies_priority ON mirror_policies(priority DESC);

-- Insert default global policies (allow all from registry.terraform.io)
-- These apply when no organization-specific policies exist
INSERT INTO mirror_policies (organization_id, name, description, policy_type, upstream_registry, namespace_pattern, provider_pattern, priority, is_active, requires_approval)
VALUES
(NULL, 'default-allow-hashicorp', 'Allow mirroring HashiCorp official providers', 'allow', 'https://registry.terraform.io', 'hashicorp', '*', 100, true, false),
(NULL, 'default-require-approval-other', 'Require approval for non-HashiCorp providers', 'allow', 'https://registry.terraform.io', '*', '*', 0, true, true);

-- ============================================================================
-- 4. Add role_template_id to API keys for easier role assignment
-- ============================================================================

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS role_template_id UUID REFERENCES role_templates(id) ON DELETE SET NULL;

CREATE INDEX idx_api_keys_role_template ON api_keys(role_template_id);

-- ============================================================================
-- 5. Update mirror_configurations to track approval requirements
-- ============================================================================

ALTER TABLE mirror_configurations ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN DEFAULT false;
ALTER TABLE mirror_configurations ADD COLUMN IF NOT EXISTS approval_status VARCHAR(50) DEFAULT 'not_required' CHECK (approval_status IN ('not_required', 'pending', 'approved', 'rejected'));
