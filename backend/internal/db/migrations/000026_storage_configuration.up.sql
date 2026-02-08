-- Migration 026: Storage Configuration
-- Stores storage backend configuration in the database for first-run setup
-- and allows configuration via the admin UI.

-- System settings table for global configuration (singleton pattern)
CREATE TABLE IF NOT EXISTS system_settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1), -- Ensures only one row
    storage_configured BOOLEAN NOT NULL DEFAULT false,
    storage_configured_at TIMESTAMP,
    storage_configured_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default row if not exists
INSERT INTO system_settings (id, storage_configured)
VALUES (1, false)
ON CONFLICT (id) DO NOTHING;

-- Storage configuration table
-- Secrets are encrypted before storage
CREATE TABLE IF NOT EXISTS storage_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Backend type: local, azure, s3, gcs
    backend_type VARCHAR(20) NOT NULL,

    -- Common settings
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Local storage settings
    local_base_path VARCHAR(1024),
    local_serve_directly BOOLEAN DEFAULT true,

    -- Azure Blob Storage settings
    azure_account_name VARCHAR(255),
    azure_account_key_encrypted TEXT, -- Encrypted
    azure_container_name VARCHAR(255),
    azure_cdn_url VARCHAR(1024),

    -- S3 settings
    s3_endpoint VARCHAR(1024),
    s3_region VARCHAR(100),
    s3_bucket VARCHAR(255),
    s3_auth_method VARCHAR(50), -- default, static, oidc, assume_role
    s3_access_key_id_encrypted TEXT, -- Encrypted
    s3_secret_access_key_encrypted TEXT, -- Encrypted
    s3_role_arn VARCHAR(255),
    s3_role_session_name VARCHAR(100),
    s3_external_id VARCHAR(255),
    s3_web_identity_token_file VARCHAR(1024),

    -- GCS settings
    gcs_bucket VARCHAR(255),
    gcs_project_id VARCHAR(255),
    gcs_auth_method VARCHAR(50), -- default, service_account, workload_identity
    gcs_credentials_file VARCHAR(1024),
    gcs_credentials_json_encrypted TEXT, -- Encrypted
    gcs_endpoint VARCHAR(1024),

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    -- Constraints
    CONSTRAINT valid_backend_type CHECK (backend_type IN ('local', 'azure', 's3', 'gcs'))
);

-- Create index for active config lookup
CREATE INDEX IF NOT EXISTS idx_storage_config_active ON storage_config(is_active) WHERE is_active = true;

-- Add comment
COMMENT ON TABLE storage_config IS 'Stores storage backend configuration. Sensitive fields are encrypted.';
COMMENT ON TABLE system_settings IS 'Global system settings (singleton). Controls first-run setup state.';
