-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Organizations table for multi-tenancy support
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    oidc_sub VARCHAR(255) UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- API keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Organization membership and roles
CREATE TABLE organization_members (
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

-- Modules table
CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    namespace VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    system VARCHAR(255) NOT NULL,
    description TEXT,
    source VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, namespace, name, system)
);

-- Module versions table
CREATE TABLE module_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    module_id UUID REFERENCES modules(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    storage_path VARCHAR(1024) NOT NULL,
    storage_backend VARCHAR(50) NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    published_by UUID REFERENCES users(id),
    download_count BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (module_id, version)
);

-- Providers table
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    namespace VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    description TEXT,
    source VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, namespace, type)
);

-- Provider versions table
CREATE TABLE provider_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID REFERENCES providers(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    protocols JSONB NOT NULL DEFAULT '[]'::jsonb,
    gpg_public_key TEXT NOT NULL,
    shasums_url VARCHAR(1024) NOT NULL,
    shasums_signature_url VARCHAR(1024) NOT NULL,
    published_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, version)
);

-- Provider platforms table
CREATE TABLE provider_platforms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_version_id UUID REFERENCES provider_versions(id) ON DELETE CASCADE,
    os VARCHAR(50) NOT NULL,
    arch VARCHAR(50) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    storage_path VARCHAR(1024) NOT NULL,
    storage_backend VARCHAR(50) NOT NULL,
    size_bytes BIGINT NOT NULL,
    shasum VARCHAR(64) NOT NULL,
    download_count BIGINT DEFAULT 0,
    UNIQUE (provider_version_id, os, arch)
);

-- Download events table for analytics
CREATE TABLE download_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    version_id UUID NOT NULL,
    user_id UUID REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Audit logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    organization_id UUID REFERENCES organizations(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB,
    ip_address INET,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oidc_sub ON users(oidc_sub);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_modules_org ON modules(organization_id);
CREATE INDEX idx_modules_namespace ON modules(namespace, name, system);
CREATE INDEX idx_module_versions_module ON module_versions(module_id);
CREATE INDEX idx_module_versions_version ON module_versions(version);
CREATE INDEX idx_providers_org ON providers(organization_id);
CREATE INDEX idx_providers_namespace ON providers(namespace, type);
CREATE INDEX idx_provider_versions_provider ON provider_versions(provider_id);
CREATE INDEX idx_provider_versions_version ON provider_versions(version);
CREATE INDEX idx_provider_platforms_version ON provider_platforms(provider_version_id);
CREATE INDEX idx_provider_platforms_os_arch ON provider_platforms(os, arch);
CREATE INDEX idx_download_events_resource ON download_events(resource_type, resource_id);
CREATE INDEX idx_download_events_created ON download_events(created_at);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_organization ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- Insert default organization for single-tenant mode
INSERT INTO organizations (name, display_name) VALUES ('default', 'Default Organization');
